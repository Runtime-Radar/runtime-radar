//go:build !tinygo.wasm

package main

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	runtime_event "github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/client"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/config"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/consumer"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/database"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor/detector"
	kube_manager "github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	notifier_api "github.com/runtime-radar/runtime-radar/notifier/api"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Local listen address for the in-process gRPC server. Same value as the
	// production default; chosen explicitly here so subtests don't depend on
	// production config defaults.
	listenGRPCAddr = "127.0.0.1:8000"

	// Smoke-test deadline: how long the smoke test waits for a RuntimeEvent
	// to surface on history_events after publishing an empty event to
	// runtime_events. The first event the worker processes also triggers
	// wazero to compile every seeded WASM detector — that's a ~30-90s
	// one-shot tax depending on the test host. After this first event
	// completes, per-fixture sync in TestDetectorPipeline uses the much
	// smaller perFixtureDeadline.
	historyDrainDeadline = 3 * time.Minute
)

// Package-level state shared with subtests.
//
// The downstream-client mocks are exposed as package-level vars so subtests
// (TestDetectorPipeline, etc.) can configure return values per-fixture and
// inspect captured call args. Access is single-threaded by virtue of subtests
// running sequentially within TestDetectorPipeline; the mutex is used by the
// smoke test only as defensive belt-and-braces.
var (
	testCfg *config.Config

	enforcerMock      *client.EnforcerClientMock
	notifierMock      *client.NotifierClientMock
	podControllerMock *client.PodControllerClientMock
	// mocksMu serializes mock reconfiguration done by per-fixture subtests.
	// Subtests run sequentially today, but the smoke test publishes through
	// the same shared mocks; the lock keeps mock reconfiguration honest if
	// future work introduces concurrency.
	mocksMu sync.Mutex

	// Channel of history-queue deliveries surfaced by the in-test consumer
	// in TestMain. Subtests pull from here via drainHistoryEvent (helpers_test.go).
	historyDeliveries chan amqp.Delivery

	// runtimePublisher is the lib/rabbit MessageBroker bound to the
	// runtime_events queue, used by tests to inject Tetragon events into
	// the production consumer.
	runtimePublisher *rabbit.MessageBroker

	// historyConsumeChannel is the raw amqp.Channel used by the in-test
	// history-queue consumer. Exposed so subtests can call purgeQueue on
	// runtime_events / history_events between iterations.
	historyConsumeChannel *amqp.Channel
)

