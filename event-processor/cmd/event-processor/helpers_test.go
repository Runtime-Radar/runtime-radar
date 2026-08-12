package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	amqp "github.com/rabbitmq/amqp091-go"
	runtime_event "github.com/runtime-radar/runtime-radar/event-processor/api"
	kube_manager "github.com/runtime-radar/runtime-radar/kube-manager/api"
	notifier_api "github.com/runtime-radar/runtime-radar/notifier/api"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

// drainHistoryEvent blocks until either a history-queue delivery arrives on
// ch or deadline elapses. It unmarshals the delivery body into a
// *runtime_event.RuntimeEvent and returns it. The provided ctx is honored in
// addition to the deadline — whichever fires first wins.
//
// On timeout or context cancellation, the test is failed via t.Fatalf and
// nil is returned (unreachable, but keeps the call sites short).
//
// The delivery is implicitly auto-acked: this helper does not call
// delivery.Ack, so the consumer that produced ch must have been registered
// with autoAck=true. That matches the simple in-test consumer set up by
// TestMain.
//
//nolint:revive // *testing.T conventionally precedes context in test helpers
func drainHistoryEvent(t *testing.T, ctx context.Context, ch <-chan amqp.Delivery, deadline time.Duration) *runtime_event.RuntimeEvent {
	t.Helper()

	timeoutCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	select {
	case d, ok := <-ch:
		if !ok {
			t.Fatalf("drainHistoryEvent: history channel closed before delivery")
			return nil
		}
		ev := &runtime_event.RuntimeEvent{}
		if err := proto.Unmarshal(d.Body, ev); err != nil {
			t.Fatalf("drainHistoryEvent: unmarshal RuntimeEvent: %v", err)
			return nil
		}
		return ev
	case <-timeoutCtx.Done():
		// timeoutCtx is created with WithTimeout (no cause), so
		// timeoutCtx.Err() is the only signal we have — DeadlineExceeded
		// for the deadline path, Canceled if the parent ctx was
		// cancelled while we were blocked.
		t.Fatalf("drainHistoryEvent: timed out after %s waiting for history event (err: %v)", deadline, timeoutCtx.Err())
		return nil
	}
}

// purgeQueue removes all messages from the named queue. Intended for use as
// a t.Cleanup hook so each subtest starts with empty queues regardless of
// what the previous test produced.
//
// Failures are reported via t.Errorf (not Fatalf) so cleanup keeps running
// even when the queue server has misbehaved.
func purgeQueue(t *testing.T, channel *amqp.Channel, queueName string) {
	t.Helper()
	if channel == nil {
		t.Errorf("purgeQueue: nil amqp channel for queue %q", queueName)
		return
	}
	if _, err := channel.QueuePurge(queueName, false /* noWait */); err != nil {
		t.Errorf("purgeQueue: %s: %v", queueName, err)
	}
}

// assertPolicyCall asserts that the captured EvaluatePolicyRuntimeEvent
// invocations match the expectation in the fixture.
//
// Semantics:
//   - want == nil:  zero invocations expected
//   - want != nil:  exactly one invocation expected, whose Result.Events
//     contains an entry matching want.DetectorID +
//     want.Severity. The captured request only carries
//     detector id + severity (see policy-enforcer enforcer.proto),
//     so reason_contains and tactics are checked separately by
//     assertReasonContains and assertTactics against the
//     drained RuntimeEvent — that is the same data
//     policy-enforcer would observe in production.
//
// The assertion is intentionally loose on field ordering: we use
// cmpopts.SortSlices on string slices and protocmp.Transform() on the
// captured proto so that detector emit order does not affect outcomes.
func assertPolicyCall(t *testing.T, want *expectPolicyCall, calls []*enforcer_api.EvaluatePolicyRuntimeEventReq) {
	t.Helper()

	if want == nil {
		if len(calls) != 0 {
			t.Errorf("assertPolicyCall: got %d EvaluatePolicyRuntimeEvent calls, want 0", len(calls))
		}
		return
	}

	if len(calls) != 1 {
		t.Errorf("assertPolicyCall: got %d EvaluatePolicyRuntimeEvent calls, want 1", len(calls))
		return
	}
	got := calls[0]
	if got == nil || got.GetResult() == nil {
		t.Errorf("assertPolicyCall: captured request missing Result")
		return
	}

	// Check that the Events slice contains an entry matching detector
	// id + severity. We don't pin the slice order — workers may emit
	// detectors in arbitrary order. Severity comparison is
	// case-insensitive: fixtures store uppercase ("HIGH") while
	// detectors emit lowercase ("high").
	found := false
	for _, e := range got.GetResult().GetEvents() {
		if e.GetDetectorId() == want.DetectorID && strings.EqualFold(e.GetSeverity(), want.Severity) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("assertPolicyCall: no Events entry matched detector_id=%q severity=%q (case-insensitive); got events: %v",
			want.DetectorID, want.Severity, got.GetResult().GetEvents())
	}
}

