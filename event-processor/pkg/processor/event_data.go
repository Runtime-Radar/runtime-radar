package processor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/arch"
	history_model "github.com/runtime-radar/runtime-radar/history-api/pkg/model"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// eventData keeps some some general data of Tetragon runtime event mostly for use in invocation of notifier.
type eventData struct {
	RegisteredAt time.Time `json:"registered_at"`
	EventType    string    `json:"event_type"`

	PodNamespace string `json:"pod_namespace"`
	PodName      string `json:"pod_name"`

	ContainerID    string `json:"container_id"`
	ContainerName  string `json:"container_name"`
	ContainerImage string `json:"container_image"`

	FunctionName   string `json:"function_name"`
	FunctionArgs   string `json:"function_args"`
	FunctionReturn string `json:"function_return"`

	ProcessPID          *uint32  `json:"process_pid"`
	ProcessUID          *uint32  `json:"process_uid"`
	ProcessBinary       string   `json:"process_binary"`
	ProcessArguments    string   `json:"process_arguments"`
	ProcessSetuid       *uint32  `json:"process_setuid"`
	ProcessSetgid       *uint32  `json:"process_setgid"`
	ProcessCapEffective []string `json:"process_cap_effective"`
	ProcessCapPerimtted []string `json:"process_cap_permitted"`

	ParentPID       *uint32 `json:"parent_pid"`
	ParentUID       *uint32 `json:"parent_uid"`
	ParentBinary    string  `json:"parent_binary"`
	ParentArguments string  `json:"parent_arguments"`

	NodeName string `json:"node_name"`
}

// getEventData parses event, fills and returns corresponding eventData struct.
func getEventData(e *tetragon.GetEventsResponse) (*eventData, error) {
	ed := &eventData{
		RegisteredAt: e.GetTime().AsTime(),
		NodeName:     e.GetNodeName(),
	}

	switch ev := e.GetEvent().(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		exec := ev.ProcessExec
		ed.EventType = history_model.RuntimeEventTypeProcessExec

		setCommonAttributes(exec.GetProcess(), exec.GetParent(), ed)
	case *tetragon.GetEventsResponse_ProcessExit:
		exit := ev.ProcessExit
		ed.EventType = history_model.RuntimeEventTypeProcessExit

		setCommonAttributes(exit.GetProcess(), exit.GetParent(), ed)
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		ed.EventType = history_model.RuntimeEventTypeProcessKprobe
		ed.FunctionName = kprobe.GetFunctionName()

		args, err := marshalKprobeArgs(kprobe.GetArgs())
		if err != nil {
			return nil, err
		}
		ed.FunctionArgs = args

		ret, err := marshalKprobeReturn(kprobe.GetReturn())
		if err != nil {
			return nil, err
		}
		ed.FunctionReturn = ret

		setCommonAttributes(kprobe.GetProcess(), kprobe.GetParent(), ed)
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		tp := ev.ProcessTracepoint
		ed.EventType = history_model.RuntimeEventTypeProcessTracepoint

		args, err := marshalKprobeArgs(tp.GetArgs())
		if err != nil {
			return nil, err
		}
		ed.FunctionArgs = args

		setCommonAttributes(tp.GetProcess(), tp.GetParent(), ed)
	case *tetragon.GetEventsResponse_ProcessLoader:
		ed.EventType = history_model.RuntimeEventTypeProcessLoader

		setCommonAttributes(ev.ProcessLoader.GetProcess(), nil, ed)
	case *tetragon.GetEventsResponse_ProcessUprobe:
		uprobe := ev.ProcessUprobe
		ed.EventType = history_model.RuntimeEventTypeProcessUprobe

		args, err := marshalKprobeArgs(uprobe.GetArgs())
		if err != nil {
			return nil, err
		}
		ed.FunctionArgs = args

		setCommonAttributes(uprobe.GetProcess(), uprobe.GetParent(), ed)
	}

	return ed, nil
}