// TestMain bootstraps the full event-processor pipeline against real
// Postgres + RabbitMQ + WASM-detector seed, with downstream gRPC clients
// (policy-enforcer, notifier, kube-manager) substituted by minimock
// instances. The shape mirrors policy-enforcer/cmd/policy-enforcer/main_test.go.
//
// Bootstrap steps (in order):
//  1. Build cfg from env (config.New parses flags), init logger.
//  2. Open Postgres and run database.Migrate(db, true) to wipe + replay.
//  3. Open two lib/rabbit MessageBrokers — one publisher on runtime_events
//     (used by tests to inject events), one history broker passed to the
//     worker pool for publishing RuntimeEvents.
//  4. Open a separate amqp connection to consume from history_events and
//     funnel deliveries into a buffered Go channel.
//  5. Construct minimock client instances for all three downstream services
//     and stash them in package-level vars.
//  6. Override cfg.DeployDir to point at the repo's compiled WASM blobs
//     (../../deploy relative to cmd/event-processor) and call the
//     production ensureDetectors() to seed Postgres.
//  7. Build the worker pool via getPool() requesting poolSize=1; note that
//     processor.NewWorkersPool enforces minPoolSize=2 internally, so two
//     worker goroutines are actually started. Subtests assume the shared
//     job-channel ordering this implies (see runFixtureSubtest comments).
//  8. Force-set the pool config to HistoryControl=ALL so that every
//     processed event produces exactly one history-queue publish — that's
//     the deterministic sync point the smoke test and TestDetectorPipeline
//     rely on. (Default is HistoryControl=NONE which would skip the publish
//     entirely.)
//  9. Start the production consumer (consumer.Run) on a stop channel.
//
// 10. Start the gRPC server on 127.0.0.1:8000 wired via composeServices().
// 11. m.Run().
// 12. Graceful teardown: stop server, close consumer, drain rabbit, close DB.
func TestMain(m *testing.M) {
	testCfg = config.New()

	if testing.Verbose() {
		logger.Init("", "DEBUG")
	} else {
		logger.Init("", "INFO")
	}

	// Default DeployDir is "deploy" relative to cwd. Tests run from
	// cmd/event-processor/, so the compiled blobs live at ../../deploy.
	// Only override if the user hasn't already pointed somewhere explicit
	// (e.g. via DEPLOY_DIR env var or -deployDir flag).
	if testCfg.DeployDir == "deploy" {
		testCfg.DeployDir = "../../deploy"
	}

	// Postgres
	db, closeDB, err := database.New(
		testCfg.PostgresAddr,
		testCfg.PostgresDB+"_test",
		testCfg.PostgresUser,
		testCfg.PostgresPassword,
		testCfg.PostgresSSLMode,
		testCfg.PostgresSSLCheckCert,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to open DB: %v", err)
	}
	if err := database.Migrate(db, true); err != nil {
		log.Fatal().Msgf("### Failed to migrate DB: %v", err)
	}

	// RabbitMQ — runtime_events publisher (tests publish events here; the
	// production Consumer below pulls them off).
	runtimePublisher, err = rabbit.NewMessageBroker(
		testCfg.RabbitAddr,
		testCfg.RabbitUser,
		testCfg.RabbitPassword,
		testCfg.RabbitRuntimeEventsQueue,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize runtime publisher: %v", err)
	}

	// runtime_events consumer broker — what the production Consumer reads
	// from. Same wiring as production main(): WithConsumer + a prefetch
	// count.
	runtimeConsumerMB, err := rabbit.NewMessageBroker(
		testCfg.RabbitAddr,
		testCfg.RabbitUser,
		testCfg.RabbitPassword,
		testCfg.RabbitRuntimeEventsQueue,
		rabbit.WithConsumer("event-processor-test", testCfg.RabbitRuntimeEventsQueuePrefetchCount),
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize runtime consumer broker: %v", err)
	}

	// history broker — the worker pool publishes RuntimeEvents here; the
	// in-test consumer below reads them.
	historyMB, err := rabbit.NewMessageBroker(
		testCfg.RabbitAddr,
		testCfg.RabbitUser,
		testCfg.RabbitPassword,
		testCfg.RabbitHistoryEventsQueue,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize history broker: %v", err)
	}

	// In-test history-queue consumer. We open a separate amqp connection
	// here (not via lib/rabbit, which assumes proto-binary Consume) so we
	// have raw amqp.Delivery handles to feed into helpers_test.go's
	// drainHistoryEvent.
	historyConn, historyChan, historyDeliveriesIn, err := openHistoryConsumer(testCfg)
	if err != nil {
		log.Fatal().Msgf("### Failed to open history consumer: %v", err)
	}
	historyConsumeChannel = historyChan
	historyDeliveries = make(chan amqp.Delivery, 64)

	// Funnel: forward deliveries into the buffered package channel so a
	// slow subtest can't block the amqp consumer goroutine.
	go func() {
		for d := range historyDeliveriesIn {
			select {
			case historyDeliveries <- d:
			default:
				// Channel full: drop silently. Tests that care
				// will purge before publishing; the smoke test
				// exercises the happy path with one outstanding
				// delivery.
				log.Warn().Msgf("history delivery dropped — buffered channel full")
			}
		}
	}()

	// Mocks. We pass a noopTester (rather than *testing.T from m.Run's
	// per-test instances) because the mocks live for the whole TestMain
	// process — a per-test Tester would attach a Cleanup hook to the wrong
	// scope. Per-fixture subtests reset call recorders via t.Cleanup;
	// nothing here needs minimock's Fatal/Errorf wiring.
	tester := &noopTester{}
	enforcerMock = client.NewEnforcerClientMock(tester)
	notifierMock = client.NewNotifierClientMock(tester)
	podControllerMock = client.NewPodControllerClientMock(tester)

	// Detector plugin (wazero host). Same construction path as production
	// main() so the seed-and-load codepath is byte-identical to prod.
	plugin, err := detector.NewPlugin(context.Background())
	if err != nil {
		log.Fatal().Msgf("### Failed to instantiate detector plugin: %v", err)
	}

	// Seed Postgres with the compiled WASM blobs. Production seed path.
	if err := ensureDetectors(db, plugin, testCfg.DeployDir); err != nil {
		log.Fatal().Msgf("### ensureDetectors failed: %v", err)
	}
	// Sanity: confirm we actually inserted some detectors. This catches a
	// stale or empty deploy/ directory before any subtest runs.
	repo := &database.DetectorDatabase{DB: db}
	if count, err := repo.GetCount(context.Background(), nil); err != nil {
		log.Fatal().Msgf("### Failed to count detectors: %v", err)
	} else if count == 0 {
		log.Fatal().Msgf("### ensureDetectors produced 0 rows; deploy dir %q is empty", testCfg.DeployDir)
	}

	// Worker pool with poolSize=1 requested; processor.NewWorkersPool
	// enforces minPoolSize=2, so two workers actually run. That's fine for
	// these tests because the per-fixture purgeQueue + drainPending barriers
	// already isolate one fixture from the next, and HistoryControl=ALL
	// guarantees one history-queue publish per processed event.
	//
	// Inlined from getPool() in main.go (load bins, load config, build pool)
	// so we can pass processor.WithReports(). Reports are consumed by the
	// pre-warm loop below to detect when each worker has finished its initial
	// getMatcher() call — see comment on the pre-warm block.
	configRepo := &database.ConfigDatabase{DB: db}
	bins, err := repo.GetAllBins(context.Background(), nil)
	if err != nil {
		log.Fatal().Msgf("### Failed to get detector bins: %v", err)
	}
	poolCfg, err := configRepo.GetLast(context.Background(), true)
	if err != nil {
		log.Fatal().Msgf("### Failed to get last config: %v", err)
	}
	pool, err := processor.NewWorkersPool(
		1, /* poolSize (clamped to minPoolSize=2 by NewWorkersPool) */
		testCfg.JobsBufferSize,
		historyMB,
		plugin,
		enforcerMock,
		notifierMock,
		podControllerMock,
		bins,
		poolCfg,
		processor.WithReports(),
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to initialize workers pool: %v", err)
	}

	// Force HistoryControl=ALL so that every processed event publishes a
	// history-queue message regardless of detection outcome. That is the
	// sync point used by the smoke test and the per-detector matrix.
	cfg := pool.Config()
	cfg.Config.HistoryControl = runtime_event.Config_ConfigJSON_ALL
	pool.SetConfig(cfg)

	// Pre-warm: each worker's first getMatcher() call compiles every WASM
	// blob via a fresh wazero.Runtime, which on slow CI runners can take
	// several minutes per worker (observed: 4m22s and 4m15s on a GitLab
	// runner). Without gating, that compile cost is paid against the smoke
	// test's 3-minute historyDrainDeadline and the per-fixture 30s deadline
	// in TestDetectorPipeline, producing flaky timeouts. Drip-publish empty
	// jobs into pool.Jobs() and observe pool.Reports() — a Report from
	// worker[id] proves that worker is past getMatcher and inside its select
	// loop. Tickered publishing prevents the faster worker from claiming
	// every event before the slower one finishes compiling.
	const expectedWorkers = 2 // matches processor.minPoolSize after clamp
	const prewarmTimeout = 10 * time.Minute

	log.Info().Int("expected_workers", expectedWorkers).Msg("pre-warm: waiting for matcher initialization")
	prewarmStart := time.Now()
	seen := make(map[int]struct{}, expectedWorkers)
	prewarmTicker := time.NewTicker(500 * time.Millisecond)
	prewarmDeadline := time.After(prewarmTimeout)
	emptyEvent := &tetragon.GetEventsResponse{}

	for len(seen) < expectedWorkers {
		select {
		case <-prewarmDeadline:
			log.Fatal().Msgf("### matcher pre-warm timed out after %v; saw workers=%v", prewarmTimeout, seen)
		case r := <-pool.Reports():
			if _, dup := seen[r.ID]; !dup {
				log.Info().Int("worker_id", r.ID).Dur("elapsed", time.Since(prewarmStart)).Msg("pre-warm: worker matcher ready")
			}
			seen[r.ID] = struct{}{}
		case <-prewarmTicker.C:
			select {
			case pool.Jobs() <- emptyEvent:
			default:
				// jobs buffer full — workers will drain. Skip this tick.
			}
		}
	}
	prewarmTicker.Stop()
	log.Info().Dur("total", time.Since(prewarmStart)).Msg("pre-warm: all matchers initialized")

	// Drain Reports() for the rest of the test process. Without this, each
	// worker's report-send blocks for the 10ms reportTimeout in
	// processor.worker.go before the report is dropped, accumulating against
	// per-fixture deadlines.
	go func() {
		//nolint:revive,empty-block
		for range pool.Reports() {
		}
	}()

	// Start the production consumer (reads runtime_events, feeds the
	// worker pool). Stop channel is closed in teardown.
	consumerStop := make(chan struct{})
	go (&consumer.Consumer{
		PublishConsumer: runtimeConsumerMB,
		Processor:       pool,
	}).Run(consumerStop)

	// gRPC server, same wiring as production.
	lis, err := net.Listen("tcp", listenGRPCAddr)
	if err != nil {
		log.Fatal().Msgf("### Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptor.Recovery, interceptor.Correlation),
	}
	grpcSrv := grpc.NewServer(opts...)

	// Auth disabled in tests; pass nil verifier and false. composeServices
	// in main.go handles this branch.
	configSvc, detectorSvc := composeServices(db, pool, plugin, nil, false)
	runtime_event.RegisterConfigControllerServer(grpcSrv, configSvc)
	runtime_event.RegisterDetectorControllerServer(grpcSrv, detectorSvc)
	reflection.Register(grpcSrv)

	go func() {
		// Serve returns nil on GracefulStop and grpc.ErrServerStopped on
		// Stop. We don't want a non-fatal Serve error from a racing
		// teardown to override m.Run's exit code via log.Fatal — log it
		// at error level instead.
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error().Msgf("gRPC Serve returned: %v", err)
		}
	}()

	// Run tests.
	res := m.Run()

	// Teardown. Order is important:
	//   1. Stop the production Consumer goroutine first so no new events
	//      are pushed into the worker pool's jobs channel.
	//   2. Close the worker pool — this closes wp.fire and waits for all
	//      workers to drain in-flight jobs and exit.
	//   3. Stop the gRPC server.
	//   4. Close downstream rabbit connections / DB.
	// Closing the pool BEFORE the consumer would race: the consumer
	// goroutine would still be writing to wp.jobs after workers stopped
	// reading, leaking that goroutine.
	close(consumerStop)
	pool.Close()
	grpcSrv.GracefulStop()
	_ = runtimePublisher.Close()
	_ = runtimeConsumerMB.Close()
	_ = historyMB.Close()
	_ = historyChan.Close()
	_ = historyConn.Close()
	_ = closeDB()

	// Manually invoke the package-level mocks' MinimockFinish hooks now
	// that no more pipeline activity will touch them. The mocks were
	// constructed with a noopTester whose Cleanup intentionally drops the
	// hook (Cleanup on a TestMain-scoped mock has nowhere sensible to
	// attach), so we run finishers explicitly here. Subtests in
	// TestDetectorPipeline configure mocks via Set() rather than
	// expectations, so under normal operation these calls have nothing to
	// verify; they exist to surface a future When/Return-based fixture
	// regression where invocation counts ARE checked.
	enforcerMock.MinimockFinish()
	notifierMock.MinimockFinish()
	podControllerMock.MinimockFinish()

	os.Exit(res)
}