// assertTactics asserts that the drained RuntimeEvent's Threats slice
// covers every tactic ID listed in want.Tactics (for the threat emitted
// by want.DetectorID). An empty want.Tactics is a no-op. Extra tactics on
// the threat are tolerated; only the assertion side must be a subset.
func assertTactics(t *testing.T, want *expectPolicyCall, ev *runtime_event.RuntimeEvent) {
	t.Helper()

	if want == nil || len(want.Tactics) == 0 {
		return
	}
	if ev == nil {
		t.Errorf("assertTactics: nil RuntimeEvent")
		return
	}

	have := make(map[string]struct{}, 8)
	matchedDetector := false
	for _, th := range ev.GetThreats() {
		if th.GetDetector().GetId() != want.DetectorID {
			continue
		}
		matchedDetector = true
		for _, mt := range th.GetTacticsCovered() {
			have[mt.GetId()] = struct{}{}
		}
	}
	if !matchedDetector {
		t.Errorf("assertTactics: detector %q produced no threat in RuntimeEvent.Threats; want tactics %v",
			want.DetectorID, want.Tactics)
		return
	}
	missing := make([]string, 0, len(want.Tactics))
	for _, w := range want.Tactics {
		if _, ok := have[w]; !ok {
			missing = append(missing, w)
		}
	}
	if len(missing) > 0 {
		gotTactics := make([]string, 0, len(have))
		for k := range have {
			gotTactics = append(gotTactics, k)
		}
		t.Errorf("assertTactics: detector %q missing tactics %v; got tactics %v",
			want.DetectorID, missing, gotTactics)
	}
}

// assertReasonContains is a small extension over assertPolicyCall that
// checks the threat reason text against expect.policy_call.reason_contains
// (case-insensitive). It reads the reason from the drained RuntimeEvent's
// Threats slice — the threats list emitted by the WASM detector chain and
// surfaced on the history queue.
//
// Diagnostic note: the error message distinguishes "detector did not fire
// at all" from "detector fired but reason text did not match" so that
// failures point at the right root cause.
func assertReasonContains(t *testing.T, want *expectPolicyCall, ev *runtime_event.RuntimeEvent) {
	t.Helper()

	if want == nil || want.ReasonContains == "" {
		return
	}
	if ev == nil {
		t.Errorf("assertReasonContains: nil RuntimeEvent")
		return
	}

	matchedDetector := false
	gotReasons := make([]string, 0, len(ev.GetThreats()))
	for _, threat := range ev.GetThreats() {
		if threat.GetDetector().GetId() != want.DetectorID {
			continue
		}
		matchedDetector = true
		if caseInsensitiveContains(threat.GetReason(), want.ReasonContains) {
			return
		}
		gotReasons = append(gotReasons, threat.GetReason())
	}
	if !matchedDetector {
		t.Errorf("assertReasonContains: detector %q produced no threat in RuntimeEvent.Threats; want reason containing %q",
			want.DetectorID, want.ReasonContains)
		return
	}
	t.Errorf("assertReasonContains: detector %q fired but no threat reason contains %q (case-insensitive); got reasons: %v",
		want.DetectorID, want.ReasonContains, gotReasons)
}