// getEventAndFunc parses event and returns corresponding event type and function.
func getEventAndFunc(ev *tetragon.GetEventsResponse) (string, string) {
	switch e := ev.GetEvent().(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		return history_model.RuntimeEventTypeProcessExec, ""
	case *tetragon.GetEventsResponse_ProcessExit:
		return history_model.RuntimeEventTypeProcessExit, ""
	case *tetragon.GetEventsResponse_ProcessKprobe:
		funcName := e.ProcessKprobe.GetFunctionName()
		_, funcName = arch.CutSyscallPrefix(funcName)
		return history_model.RuntimeEventTypeProcessKprobe, funcName
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		return history_model.RuntimeEventTypeProcessTracepoint, ""
	case *tetragon.GetEventsResponse_ProcessLoader:
		return history_model.RuntimeEventTypeProcessLoader, ""
	case *tetragon.GetEventsResponse_ProcessUprobe:
		sym := e.ProcessUprobe.GetSymbol()
		return history_model.RuntimeEventTypeProcessUprobe, sym
	default:
		return history_model.RuntimeEventTypeUndef, ""
	}
}

// setCommonAttributes sets process' and parent's (if not nil) attributes,
// that are common for different types of events.
func setCommonAttributes(process, parent *tetragon.Process, ed *eventData) {
	ed.ProcessBinary = process.GetBinary()
	ed.ProcessArguments = process.GetArguments()
	ed.ProcessUID = uint32WrapperToPtr(process.GetUid())
	ed.ProcessPID = uint32WrapperToPtr(process.GetPid())
	ed.ProcessCapEffective = capsToStrings(process.GetCap().GetEffective())
	ed.ProcessCapPerimtted = capsToStrings(process.GetCap().GetPermitted())
	ed.ProcessSetuid = uint32WrapperToPtr(process.GetBinaryProperties().GetSetuid())
	ed.ProcessSetgid = uint32WrapperToPtr(process.GetBinaryProperties().GetSetgid())

	ed.ParentBinary = parent.GetBinary()
	ed.ParentArguments = parent.GetArguments()
	ed.ParentUID = uint32WrapperToPtr(parent.GetUid())
	ed.ParentPID = uint32WrapperToPtr(parent.GetPid())

	ed.PodNamespace = process.GetPod().GetNamespace()
	ed.PodName = process.GetPod().GetName()

	ed.ContainerID = process.GetPod().GetContainer().GetId()
	ed.ContainerName = process.GetPod().GetContainer().GetName()
	ed.ContainerImage = process.GetPod().GetContainer().GetImage().GetName()
}

func marshalKprobeArgs(args []*tetragon.KprobeArgument) (string, error) {
	raws := make([]json.RawMessage, 0, len(args))
	mo := marshalOptions()

	for _, arg := range args {
		r, err := mo.Marshal(arg)
		if err != nil {
			return "", fmt.Errorf("can't marshal argument: %w", err)
		}

		raws = append(raws, r)
	}

	j, err := json.Marshal(raws)
	if err != nil {
		return "", fmt.Errorf("can't marshal marshalled arguments: %w", err)
	}

	return string(j), nil
}

func marshalKprobeReturn(ret *tetragon.KprobeArgument) (string, error) {
	mo := marshalOptions()

	marshalled, err := mo.Marshal(ret)
	if err != nil {
		return "", fmt.Errorf("can't marshal return value: %w", err)
	}

	return string(marshalled), nil
}

func uint32WrapperToPtr(v *wrapperspb.UInt32Value) *uint32 {
	if v != nil {
		return &v.Value
	}

	return nil
}

func capsToStrings(caps []tetragon.CapabilitiesType) []string {
	ss := make([]string, 0, len(caps))

	for _, cap := range caps {
		ss = append(ss, cap.String())
	}

	return ss
}

func marshalOptions() protojson.MarshalOptions {
	return protojson.MarshalOptions{UseProtoNames: true} // enable snake_case for field names
}