// openHistoryConsumer opens a dedicated amqp connection + channel and starts
// a consumer on the history_events queue. It returns the connection (so the
// caller can close it on teardown), the channel (so the caller can purge the
// queue between subtests), and the raw delivery channel.
//
// We bypass lib/rabbit here because lib/rabbit.Consume is proto-aware and
// auto-acks on receive; tests want raw amqp.Delivery values so they can use
// existing helpers (helpers_test.go).
func openHistoryConsumer(cfg *config.Config) (*amqp.Connection, *amqp.Channel, <-chan amqp.Delivery, error) {
	url := "amqp://" + cfg.RabbitUser + ":" + cfg.RabbitPassword + "@" + cfg.RabbitAddr
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, nil, err
	}

	// Declare the queue with the same parameters the production
	// MessageBroker uses (durable=true, autoDelete=false, exclusive=false,
	// noWait=false). Idempotent if the queue already exists with matching
	// settings.
	if _, err := ch.QueueDeclare(cfg.RabbitHistoryEventsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, nil, err
	}

	// autoAck=true to keep the test code simple. Tests in this package do
	// not need at-least-once redelivery semantics.
	deliveries, err := ch.Consume(
		cfg.RabbitHistoryEventsQueue,
		"event-processor-test-history",
		true,  /* autoAck */
		false, /* exclusive */
		false, /* noLocal */
		false, /* noWait */
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, nil, err
	}

	return conn, ch, deliveries, nil
}