// assertKillPodCall asserts that the captured KillPod invocations match the
// expectation. nil want means "no Kill call expected".
func assertKillPodCall(t *testing.T, want *expectKillPodCall, calls []*kube_manager.KillPodReq) {
	t.Helper()

	if want == nil {
		if len(calls) != 0 {
			t.Errorf("assertKillPodCall: got %d KillPod calls, want 0", len(calls))
		}
		return
	}

	if len(calls) != 1 {
		t.Errorf("assertKillPodCall: got %d KillPod calls, want 1", len(calls))
		return
	}
	got := calls[0]
	wantReq := &kube_manager.KillPodReq{
		Namespace: want.Namespace,
		Name:      want.Name,
	}
	if diff := cmp.Diff(wantReq, got, protocmp.Transform()); diff != "" {
		t.Errorf("assertKillPodCall: KillPod request mismatch (-want +got):\n%s", diff)
	}
}

// assertNotifyCall asserts that the captured Notify invocations match the
// expectation.
//
// Today the only assertion is on the set of rule names attached to the
// notifications: ordering is irrelevant, so we use cmpopts.SortSlices to
// flatten the comparison. Richer assertions (target IDs, message bodies)
// can be added as fixtures start exercising those paths.
func assertNotifyCall(t *testing.T, want *expectNotifyCall, calls []*notifier_api.NotifyReq) {
	t.Helper()

	if want == nil {
		if len(calls) != 0 {
			t.Errorf("assertNotifyCall: got %d Notify calls, want 0", len(calls))
		}
		return
	}

	if len(calls) != 1 {
		t.Errorf("assertNotifyCall: got %d Notify calls, want 1", len(calls))
		return
	}

	gotRules := make([]string, 0, len(calls[0].GetNotifications()))
	for _, msg := range calls[0].GetNotifications() {
		gotRules = append(gotRules, msg.GetRuntimeEvent().GetRuleName())
	}

	sortStrings := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(want.RuleNames, gotRules, sortStrings); diff != "" {
		t.Errorf("assertNotifyCall: rule names mismatch (-want +got):\n%s", diff)
	}
}

// assertHistoryEvent asserts that the drained RuntimeEvent matches the
// fixture's expect.history_event.
//
// Semantics:
//   - want.IsThreat compares against len(got.Threats) > 0. A "threat" in
//     the fixture vocabulary is a detector emission; that lives on
//     RuntimeEvent.Threats. (RuntimeEvent.IsIncident is only set when a
//     policy rule actually fires, which most positive fixtures don't
//     exercise.)
//   - When want.IsThreat is true and want.Severity is non-empty, at least
//     one Threat in got.Threats must have a matching severity.
//     Comparison is case-insensitive: fixtures store "HIGH"/"MEDIUM"/etc.
//     while detectors emit lowercase variants.
//   - When want.IsThreat is false, want.Severity is ignored (negatives
//     have no threats to carry severity).
func assertHistoryEvent(t *testing.T, want expectHistoryEvent, got *runtime_event.RuntimeEvent) {
	t.Helper()

	if got == nil {
		t.Errorf("assertHistoryEvent: got nil RuntimeEvent")
		return
	}

	gotIsThreat := len(got.GetThreats()) > 0
	if gotIsThreat != want.IsThreat {
		t.Errorf("assertHistoryEvent: is_threat = %v (len(threats)=%d), want %v",
			gotIsThreat, len(got.GetThreats()), want.IsThreat)
		return
	}

	if !want.IsThreat {
		return
	}

	if want.Severity == "" {
		return
	}

	for _, th := range got.GetThreats() {
		if strings.EqualFold(th.GetSeverity(), want.Severity) {
			return
		}
	}

	gotSeverities := make([]string, 0, len(got.GetThreats()))
	for _, th := range got.GetThreats() {
		gotSeverities = append(gotSeverities, th.GetSeverity())
	}
	t.Errorf("assertHistoryEvent: no threat with severity %q (case-insensitive); got severities: %v",
		want.Severity, gotSeverities)
}

// caseInsensitiveContains is a small helper used by the policy-call reason
// assertion. Kept here (not in the loader) to keep all assertion utilities
// in one file.
func caseInsensitiveContains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
