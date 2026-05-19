package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/model"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	connectTimeout  = time.Second
	dispatchTimeout = time.Second
	metricsInterval = 15 * time.Second
)

var (
	errClientContextCanceled = errors.New("client context canceled")
)

// Monitor is interface of Tetra monitoring instance.
type Monitor interface {
	Config() (sel config.Selector, cfg *model.Config)
	SetConfig(sel config.Selector, cfg *model.Config)
	Init(ctx context.Context, cfg *model.Config) error
	Update(sel config.Selector, cfg *model.Config)
	LastInitErr() error
	Run(stop <-chan struct{}) error
	Events() <-chan *tetragon.GetEventsResponse
}

// Tetra is implementation of Monitor.
type Tetra struct {
	Version string

	sensorsClient tetragon.FineGuidanceSensorsClient
	eventsClient  tetragon.FineGuidanceSensors_GetEventsClient

	eventsCtx         context.Context
	eventsCancelCause context.CancelCauseFunc

	config   *model.Config
	selector config.Selector
	configMu sync.RWMutex

	ready  chan struct{}
	update chan bool
	events chan *tetragon.GetEventsResponse

	lastInitErr error // err occurred during last config initialization
	errMu       sync.RWMutex
}

// NewTetra creates new Tetra instance. It returns any possible error and closing function which is supposed to be put in defer statement in main.
func NewTetra(address string, bufferSize int) (*Tetra, func() error, error) {
	connCtx, connCancel := context.WithTimeout(context.Background(), connectTimeout)
	defer connCancel()

	conn, err := grpc.DialContext(connCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("can't dial tetragon gRPC service: %w", err)
	}

	sensors := tetragon.NewFineGuidanceSensorsClient(conn)

	resp, err := sensors.GetVersion(context.Background(), &tetragon.GetVersionRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("cant' get tetragon version: %w", err)
	}

	t := &Tetra{
		Version:       resp.GetVersion(),
		sensorsClient: sensors,

		ready:  make(chan struct{}),
		update: make(chan bool, 1),
		events: make(chan *tetragon.GetEventsResponse, bufferSize),
	}

	return t, conn.Close, nil
}

// Config returns current Tetra config. It's safe for concurrent use.
func (t *Tetra) Config() (sel config.Selector, cfg *model.Config) {
	t.configMu.RLock()
	defer t.configMu.RUnlock()

	return t.selector, t.config
}

// SetConfig sets new Tetra config (but does not apply it). It's safe for concurrent use.
func (t *Tetra) SetConfig(sel config.Selector, cfg *model.Config) {
	t.configMu.Lock()
	defer t.configMu.Unlock()

	t.selector = sel
	t.config = cfg
}

// Events returns read-only channel for consuming runtime events elsewhere.
func (t *Tetra) Events() <-chan *tetragon.GetEventsResponse {
	return t.events
}

// Update reinitializes Tetra based on config.Selector and given config.
func (t *Tetra) Update(sel config.Selector, cfg *model.Config) {
	t.configMu.Lock()
	defer t.configMu.Unlock()

	// Or t.SetConfig
	t.config = cfg
	t.selector = sel

	select {
	case t.update <- true:
	default:
		// Do nothing. If sending blocked, there is already update pending.
	}
}

// Init should be run on new Tetra instance, in order to initialize and prepare it for Run. Unlike New, Init takes
// an ctx argument, which can be configured for cancellation on init phase. Cancellation or expiration of ctx
// does not affect further stream processing.
func (t *Tetra) Init(ctx context.Context, cfg *model.Config) error {
	if cfg == nil {
		return errors.New("nil Tetra config")
	}

	sel := config.Selector{true, true, true}

	t.config = cfg
	t.selector = sel

	if err := t.initBySelector(ctx, sel, cfg); err != nil {
		return fmt.Errorf("can't init Tetra config: %w", err)
	}

	close(t.ready)

	return nil
}

func (t *Tetra) initBySelector(ctx context.Context, sel config.Selector, cfg *model.Config) (err error) {
	defer func() {
		t.errMu.Lock()
		t.lastInitErr = err
		t.errMu.Unlock()
	}()

	t.errMu.RLock()
	if t.lastInitErr != nil {
		sel = config.Selector{true, true, true}
	}
	t.errMu.RUnlock()

	if sel.EventsClient {
		if err := t.initEventsClient(ctx, cfg); err != nil {
			return err
		}
	}
	if sel.TracingPolicies {
		if err := t.initTracingPolicies(ctx, cfg); err != nil {
			return err
		}
		if err := t.initTracingPolicyStates(ctx, cfg); err != nil {
			return err
		}
	} else if sel.TracingPolicyStates {
		if err := t.initTracingPolicyStates(ctx, cfg); err != nil {
			return err
		}
	}

	return nil
}

