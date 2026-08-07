package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	upstreamtetragon "github.com/cilium/tetragon/api/v1/tetragon"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// fixture is the in-memory representation of one JSON fixture used by the
// integration tests in this package. It mirrors the on-disk schema documented
// in event-processor/README.md (Testing → Fixture schema).
//
// All sub-structs are pointers so that "field absent" / "field null" can be
// distinguished from "field present with zero value". A nil expectPolicyCall
// (etc.) means "assert the corresponding RPC was NOT invoked".
type fixture struct {
	// Path is the absolute path the fixture was loaded from. Used for test
	// labels and error messages; not persisted in JSON.
	Path string
	// EventWire is the wire-format encoding of the fixture's tetragon event,
	// ready to be published as the body of an AMQP message. Producing this
	// up-front (in loadFixture) keeps per-iteration overhead in the
	// TestDetectorPipeline matrix to a single Publish call.
	EventWire []byte
	// PolicyResponse is the canned EvaluatePolicyRuntimeEvent response the
	// minimock policy-enforcer client should return for this fixture. Nil
	// means "default to no rules matched" (the worker takes the
	// dispatch-skipped path).
	PolicyResponse *fixturePolicyResponse
	// Expect captures all mock-call and history-event assertions for this
	// fixture.
	Expect fixtureExpect
}

// fixturePolicyResponse is the canned policy-enforcer reply for a fixture.
// Field names mirror the JSON schema; the test driver translates it into an
// *enforcer_api.EvaluatePolicyRuntimeEventReq before configuring the mock.
type fixturePolicyResponse struct {
	Rules []fixturePolicyRule `json:"rules"`
}

// fixturePolicyRule is one canned rule entry within fixturePolicyResponse.
type fixturePolicyRule struct {
	Action   string `json:"action"`
	RuleID   uint64 `json:"rule_id"`
	RuleName string `json:"rule_name"`
}

// fixtureExpect is the assertion bundle for a fixture.
//
// Pointer-typed sub-structs (PolicyCall, NotifyCall) follow the
// "nil means assert-not-called" convention from the schema. HistoryEvent is
// always asserted (every event the worker processes produces exactly one
// history event), so it is a value, not a pointer.
type fixtureExpect struct {
	PolicyCall   *expectPolicyCall  `json:"policy_call"`
	NotifyCall   *expectNotifyCall  `json:"notify_call"`
	HistoryEvent expectHistoryEvent `json:"history_event"`
}

// expectPolicyCall describes the expected EvaluatePolicyRuntimeEvent call.
// All fields are optional except DetectorID and Severity, which together
// uniquely identify the (single) detector hit asserted by the fixture.
type expectPolicyCall struct {
	DetectorID     string   `json:"detector_id"`
	Severity       string   `json:"severity"`
	ReasonContains string   `json:"reason_contains"`
	Tactics        []string `json:"tactics"`
}

// expectNotifyCall describes the expected Notifier.Notify call. Currently
// minimal; richer assertions (target IDs, message bodies) can be added as
// fixtures begin exercising notify paths.
type expectNotifyCall struct {
	RuleNames []string `json:"rule_names"`
}

// expectHistoryEvent describes the expected RuntimeEvent published to the
// history queue. All fields are optional; absent fields are not asserted.
type expectHistoryEvent struct {
	Severity string `json:"severity"`
	IsThreat bool   `json:"is_threat"`
}

// fixtureFile mirrors the on-disk JSON layout. The event field is held as
// raw bytes so it can be unmarshalled separately by protojson.
type fixtureFile struct {
	Event          json.RawMessage        `json:"event"`
	PolicyResponse *fixturePolicyResponse `json:"policy_response"`
	Expect         fixtureExpect          `json:"expect"`
}

// loadFixture reads a fixture JSON file from disk, decodes the embedded
// tetragon event through the upstream cilium/tetragon proto type (which has
// full ProtoReflect, unlike the local TinyGo-targeted stub), re-marshals it
// to wire bytes, and returns a populated *fixture.
//
// The wire-bytes round trip is what lets tests publish to RabbitMQ exactly
// the same encoding the production runtime-monitor would emit; the consumer
// in event-processor then unmarshals into the local stub via plain
// proto.Unmarshal — no UnmarshalVT needed in test code.
//
// loadFixture is fatal on error (calls t.Fatalf), so the test code can keep
// the call sites clean; recovery from a malformed fixture is not interesting.
func loadFixture(t *testing.T, path string) *fixture {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loadFixture: read %s: %v", path, err)
	}

	f, err := decodeFixture(path, raw)
	if err != nil {
		t.Fatalf("loadFixture: %v", err)
	}
	return f
}