// TestServerBootsAndConsumesNoOpEvent is the smoke test that proves the full
// pipeline (TestMain bootstrap → publish → Consumer → WorkersPool → history)
// is alive before the per-detector matrix runs.
//
// We publish an empty *tetragon.GetEventsResponse (no events embedded) to
// runtime_events. The production Consumer reads it; the worker pool's
// detector chain runs (no detectors fire because there's no body); a single
// RuntimeEvent is still emitted because we pinned HistoryControl=ALL in
// TestMain. The test asserts that the RuntimeEvent surfaces on
// history_events within historyDrainDeadline.
//
// If this test fails, something in the bootstrap is broken — wrong queue
// names, mocks not wired, history publish path skipped, etc.
func TestServerBootsAndConsumesNoOpEvent(t *testing.T) {
	// Purge both queues before draining buffered deliveries: stale messages
	// from a prior crashed run could otherwise be drained and mistaken for
	// the smoke test's own publish. drainPending after purgeQueue ensures
	// any in-flight delivery the funnel goroutine forwarded to the buffered
	// channel is also discarded.
	purgeQueue(t, historyConsumeChannel, testCfg.RabbitRuntimeEventsQueue)
	purgeQueue(t, historyConsumeChannel, testCfg.RabbitHistoryEventsQueue)
	drainPending()

	// Publish an empty event.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := runtimePublisher.Publish(ctx, &tetragon.GetEventsResponse{}); err != nil {
		t.Fatalf("publish empty event: %v", err)
	}

	// Block on history_events with the configured deadline. drainHistoryEvent
	// (helpers_test.go) handles unmarshalling and timeout reporting.
	ev := drainHistoryEvent(t, context.Background(), historyDeliveries, historyDrainDeadline)
	if ev == nil {
		// drainHistoryEvent already called t.Fatalf, but the early
		// return keeps the type checker happy.
		return
	}

	// Sanity: the RuntimeEvent should have an Id and a TetragonVersion set
	// (the worker stamps these unconditionally).
	if ev.GetId() == "" {
		t.Errorf("RuntimeEvent.Id empty")
	}
	if ev.GetTetragonVersion() == "" {
		t.Errorf("RuntimeEvent.TetragonVersion empty")
	}

	// Confirm the empty path: no threats, not an incident.
	if got := len(ev.GetThreats()); got != 0 {
		t.Errorf("RuntimeEvent.Threats: got %d, want 0", got)
	}
	if ev.GetIsIncident() {
		t.Errorf("RuntimeEvent.IsIncident: got true, want false")
	}

	// Sanity: the embedded tetragon event round-trips. This catches a
	// future regression where the Consumer or worker drops the original
	// event payload.
	if ev.GetEvent() == nil {
		t.Errorf("RuntimeEvent.Event nil; want round-tripped tetragon.GetEventsResponse")
	} else {
		// Make sure proto.Marshal works on it (i.e. it's a valid
		// proto), even if all fields are zero-valued.
		if _, err := proto.Marshal(ev.GetEvent()); err != nil {
			t.Errorf("marshal embedded tetragon event: %v", err)
		}
	}
}