//nolint:govet
func (t *Tetra) initEventsClient(ctx context.Context, cfg *model.Config) error {
	// eventsCtx SHOULD NOT be derived from ctx, as it may be used for different purpose,
	// for instance parent context can be configured for cancellation when invoking tetragon gRPC methods,
	// however ctx is checked for being valid before returning
	eventsCtx, eventsCancelCause := context.WithCancelCause(context.Background())
	eventsClient, err := t.sensorsClient.GetEvents(eventsCtx, &tetragon.GetEventsRequest{
		AllowList:          cfg.Config.AllowList,
		DenyList:           cfg.Config.DenyList,
		AggregationOptions: cfg.Config.AggregationOptions,
	})
	if err != nil {
		return fmt.Errorf("can't initialize tetragon events client: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		cancel := t.eventsCancelCause
		if cancel != nil {
			defer cancel(errClientContextCanceled)
		}

		t.eventsClient = eventsClient
		t.eventsCtx = eventsCtx
		t.eventsCancelCause = eventsCancelCause
	}

	return nil
}

func (t *Tetra) initTracingPolicies(ctx context.Context, cfg *model.Config) error {
	ltpResp, err := t.sensorsClient.ListTracingPolicies(ctx, &tetragon.ListTracingPoliciesRequest{})
	if err != nil {
		return fmt.Errorf("can't list tracing policies: %v", err)
	}
	log.Debug().Interface("policies", ltpResp.GetPolicies()).Msgf("Tetragon tracing policies before init tracing policies")

	// Re-initialization of tracing policies works in two steps:
	// 1. Delete existing policies
	// 2. Add requested policies
	//
	// This means that it covers all possible scenarios: addition/update/deletion of tracing policies, and tend to be more "secure",
	// as effectively everything is being recreated from scratch. This also makes us sure, that nothing is added to Tetragon via CRD,
	// or with use of some side channel apart from runtime-monitor.
	// However it could be possible to implement more precise logic and make changes only on added/updated/deleted
	// policies, not touching unchanged policies. In this scenario config.Selector and config.Diff
	// should be implemented accordingly to return not only bool value, indicating that tracing policies changed,
	// but more accurate structure with lists of policies to add and delete (where update == delete + add).

	for _, tps := range ltpResp.GetPolicies() {
		if _, err := t.sensorsClient.DeleteTracingPolicy(ctx, &tetragon.DeleteTracingPolicyRequest{Name: tps.Name}); err != nil {
			return fmt.Errorf("can't delete tracing policy '%s': %w", tps.Name, err)
		}
		log.Info().Msgf("Tetragon tracing policy '%s' deleted", tps.Name)
	}

	for name, tp := range cfg.Config.TracingPolicies {
		if _, err := t.sensorsClient.AddTracingPolicy(ctx, &tetragon.AddTracingPolicyRequest{Yaml: tp.GetYaml()}); err != nil {
			return fmt.Errorf("can't add tracing policy '%s': %w", name, err)
		}
		log.Info().Str("policy", tp.GetYaml()).Msgf("Tetragon tracing policy '%s' added", name)
	}

	return nil
}

func (t *Tetra) initTracingPolicyStates(ctx context.Context, cfg *model.Config) error {
	ltpResp, err := t.sensorsClient.ListTracingPolicies(ctx, &tetragon.ListTracingPoliciesRequest{})
	if err != nil {
		return fmt.Errorf("can't list tracing policies: %v", err)
	}
	log.Debug().Interface("policies", ltpResp.GetPolicies()).Msgf("Tetragon tracing policies before init tracing policy states")

	for _, tps := range ltpResp.GetPolicies() {
		policyName := tps.GetName()
		policyConfig, ok := cfg.Config.TracingPolicies[policyName]
		if !ok {
			return fmt.Errorf("loaded tracing policy '%s' has no configuration", policyName)
		}

		if isTracingPolicyEnabled(tps) != policyConfig.Enabled {
			if policyConfig.Enabled {
				if _, err := t.sensorsClient.EnableTracingPolicy(ctx, &tetragon.EnableTracingPolicyRequest{Name: policyName}); err != nil {
					return fmt.Errorf("can't enable tracing policy '%s': %w", policyName, err)
				}
				log.Info().Msgf("Tetragon tracing policy '%s' enabled", policyName)
			} else {
				if _, err := t.sensorsClient.DisableTracingPolicy(ctx, &tetragon.DisableTracingPolicyRequest{Name: policyName}); err != nil {
					return fmt.Errorf("can't disable tracing policy '%s': %w", policyName, err)
				}
				log.Info().Msgf("Tetragon tracing policy '%s' disabled", policyName)
			}
		}
	}

	return nil
}

// Run runs the monitor. Init should be invoked before Run.
func (t *Tetra) Run(stop <-chan struct{}) error {
	log.Info().Msgf("Tetra monitor started")
	defer log.Info().Msgf("Tetra monitor stopped")

	<-t.ready // <-- wait for ready status

	// There should always be exactly one worker on a stream
	worker := make(chan bool, 1)
	err := make(chan error, 1)
	worker <- true

	wg := &sync.WaitGroup{}

	defer close(t.events)
	defer wg.Wait()
	defer t.eventsCancelCause(errClientContextCanceled)

	go t.runMetrics(metricsInterval, stop)

	for {
		select {
		case <-t.update:
			sel, cfg := t.Config()
			if err := t.initBySelector(context.Background(), sel, cfg); err != nil {
				log.Error().Err(err).Interface("selector", sel).Interface("config", cfg).Msgf("Can't init tetra")
				// Just log, don't break the monitor loop
			}
		case <-worker:
			wg.Add(1)

			go func() {
				defer wg.Done()

				workerID := security.RandAlphaNum(5)
				log.Info().Msgf("Worker[%s] started", workerID)
				defer log.Info().Msgf("Worker[%s] stopped", workerID)

				err <- t.processStream(t.eventsCtx)
			}()
		case errProc := <-err:
			if !errors.Is(errProc, errClientContextCanceled) {
				// We assume that Tetragon has crashed or restarted and runtime-monitor needs to be restarted as well, in order to re-connect and configure it
				return fmt.Errorf("can't process stream: %w", errProc) // <-- return
			}

			// Restart all workers
			worker <- true
		case <-stop:
			return nil // <-- return
		}
	}
}

func (t *Tetra) LastInitErr() error {
	t.errMu.RLock()
	defer t.errMu.RUnlock()

	return t.lastInitErr
}

// processStream processes the events stream, it intercepts any possible panic and converts it to error.
// eventsCtx is a client context which will be checked in case of an error. As a result It lets returning
// specific error and distinguishing it from generic status error with codes.Canceled.
func (t *Tetra) processStream(eventsCtx context.Context) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if errPanic, ok := p.(error); ok {
				err = errPanic
			} else if str, ok := p.(string); ok {
				err = errors.New(str)
			} else {
				panic(p) // something very special happened
			}
		}
	}()

	for {
		resp, err := t.eventsClient.Recv()
		// Experimentally, it was discovered that when Tetragon crashes or being killed most of the time Recv returns
		// status error with codes.Unavailable. But sometimes, approx ~1/10 all cases, for some reason it returns codes.Canceled,
		// which makes error indistinguishable from the one returned in case of intentional context cancellation when re-initializing
		// the client. Usage of context.CancelCauseFunc in this scenario has no effect as well as Context method of a stream client.
		if err != nil {
			select {
			case <-eventsCtx.Done():
				return fmt.Errorf("can't receive message from events stream: %w; client context is done: %w", err, context.Cause(eventsCtx))
			default:
				return fmt.Errorf("can't receive message from events stream: %w", err)
			}
		}

		timer := time.NewTimer(dispatchTimeout)

		select {
		case t.events <- resp:
		case <-timer.C:
			metrics.TetragonEventsDroppedCount.Inc()
			log.Error().Interface("event", resp.GetEvent()).Msgf("Timeout dispatching event")
		}

		timer.Stop()
	}
}

func (t *Tetra) runMetrics(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return

		case <-ticker.C:
			metrics.TetragonEventsBufferSizeGauge.Set(float64(len(t.events)))
		}
	}
}

func isTracingPolicyEnabled(tps *tetragon.TracingPolicyStatus) bool {
	if tps.GetState() == tetragon.TracingPolicyState_TP_STATE_ENABLED {
		return true
	}

	return false
}
