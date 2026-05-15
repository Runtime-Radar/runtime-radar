package detector

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
)

func TestMatcher(t *testing.T) {
	t.Parallel()

	allDetectors := []detector_api.Detector{
		&TestExecAll{},
		&TestKprobeTCP{},
		&TestKprobeFile1{},
		&TestKprobeFile2{},
		&TestKprobeAll{},
	}

	testcases := []struct {
		Name       string
		Detectors  []detector_api.Detector
		EventType  string
		FuncNames  []string
		MatchedIDs []string
	}{
		{
			"Test exec all detector",
			allDetectors,
			"PROCESS_EXEC",
			[]string{},
			[]string{"TEST_EXEC_ALL"},
		},
		{
			"Test kprobe tcp_connect detector",
			allDetectors,
			"PROCESS_KPROBE",
			[]string{"tcp_connect"},
			[]string{"TEST_KPROBE_ALL", "TEST_KPROBE_TCP"},
		},
		{
			"Test kprobe security_file_permission detector",
			allDetectors,
			"PROCESS_KPROBE",
			[]string{"security_file_permission"},
			[]string{"TEST_KPROBE_ALL", "TEST_KPROBE_FILE_1", "TEST_KPROBE_FILE_2"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			matcher, err := NewMatcher(context.Background(), tc.Detectors...)
			if err != nil {
				t.Fatalf("Can't create matcher: %v", err)
			}

			chain := matcher.MatchChain(tc.EventType, tc.FuncNames...)
			ids := getIDs(chain)

			if diff := cmp.Diff(ids, tc.MatchedIDs,
				cmpopts.SortSlices(func(a, b string) bool {
					return a < b // particular way of sorting does not matter
				}),
			); diff != "" {
				t.Fatalf("Expected detectors != actual: %s", diff)
			}
		})
	}
}

func getIDs(chain Chain) (ids []string) {
	for k := range chain {
		ids = append(ids, k.ID)
	}

	return
}

type TestExecAll struct {
	detector_api.Detector
}

func (d *TestExecAll) Info(_ context.Context, _ *detector_api.InfoReq) (*detector_api.InfoResp, error) {
	resp := &detector_api.InfoResp{
		Id:      "TEST_EXEC_ALL",
		Version: 1,
	}

	return resp, nil
}

func (d *TestExecAll) TriggerCriteria(_ context.Context, _ *detector_api.TriggerCriteriaReq) (*detector_api.TriggerCriteriaResp, error) {
	resp := &detector_api.TriggerCriteriaResp{
		Criteria: map[string]*detector_api.TriggerCriteriaResp_FuncNames{
			"PROCESS_EXEC": {
				FuncNames: []string{"*"},
			},
		},
	}

	return resp, nil
}

type TestKprobeTCP struct {
	detector_api.Detector
}

func (d *TestKprobeTCP) Info(_ context.Context, _ *detector_api.InfoReq) (*detector_api.InfoResp, error) {
	resp := &detector_api.InfoResp{
		Id:      "TEST_KPROBE_TCP",
		Version: 1,
	}

	return resp, nil
}

func (d *TestKprobeTCP) TriggerCriteria(_ context.Context, _ *detector_api.TriggerCriteriaReq) (*detector_api.TriggerCriteriaResp, error) {
	resp := &detector_api.TriggerCriteriaResp{
		Criteria: map[string]*detector_api.TriggerCriteriaResp_FuncNames{
			"PROCESS_KPROBE": {
				FuncNames: []string{"tcp_connect", "tcp_close", "tcp_sendmsg"},
			},
		},
	}

	return resp, nil
}

type TestKprobeFile1 struct {
	detector_api.Detector
}

func (d *TestKprobeFile1) Info(_ context.Context, _ *detector_api.InfoReq) (*detector_api.InfoResp, error) {
	resp := &detector_api.InfoResp{
		Id:      "TEST_KPROBE_FILE_1",
		Version: 1,
	}

	return resp, nil
}

func (d *TestKprobeFile1) TriggerCriteria(_ context.Context, _ *detector_api.TriggerCriteriaReq) (*detector_api.TriggerCriteriaResp, error) {
	resp := &detector_api.TriggerCriteriaResp{
		Criteria: map[string]*detector_api.TriggerCriteriaResp_FuncNames{
			"PROCESS_KPROBE": {
				FuncNames: []string{"security_file_permission", "security_mmap_file", "security_path_truncate"},
			},
		},
	}

	return resp, nil
}

type TestKprobeFile2 struct {
	detector_api.Detector
}

func (d *TestKprobeFile2) Info(_ context.Context, _ *detector_api.InfoReq) (*detector_api.InfoResp, error) {
	resp := &detector_api.InfoResp{
		Id:      "TEST_KPROBE_FILE_2",
		Version: 1,
	}

	return resp, nil
}

func (d *TestKprobeFile2) TriggerCriteria(_ context.Context, _ *detector_api.TriggerCriteriaReq) (*detector_api.TriggerCriteriaResp, error) {
	resp := &detector_api.TriggerCriteriaResp{
		Criteria: map[string]*detector_api.TriggerCriteriaResp_FuncNames{
			"PROCESS_KPROBE": {
				FuncNames: []string{"security_file_permission", "sys_write"},
			},
		},
	}

	return resp, nil
}

type TestKprobeAll struct {
	detector_api.Detector
}

func (d *TestKprobeAll) Info(_ context.Context, _ *detector_api.InfoReq) (*detector_api.InfoResp, error) {
	resp := &detector_api.InfoResp{
		Id:      "TEST_KPROBE_ALL",
		Version: 1,
	}

	return resp, nil
}

func (d *TestKprobeAll) TriggerCriteria(_ context.Context, _ *detector_api.TriggerCriteriaReq) (*detector_api.TriggerCriteriaResp, error) {
	resp := &detector_api.TriggerCriteriaResp{
		Criteria: map[string]*detector_api.TriggerCriteriaResp_FuncNames{
			"PROCESS_KPROBE": {},
		},
	}

	return resp, nil
}