// drainPending non-blockingly removes any history deliveries already
// buffered at the time of the call. Called at the start of subtests that
// need a clean baseline. It does NOT wait for new deliveries.
func drainPending() {
	for {
		select {
		case <-historyDeliveries:
		default:
			return
		}
	}
}

// noopTester satisfies minimock.Tester for the long-lived package-level
// mock instances constructed in TestMain. Per-subtest assertions read
// recorded call args directly from the mock structs; minimock's own
// pass/fail wiring (which would need a *testing.T) is not used.
//
// Calls that minimock-generated code routes through Tester are funneled
// into the package logger so that genuine misuse (e.g. an unconfigured mock
// being invoked) shows up in test output without aborting TestMain.
type noopTester struct{}

func (n *noopTester) Fatal(args ...interface{}) {
	log.Error().Msgf("noopTester.Fatal: %v", args)
}
func (n *noopTester) Fatalf(format string, args ...interface{}) {
	log.Error().Msgf("noopTester.Fatalf: "+format, args...)
}
func (n *noopTester) Error(args ...interface{}) {
	log.Error().Msgf("noopTester.Error: %v", args)
}
func (n *noopTester) Errorf(format string, args ...interface{}) {
	log.Error().Msgf("noopTester.Errorf: "+format, args...)
}
func (n *noopTester) FailNow()         {}
func (n *noopTester) Cleanup(_ func()) {}
func (n *noopTester) Helper()          {}

