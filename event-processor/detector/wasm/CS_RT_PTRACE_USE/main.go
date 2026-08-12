//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PTRACE_USE"
	Name        = "Code injection into executable process through ptrace"
	Description = "The detector detects code injection into an executable process through the ptrace() system call, which may indicate attempts to masquerade as a legitimate process, access the process memory and system or network resources, and escalate privileges."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeInjectionNoArgs  = "Detected a code injection into the process with PID `%d` using the `ptrace(%s)` system call by the `%s` process"
	KprobeInjectionDefault = "Detected a code injection into the process with PID `%d` using the `ptrace(%s)` system call by the `%s` process, which was started using the `%s` arguments"
	KprobeOtherNoArgs      = "Detected a connection to the process with PID `%d` using the `ptrace(%s)` system call initiated by the `%s` process"
	KprobeOtherDefault     = "Detected a connection to the process with PID `%d` using the `ptrace(%s)` system call initiated by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"sys_ptrace"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

var (
	mitreTactics = []*api.MitreTactic{
		{
			Id: "TA0003",
			Techniques: []string{
				"T1543",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1543",
				"T1055.008",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1055.008",
			},
		},
	}
)

var (
	// Taken from here (link from ESC): https://github.com/torvalds/linux/blob/master/include/uapi/linux/ptrace.h
	ptraceMaliciousRequests = []int32{
		4,     // PTRACE_POKETEXT (Copy word *data* to the address *addr*)
		5,     // PTRACE_POKEDATA (Copy word *data* to the address *addr*)
		6,     // PTRACE_POKEUSR (Copy the word *data* to offset *addr* in the traced process' USER area)
		13,    // PTRACE_SETREGS (Set all general purpose registers used by a processes)
		15,    // PTRACE_SETFPREGS (Set all floating point registers used by a processes)
		16,    // PTRACE_ATTACH (Attach to a process with SIGSTOP sending)
		17,    // PTRACE_DETACH (Detach from a traced process)
		19,    // PTRACE_SETFPXREGS (Set all extended floating point registers used by a processes)
		16899, // PTRACE_SETSIGINFO (Set signal information for a traced process)
		16901, // PTRACE_SETREGSET (Modify a traced process registers)
		16902, // PTRACE_SEIZE (Attach to a process without SIGSTOP sending)
		16903, // PTRACE_INTERRUPT (Try to stop traced process)
	}

	ptraceRequestTypes = map[int32]string{
		0:     "PTRACE_TRACEME",
		1:     "PTRACE_PEEKTEXT",
		2:     "PTRACE_PEEKDATA",
		3:     "PTRACE_PEEKUSR",
		4:     "PTRACE_POKETEXT",
		5:     "PTRACE_POKEDATA",
		6:     "PTRACE_POKEUSR",
		7:     "PTRACE_CONT",
		8:     "PTRACE_KILL",
		9:     "PTRACE_SINGLESTEP",
		12:    "PTRACE_GETREGS",
		13:    "PTRACE_SETREGS",
		14:    "PTRACE_GETFPREGS",
		15:    "PTRACE_SETFPREGS",
		16:    "PTRACE_ATTACH",
		17:    "PTRACE_DETACH",
		18:    "PTRACE_GETFPXREGS",
		19:    "PTRACE_SETFPXREGS",
		21:    "PTRACE_OLDSETOPTIONS",
		23:    "PTRACE_SET_SYSCALL",
		24:    "PTRACE_SYSCALL",
		25:    "PTRACE_GET_THREAD_AREA",
		26:    "PTRACE_SET_THREAD_AREA",
		30:    "PTRACE_ARCH_PRCTL",
		31:    "PTRACE_SYSEMU",
		32:    "PTRACE_SYSEMU_SINGLESTEP",
		33:    "PTRACE_SINGLEBLOCK",
		16896: "PTRACE_SETOPTIONS",
		16897: "PTRACE_GETEVENTMSG",
		16898: "PTRACE_GETSIGINFO",
		16899: "PTRACE_SETSIGINFO",
		16900: "PTRACE_GETREGSET",
		16901: "PTRACE_SETREGSET",
		16902: "PTRACE_SEIZE",
		16903: "PTRACE_INTERRUPT",
		16904: "PTRACE_LISTEN",
		16905: "PTRACE_PEEKSIGINFO",
		16906: "PTRACE_GETSIGMASK",
		16907: "PTRACE_SETSIGMASK",
		16908: "PTRACE_SECCOMP_GET_FILTER",
		16909: "PTRACE_SECCOMP_GET_METADATA",
		16910: "PTRACE_GET_SYSCALL_INFO",
		16914: "PTRACE_SET_SYSCALL_INFO",
	}
)

// init is required for TinyGo to compile to Wasm.
func init() {
	api.RegisterDetector(Detector{})
}

type Detector struct{}

func (d Detector) Info(ctx context.Context, req *api.InfoReq) (*api.InfoResp, error) {
	return &api.InfoResp{
		Id:             ID,
		Name:           Name,
		Description:    Description,
		Version:        Version,
		Author:         Author,
		License:        License,
		TacticsCovered: mitreTactics,
	}, nil
}

func (d Detector) TriggerCriteria(ctx context.Context, req *api.TriggerCriteriaReq) (*api.TriggerCriteriaResp, error) {
	resp := &api.TriggerCriteriaResp{
		Criteria: make(map[string]*api.TriggerCriteriaResp_FuncNames, len(triggerCriteria)),
	}

	for k, v := range triggerCriteria {
		resp.Criteria[k] = &api.TriggerCriteriaResp_FuncNames{FuncNames: v}
	}

	return resp, nil
}

func (d Detector) Detect(ctx context.Context, req *api.DetectReq) (*api.DetectResp, error) {
	// Detector info added to DetectResp because detector info is always correlated to response, thus
	// to avoid +1 Wasm call on detect.
	resp := &api.DetectResp{
		// Default response indicates that nothing detected (this is redundant and put here just for reference,
		// as Severity == api.DetectResp_NONE == 0 when omitted (default zero value)).
		Severity: api.DetectResp_NONE,
	}

	event := req.GetEvent().GetEvent()

	switch ev := event.(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		binaryArgs := kprobe.GetProcess().GetArguments()
		functionName := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if strings.HasSuffix(functionName, "sys_ptrace") {
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >=1", len(args))
			}
			req := args[0].GetIntArg()
			pid := args[1].GetIntArg()

			switch {
			case slices.Contains(ptraceMaliciousRequests, req) && (binaryArgs == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(KprobeInjectionNoArgs, pid, ptraceRequestTypes[req], binary)
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			case slices.Contains(ptraceMaliciousRequests, req) && (binaryArgs != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(KprobeInjectionDefault, pid, ptraceRequestTypes[req], binary, binaryArgs)
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			case !slices.Contains(ptraceMaliciousRequests, req) && (binaryArgs == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(KprobeOtherNoArgs, pid, ptraceRequestTypes[req], binary)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			case !slices.Contains(ptraceMaliciousRequests, req) && (binaryArgs != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(KprobeOtherDefault, pid, ptraceRequestTypes[req], binary, binaryArgs)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

		return resp, nil

	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

// tacticTechniques is a constructor for *api.MitreTactic which makes its initialization less verbose.
func tacticTechniques(tacticID string, techniqueIDs ...string) *api.MitreTactic {
	return &api.MitreTactic{
		Id:         tacticID,
		Techniques: techniqueIDs,
	}
}