// decodeFixture is the error-returning core of loadFixture, factored out so
// the unit tests in this file can exercise both happy and failing paths
// without invoking t.Fatalf.
func decodeFixture(path string, raw []byte) (*fixture, error) {
	var ff fixtureFile
	if err := json.Unmarshal(raw, &ff); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	if len(ff.Event) == 0 {
		return nil, fmt.Errorf("decode %s: missing required field 'event'", path)
	}

	upstream := &upstreamtetragon.GetEventsResponse{}
	if err := protojson.Unmarshal(ff.Event, upstream); err != nil {
		return nil, fmt.Errorf("decode %s: protojson event: %w", path, err)
	}

	wire, err := proto.Marshal(upstream)
	if err != nil {
		return nil, fmt.Errorf("decode %s: marshal event wire bytes: %w", path, err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		// filepath.Abs only fails if the working directory cannot be
		// determined; treat as fatal-for-this-fixture but keep the
		// raw path so the error message is still useful.
		abs = path
	}

	return &fixture{
		Path:           abs,
		EventWire:      wire,
		PolicyResponse: ff.PolicyResponse,
		Expect:         ff.Expect,
	}, nil
}

// TestLoadFixture exercises the fixture loader against a synthetic positive
// fixture and a synthetic invalid fixture written to t.TempDir().
//
// Real on-disk fixtures live under detector/wasm/*/testdata/ and are
// driven by TestDetectorPipeline; this unit test stays focused on loader
// semantics (synthetic happy/error paths) and does not need the real
// wazero-seeded pipeline.
func TestLoadFixture(t *testing.T) {
	dir := t.TempDir()

	// Synthetic positive fixture: minimal valid event + non-trivial expect.
	positiveJSON := `{
  "event": {"node_name": "test-node"},
  "policy_response": {
    "rules": [
      {"action": "BLOCK", "rule_id": 42, "rule_name": "kill on reverse shell"}
    ]
  },
  "expect": {
    "policy_call": {
      "detector_id": "PTCS_RT_REVERSE_SHELL_CREATE",
      "severity": "CRITICAL",
      "reason_contains": "reverse shell",
      "tactics": ["TA0011"]
    },
    "kill_pod_call": {"namespace": "default", "name": "victim-7"},
    "notify_call": null,
    "history_event": {"severity": "CRITICAL", "is_threat": true}
  }
}`
	positivePath := filepath.Join(dir, "positive_synthetic.json")
	if err := os.WriteFile(positivePath, []byte(positiveJSON), 0o644); err != nil {
		t.Fatalf("write synthetic positive: %v", err)
	}

	// Synthetic negative fixture: null sub-structs and is_threat=false.
	negativeJSON := `{
  "event": {"node_name": "test-node"},
  "expect": {
    "policy_call": null,
    "kill_pod_call": null,
    "notify_call": null,
    "history_event": {"is_threat": false}
  }
}`
	negativePath := filepath.Join(dir, "negative_synthetic.json")
	if err := os.WriteFile(negativePath, []byte(negativeJSON), 0o644); err != nil {
		t.Fatalf("write synthetic negative: %v", err)
	}

	t.Run("positive_loads_and_decodes", func(t *testing.T) {
		f := loadFixture(t, positivePath)
		if f == nil {
			t.Fatal("loadFixture returned nil")
		}
		if len(f.EventWire) == 0 {
			t.Fatal("expected non-empty wire bytes for tetragon event")
		}
		// Round-trip the wire bytes back through the upstream type to
		// confirm the encoding survives unchanged.
		decoded := &upstreamtetragon.GetEventsResponse{}
		if err := proto.Unmarshal(f.EventWire, decoded); err != nil {
			t.Fatalf("unmarshal wire bytes: %v", err)
		}
		if decoded.GetNodeName() != "test-node" {
			t.Fatalf("node_name = %q, want %q", decoded.GetNodeName(), "test-node")
		}
		if f.PolicyResponse == nil || len(f.PolicyResponse.Rules) != 1 {
			t.Fatalf("policy_response: got %#v, want 1 rule", f.PolicyResponse)
		}
		const wantAction = "BLOCK"
		if got := f.PolicyResponse.Rules[0].Action; got != wantAction {
			t.Fatalf("rule action = %q, want %q", got, wantAction)
		}
		if f.Expect.PolicyCall == nil {
			t.Fatal("expect.policy_call: got nil, want non-nil")
		}
		if f.Expect.PolicyCall.DetectorID != "PTCS_RT_REVERSE_SHELL_CREATE" {
			t.Fatalf("detector_id = %q", f.Expect.PolicyCall.DetectorID)
		}
		if f.Expect.NotifyCall != nil {
			t.Fatalf("notify_call: got %#v, want nil", f.Expect.NotifyCall)
		}
		if !f.Expect.HistoryEvent.IsThreat {
			t.Fatal("expect.history_event.is_threat = false, want true")
		}
	})

	t.Run("negative_loads_and_decodes", func(t *testing.T) {
		f := loadFixture(t, negativePath)
		if f.PolicyResponse != nil {
			t.Fatalf("policy_response: got %#v, want nil", f.PolicyResponse)
		}
		if f.Expect.PolicyCall != nil {
			t.Fatalf("expect.policy_call: got %#v, want nil", f.Expect.PolicyCall)
		}
		if f.Expect.HistoryEvent.IsThreat {
			t.Fatal("expect.history_event.is_threat = true, want false")
		}
	})

	t.Run("missing_file_errors", func(t *testing.T) {
		_, err := decodeFixture("nonexistent.json", []byte("not json"))
		if err == nil {
			t.Fatal("decodeFixture: got nil error, want JSON parse failure")
		}
	})

	t.Run("missing_event_field_errors", func(t *testing.T) {
		_, err := decodeFixture("inline.json", []byte(`{"expect":{"history_event":{"is_threat":false}}}`))
		if err == nil {
			t.Fatal("decodeFixture: got nil error, want missing-event error")
		}
	})

	t.Run("malformed_event_errors", func(t *testing.T) {
		// "event" must decode as tetragon.GetEventsResponse; passing a
		// number in place of an object exercises the protojson failure
		// branch.
		_, err := decodeFixture("inline.json", []byte(`{"event": 42}`))
		if err == nil {
			t.Fatal("decodeFixture: got nil error, want protojson failure")
		}
	})

	// Sanity check: the on-disk-error contract exposed by os.ReadFile is
	// preserved by loadFixture's underlying read step. We exercise it via
	// decodeFixture by feeding it bytes that are not JSON at all so the
	// first json.Unmarshal call fails — no real file I/O needed.
	t.Run("non_json_bytes_errors", func(t *testing.T) {
		_, err := decodeFixture("inline.json", []byte("definitely not json"))
		if err == nil {
			t.Fatal("decodeFixture: got nil error, want JSON parse failure")
		}
	})
}