// TestDetectorPipeline is the per-detector regression matrix. It discovers
// every fixture under event-processor/detector/wasm/<ID>/testdata/, publishes
// its tetragon event through the production pipeline, and asserts that the
// captured downstream-mock invocations and history-queue payload match the
// fixture's expectations.
//
// Subtests are nested t.Run(detectorID, ...) → t.Run(filename, ...) so a
// failing fixture surfaces with a fully-qualified name like
//
//	TestDetectorPipeline/PTCS_RT_REVERSE_SHELL_CREATE/positive_socket_dup2.json
//
// The matrix is run sequentially (no t.Parallel) because all subtests share
// the package-level mock instances and the single history-queue Go channel.
//
// Per-fixture lifecycle:
//  1. Reset all three mocks via fresh Set closures that capture invocations
//     into per-subtest slices.
//  2. Purge runtime_events / history_events queues and drain the buffered
//     historyDeliveries channel of any leftover payloads.
//  3. Publish the fixture's wire bytes to runtime_events.
//  4. Block on historyDeliveries (drainHistoryEvent) for the produced
//     RuntimeEvent.
//  5. Run the four assertion helpers from helpers_test.go against the
//     captured calls and the drained event.
func TestDetectorPipeline(t *testing.T) {
	fixtures, err := discoverFixtures()
	if err != nil {
		t.Fatalf("discoverFixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("discoverFixtures: no fixtures found under detector/wasm/*/testdata")
	}

	// Group by detector id (parent directory name). discoverFixtures returns
	// fixtures already sorted lexicographically, so per-group ordering is
	// deterministic without an additional sort.Strings on each slice — only
	// detectorIDs needs sorting because Go map iteration order is random.
	groups := make(map[string][]string, 32)
	for _, p := range fixtures {
		id := filepath.Base(filepath.Dir(filepath.Dir(p)))
		groups[id] = append(groups[id], p)
	}
	detectorIDs := make([]string, 0, len(groups))
	for id := range groups {
		detectorIDs = append(detectorIDs, id)
	}
	sort.Strings(detectorIDs)

	// Aggregate counters logged at the end of the run. Reported even if a
	// subtest fails (t.Fatalf in a subtest does not unwind the parent
	// TestDetectorPipeline goroutine). There is no skip path today (every
	// fixture under detector/wasm/<ID>/testdata is run), so no totalSkipped
	// counter is wired up.
	var (
		totalRun    int
		totalPassed int
		totalFailed int
	)

	t.Cleanup(func() {
		// Final cleanup at the end of the matrix: leave the queues
		// empty for any subsequent test in the package.
		drainPending()
		if historyConsumeChannel != nil {
			_, _ = historyConsumeChannel.QueuePurge(testCfg.RabbitRuntimeEventsQueue, false)
			_, _ = historyConsumeChannel.QueuePurge(testCfg.RabbitHistoryEventsQueue, false)
		}
		t.Logf("TestDetectorPipeline summary: detectors=%d fixtures=%d run=%d passed=%d failed=%d",
			len(detectorIDs), len(fixtures), totalRun, totalPassed, totalFailed)
	})

	// Per-fixture deadline. WASM compilation cost is amortized by the
	// smoke test in TestServerBootsAndConsumesNoOpEvent; subsequent events
	// hit the wazero compilation cache and complete in well under a
	// second. 30s is a generous belt-and-braces value.
	const perFixtureDeadline = 30 * time.Second

	for _, detectorID := range detectorIDs {
		paths := groups[detectorID]

		t.Run(detectorID, func(t *testing.T) {
			t.Logf("starting detector_id: %s", detectorID)
			for _, path := range paths {
				name := filepath.Base(path)

				t.Run(name, func(t *testing.T) {
					totalRun++
					runFixtureSubtest(t, path, perFixtureDeadline)
					if t.Failed() {
						totalFailed++
					} else {
						totalPassed++
					}
				})
			}
		})
	}
}

// runFixtureSubtest executes one fixture against the running pipeline. All
// assertion failures are reported through the supplied *testing.T; callers
// should consult t.Failed() (after the subtest returns) to decide whether
// the fixture passed.
//
// Coverage gap: every on-disk fixture currently has policy_response: null,
// so buildPolicyResponse always returns nil and the doJob branches that
// call wp.podController.Kill / wp.notify are not exercised by this matrix.
// The KillPod/Notify mock recorders are wired regardless so that a future
// BLOCK or NOTIFY fixture can assert the expected dispatches without code
// changes here. Note also that NOTIFY rules built by buildPolicyResponse
// have an empty Targets slice — when a NOTIFY fixture is added, it must
// either populate Targets in the canned response or assert zero
// notifications (notifier.go:75 iterates over Targets, producing zero
// Notify calls otherwise).
//
// Concurrency note: NewWorkersPool clamps requested poolSize=1 up to
// minPoolSize=2, so two worker goroutines compete for jobs from the
// shared channel. Per-fixture isolation is achieved by purging both
// queues + draining the buffered deliveries channel before each publish,
// and by relying on HistoryControl=ALL to produce exactly one history
// publish per processed event (the deterministic sync barrier this test
// blocks on). Any stale message is therefore caught and discarded
// before runFixtureSubtest's own Publish call.
func runFixtureSubtest(t *testing.T, path string, deadline time.Duration) {
	t.Helper()

	fx := loadFixture(t, path)

	// Per-subtest call recorders. Closures registered with the mocks
	// append into these slices; the assertion helpers read them after
	// drainHistoryEvent unblocks.
	var (
		policyCalls []*enforcer_api.EvaluatePolicyRuntimeEventReq
		killCalls   []*kube_manager.KillPodReq
		notifyCalls []*notifier_api.NotifyReq
		callsMutex  sync.Mutex
	)

	// Build the canned policy-enforcer response from the fixture.
	policyResp := buildPolicyResponse(fx.PolicyResponse)

	// Wire mocks. Each Set replaces the previous funcXxx — minimock allows
	// repeated Set calls without producing a stale-expectation error
	// because Set does not populate defaultExpectation/expectations.
	mocksMu.Lock()
	enforcerMock.EvaluatePolicyRuntimeEventMock.Set(
		func(_ context.Context, in *enforcer_api.EvaluatePolicyRuntimeEventReq, _ ...grpc.CallOption) (*enforcer_api.EvaluatePolicyRuntimeEventReq, error) {
			callsMutex.Lock()
			policyCalls = append(policyCalls, proto.Clone(in).(*enforcer_api.EvaluatePolicyRuntimeEventReq))
			callsMutex.Unlock()
			// Build the response: clone the request and overlay
			// the per-event policies declared in the fixture.
			respClone := proto.Clone(in).(*enforcer_api.EvaluatePolicyRuntimeEventReq)
			if policyResp != nil {
				for _, ev := range respClone.GetResult().GetEvents() {
					ev.Policy = proto.Clone(policyResp).(*enforcer_api.Policy)
				}
			}
			return respClone, nil
		},
	)
	podControllerMock.KillMock.Set(
		func(_ context.Context, in *kube_manager.KillPodReq, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			callsMutex.Lock()
			killCalls = append(killCalls, proto.Clone(in).(*kube_manager.KillPodReq))
			callsMutex.Unlock()
			return &emptypb.Empty{}, nil
		},
	)
	notifierMock.NotifyMock.Set(
		func(_ context.Context, in *notifier_api.NotifyReq, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			callsMutex.Lock()
			notifyCalls = append(notifyCalls, proto.Clone(in).(*notifier_api.NotifyReq))
			callsMutex.Unlock()
			return &emptypb.Empty{}, nil
		},
	)
	mocksMu.Unlock()

	// Reset state before publishing: purge any buffered messages on the
	// runtime/history queues FIRST, then drain the buffered Go channel.
	// Doing it in this order closes a window where the funnel goroutine
	// could still forward an in-flight delivery between drainPending and
	// purgeQueue, leaving stale data in historyDeliveries.
	purgeQueue(t, historyConsumeChannel, testCfg.RabbitRuntimeEventsQueue)
	purgeQueue(t, historyConsumeChannel, testCfg.RabbitHistoryEventsQueue)
	drainPending()

	t.Cleanup(func() {
		// Best-effort post-fixture cleanup. The next subtest's setup
		// will also drain/purge, so this is mostly defensive — it
		// keeps the queues empty between fixtures even when the
		// driver itself stops iterating early (e.g. on t.Fatalf).
		purgeQueue(t, historyConsumeChannel, testCfg.RabbitRuntimeEventsQueue)
		purgeQueue(t, historyConsumeChannel, testCfg.RabbitHistoryEventsQueue)
		drainPending()
	})

	// Publish the fixture's wire bytes directly via the underlying amqp
	// channel. lib/rabbit's MessageBroker.Publish only accepts a
	// proto.Message and re-marshals; we already have the wire form on
	// disk, and round-tripping through the upstream tetragon proto here
	// would mask any encoding drift between protojson and proto.Marshal.
	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := publishWire(pubCtx, fx.EventWire); err != nil {
		// t.Fatalf calls runtime.Goexit; control never reaches the
		// rest of this function.
		t.Fatalf("publishWire: %v", err)
	}

	// Block on the history queue. drainHistoryEvent always calls t.Fatalf
	// before returning nil, so the nil check below is structurally
	// unreachable — but keeping the explicit guard makes the data flow
	// obvious and matches the smoke test's defensive style.
	ev := drainHistoryEvent(t, context.Background(), historyDeliveries, deadline)
	if ev == nil {
		return
	}

	// Snapshot captured calls under the mutex so the assertion helpers
	// observe a consistent slice. The history-queue publish that
	// drainHistoryEvent unblocks on happens BEFORE worker.go calls
	// wp.podController.Kill (worker.go:195) and wp.notify (worker.go:201),
	// so KillPod/Notify mock invocations may not yet be visible at this
	// point for fixtures that exercise BLOCK/NOTIFY rules. No fixture in
	// the on-disk corpus does today (policy_response is null everywhere);
	// when one is added, this function will need an additional sync
	// barrier (e.g. an explicit await on the killCalls/notifyCalls slice
	// length, or a different sync point in worker.go).
	callsMutex.Lock()
	pCalls := append([]*enforcer_api.EvaluatePolicyRuntimeEventReq(nil), policyCalls...)
	kCalls := append([]*kube_manager.KillPodReq(nil), killCalls...)
	nCalls := append([]*notifier_api.NotifyReq(nil), notifyCalls...)
	callsMutex.Unlock()

	assertPolicyCall(t, fx.Expect.PolicyCall, pCalls)
	assertReasonContains(t, fx.Expect.PolicyCall, ev)
	assertTactics(t, fx.Expect.PolicyCall, ev)
	assertKillPodCall(t, fx.Expect.KillPodCall, kCalls)
	assertNotifyCall(t, fx.Expect.NotifyCall, nCalls)
	assertHistoryEvent(t, fx.Expect.HistoryEvent, ev)
}

// discoverFixtures walks event-processor/detector/wasm/*/testdata/*.json
// relative to the test binary's working directory (cmd/event-processor/) and
// returns absolute paths. Fixtures named anything other than the
// positive_*.json / negative_*.json prefixes are silently skipped.
func discoverFixtures() ([]string, error) {
	// cmd/event-processor → ../../detector/wasm
	root := filepath.Join("..", "..", "detector", "wasm")

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		td := filepath.Join(root, e.Name(), "testdata")
		files, err := os.ReadDir(td)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if filepath.Ext(name) != ".json" {
				continue
			}
			if !strings.HasPrefix(name, "positive_") && !strings.HasPrefix(name, "negative_") {
				continue
			}
			paths = append(paths, filepath.Join(td, name))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// buildPolicyResponse converts a fixture's canned policy_response into a
// policy-enforcer Policy proto. Each rule's action ("BLOCK" / "NOTIFY")
// determines whether the rule lands in BlockBy or NotifyBy. Returns nil if
// the fixture declared no canned response, in which case the worker takes
// the "no rule action" path (no Notify, no KillPod).
func buildPolicyResponse(fr *fixturePolicyResponse) *enforcer_api.Policy {
	if fr == nil || len(fr.Rules) == 0 {
		return nil
	}
	out := &enforcer_api.Policy{}
	for _, r := range fr.Rules {
		ruleID := ""
		if r.RuleID != 0 {
			ruleID = strconv.FormatUint(r.RuleID, 10)
		}
		rule := &enforcer_api.Rule{
			Id:   ruleID,
			Name: r.RuleName,
			Type: enforcer_api.Rule_TYPE_RUNTIME,
			Rule: &enforcer_api.Rule_RuleJSON{},
		}
		switch strings.ToUpper(r.Action) {
		case "BLOCK":
			rule.Rule.Block = &enforcer_api.Rule_RuleJSON_Block{}
			out.BlockBy = append(out.BlockBy, rule)
		case "NOTIFY":
			rule.Rule.Notify = &enforcer_api.Rule_RuleJSON_Notify{}
			out.NotifyBy = append(out.NotifyBy, rule)
		default:
			// Unknown action — attach to neither slice; the worker
			// will treat the rule as a no-op. The fixture is
			// almost certainly broken; the calling test will catch
			// the resulting expectation mismatch.
			log.Warn().Msgf("buildPolicyResponse: unknown action %q on rule %d", r.Action, r.RuleID)
		}
	}
	return out
}

// publishWire injects raw fixture wire bytes directly onto runtime_events
// using the publisher's underlying amqp channel. We bypass
// lib/rabbit.MessageBroker.Publish because that helper re-marshals through
// proto.Marshal — and we explicitly want the wire bytes from disk, not a
// re-encoded version of them.
func publishWire(ctx context.Context, body []byte) error {
	if runtimePublisher == nil || runtimePublisher.Channel == nil {
		return errors.New("runtimePublisher is nil or has no channel")
	}
	return runtimePublisher.Channel.PublishWithContext(
		ctx,
		"", // default exchange
		runtimePublisher.Queue.Name,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Body:         body,
		},
	)
}
