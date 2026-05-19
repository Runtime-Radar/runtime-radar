package convert

import (
	"errors"
	"fmt"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/google/uuid"
	processor_api "github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/model"
	enf_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

func DetectorCountersToProto(cs []*model.DetectorCounter) *api.DetectorRatingResp {
	proto := api.DetectorRatingResp{Counters: make([]*api.DetectorRatingResp_Counter, 0, len(cs))}

	for _, c := range cs {
		proto.Counters = append(proto.Counters, &api.DetectorRatingResp_Counter{
			DetectorId: c.DetectorID,
			Severity:   c.Severity.String(),
			Count:      int32(c.Count),
		})
	}

	return &proto
}

func RuntimeEventFromProto(proto *processor_api.RuntimeEvent) (model.RuntimeEvent, error) {
	id, err := uuid.Parse(proto.GetId())
	if err != nil {
		return model.RuntimeEvent{}, fmt.Errorf("can't parse id: %w", err)
	}

	protoEvent := proto.GetEvent()

	incidentSeverity := enf_model.NoneSeverity
	incidentSeverity.Set(proto.GetIncidentSeverity())

	event := model.RuntimeEvent{
		ID:               id,
		NodeName:         protoEvent.GetNodeName(),
		TetragonVersion:  proto.GetTetragonVersion(),
		SourceEvent:      (*model.RuntimeEventJSON)(protoEvent),
		RegisteredAt:     protoEvent.GetTime().AsTime(),
		IsIncident:       proto.GetIsIncident(),
		IncidentSeverity: incidentSeverity,
		BlockBy:          proto.GetBlockBy(),
		NotifyBy:         proto.GetNotifyBy(),
		DetectErrors:     proto.GetDetectErrors(),
	}

	if ts := proto.GetThreats(); len(ts) > 0 {
		event.Threats = ts

		event.ThreatsSeverities = make([]enf_model.Severity, 0, len(ts))
		event.ThreatsDetectors = make([]string, 0, len(ts))

		for _, t := range ts {
			var sev enf_model.Severity
			sev.Set(t.GetSeverity())

			event.ThreatsSeverities = append(event.ThreatsSeverities, sev)
			event.ThreatsDetectors = append(event.ThreatsDetectors, t.GetDetector().GetId())
		}
	}

	switch e := protoEvent.GetEvent().(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		event.EventType = model.RuntimeEventTypeProcessExec
		if err := setRuntimeEventProcessFields(e.ProcessExec.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		if err := setRuntimeEventParentFields(e.ProcessExec.GetParent(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
	case *tetragon.GetEventsResponse_ProcessExit:
		event.EventType = model.RuntimeEventTypeProcessExit
		if err := setRuntimeEventProcessFields(e.ProcessExit.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		if err := setRuntimeEventParentFields(e.ProcessExit.GetParent(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		event.ExitSignal = e.ProcessExit.GetSignal()
		event.ExitStatus = e.ProcessExit.GetStatus()
		t := e.ProcessExit.GetTime().AsTime()
		event.ExitTime = &t
	case *tetragon.GetEventsResponse_ProcessKprobe:
		event.EventType = model.RuntimeEventTypeProcessKprobe
		if err := setRuntimeEventProcessFields(e.ProcessKprobe.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		if err := setRuntimeEventParentFields(e.ProcessKprobe.GetParent(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		event.KprobeFunctionName = e.ProcessKprobe.GetFunctionName()
		event.KprobeAction = e.ProcessKprobe.GetAction().String()
		event.PolicyName = e.ProcessKprobe.GetPolicyName()
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		event.EventType = model.RuntimeEventTypeProcessTracepoint
		if err := setRuntimeEventProcessFields(e.ProcessTracepoint.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		if err := setRuntimeEventParentFields(e.ProcessTracepoint.GetParent(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		event.TracepointSubsys = e.ProcessTracepoint.GetSubsys()
		event.TracepointEvent = e.ProcessTracepoint.GetEvent()
		event.PolicyName = e.ProcessTracepoint.GetPolicyName()
	case *tetragon.GetEventsResponse_ProcessLoader:
		event.EventType = model.RuntimeEventTypeProcessLoader
		if err := setRuntimeEventProcessFields(e.ProcessLoader.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		event.LoaderPath = e.ProcessLoader.GetPath()
		event.LoaderBuildid = e.ProcessLoader.GetBuildid()
	case *tetragon.GetEventsResponse_ProcessUprobe:
		event.EventType = model.RuntimeEventTypeProcessUprobe
		if err := setRuntimeEventProcessFields(e.ProcessUprobe.GetProcess(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		if err := setRuntimeEventParentFields(e.ProcessUprobe.GetParent(), &event); err != nil {
			return model.RuntimeEvent{}, err
		}
		event.UprobeSymbol = e.ProcessUprobe.GetSymbol()
		event.UprobePath = e.ProcessUprobe.GetPath()
		event.PolicyName = e.ProcessUprobe.GetPolicyName()
	default:
		return model.RuntimeEvent{}, fmt.Errorf("unknown event type: %T", e)
	}

	return event, nil
}

func RuntimeEventToProto(event *model.RuntimeEvent) (*processor_api.RuntimeEvent, error) {
	proto := &processor_api.RuntimeEvent{
		Id:               event.ID.String(),
		TetragonVersion:  event.TetragonVersion,
		Event:            (*tetragon.GetEventsResponse)(event.SourceEvent),
		Threats:          event.Threats,
		IsIncident:       event.IsIncident,
		IncidentSeverity: event.IncidentSeverity.String(),
		BlockBy:          event.BlockBy,
		NotifyBy:         event.NotifyBy,
		DetectErrors:     event.DetectErrors,
	}

	return proto, nil
}

func RuntimeEventsToProto(events []*model.RuntimeEvent) ([]*processor_api.RuntimeEvent, error) {
	res := make([]*processor_api.RuntimeEvent, 0, len(events))
	for _, event := range events {
		re, err := RuntimeEventToProto(event)
		if err != nil {
			return nil, err
		}
		res = append(res, re)
	}

	return res, nil
}

func setRuntimeEventProcessFields(p *tetragon.Process, re *model.RuntimeEvent) error {
	if re == nil {
		return errors.New("can't assign attributes to nil")
	}

	re.ProcessExecID = p.GetExecId()
	re.ProcessPid = p.GetPid().GetValue()
	re.ProcessUID = p.GetUid().GetValue()
	re.ProcessCwd = p.GetCwd()
	re.ProcessBinary = p.GetBinary()
	re.ProcessArguments = p.GetArguments()
	re.ProcessFlags = p.GetFlags()
	t := p.GetStartTime().AsTime()
	re.ProcessStartTime = &t
	re.ProcessAuid = p.GetAuid().GetValue()

	re.ProcessPodNamespace = p.GetPod().GetNamespace()
	re.ProcessPodName = p.GetPod().GetName()

	re.ProcessPodContainerID = p.GetPod().GetContainer().GetId()
	re.ProcessPodContainerName = p.GetPod().GetContainer().GetName()
	re.ProcessPodContainerImageID = p.GetPod().GetContainer().GetImage().GetId()
	re.ProcessPodContainerImageName = p.GetPod().GetContainer().GetImage().GetName()
	processPodContainerStartTime := p.GetPod().GetContainer().GetStartTime().AsTime()
	re.ProcessPodContainerStartTime = &processPodContainerStartTime
	re.ProcessPodContainerPid = p.GetPod().GetContainer().GetPid().GetValue()
	re.ProcessPodContainerMaybeExecProbe = p.GetPod().GetContainer().GetMaybeExecProbe()

	if pls := p.GetPod().GetPodLabels(); len(pls) > 0 {
		re.ProcessPodPodLabels = model.PodLabels(pls)
	}
	re.ProcessPodWorkload = p.GetPod().GetWorkload()
	re.ProcessPodWorkloadKind = p.GetPod().GetWorkloadKind()

	re.ProcessDocker = p.GetDocker()
	re.ProcessParentExecID = p.GetParentExecId()
	re.ProcessRefcnt = p.GetRefcnt()

	re.ProcessCapPermitted = p.GetCap().GetPermitted()
	re.ProcessCapEffective = p.GetCap().GetEffective()
	re.ProcessCapInheritable = p.GetCap().GetInheritable()

	re.ProcessNsUtsInum = p.GetNs().GetUts().GetInum()
	re.ProcessNsUtsIsHost = p.GetNs().GetUts().GetIsHost()

	re.ProcessNsIpcInum = p.GetNs().GetIpc().GetInum()
	re.ProcessNsIpcIsHost = p.GetNs().GetIpc().GetIsHost()

	re.ProcessNsMntInum = p.GetNs().GetMnt().GetInum()
	re.ProcessNsMntIsHost = p.GetNs().GetMnt().GetIsHost()

	re.ProcessNsPidInum = p.GetNs().GetPid().GetInum()
	re.ProcessNsPidIsHost = p.GetNs().GetPid().GetIsHost()

	re.ProcessNsPidForChildrenInum = p.GetNs().GetPidForChildren().GetInum()
	re.ProcessNsPidForChildrenIsHost = p.GetNs().GetPidForChildren().GetIsHost()

	re.ProcessNsNetInum = p.GetNs().GetNet().GetInum()
	re.ProcessNsNetIsHost = p.GetNs().GetNet().GetIsHost()

	re.ProcessNsTimeInum = p.GetNs().GetTime().GetInum()
	re.ProcessNsTimeIsHost = p.GetNs().GetTime().GetIsHost()

	re.ProcessNsTimeForChildrenInum = p.GetNs().GetTimeForChildren().GetInum()
	re.ProcessNsTimeForChildrenIsHost = p.GetNs().GetTimeForChildren().GetIsHost()

	re.ProcessNsCgroupInum = p.GetNs().GetCgroup().GetInum()
	re.ProcessNsCgroupIsHost = p.GetNs().GetCgroup().GetIsHost()

	re.ProcessNsUserInum = p.GetNs().GetUser().GetInum()
	re.ProcessNsUserIsHost = p.GetNs().GetUser().GetIsHost()
	re.ProcessTid = p.GetTid().GetValue()

	return nil
}

func setRuntimeEventParentFields(p *tetragon.Process, re *model.RuntimeEvent) error {
	if re == nil {
		return errors.New("can't assign attributes to nil")
	}

	re.ParentExecID = p.GetExecId()
	re.ParentPid = p.GetPid().GetValue()
	re.ParentUID = p.GetUid().GetValue()
	re.ParentCwd = p.GetCwd()
	re.ParentBinary = p.GetBinary()
	re.ParentArguments = p.GetArguments()
	re.ParentFlags = p.GetFlags()
	t := p.GetStartTime().AsTime()
	re.ParentStartTime = &t
	re.ParentAuid = p.GetAuid().GetValue()

	re.ParentPodNamespace = p.GetPod().GetNamespace()
	re.ParentPodName = p.GetPod().GetName()

	re.ParentPodContainerID = p.GetPod().GetContainer().GetId()
	re.ParentPodContainerName = p.GetPod().GetContainer().GetName()
	re.ParentPodContainerImageID = p.GetPod().GetContainer().GetImage().GetId()
	re.ParentPodContainerImageName = p.GetPod().GetContainer().GetImage().GetName()
	parentPodContainerStartTime := p.GetPod().GetContainer().GetStartTime().AsTime()
	re.ParentPodContainerStartTime = &parentPodContainerStartTime
	re.ParentPodContainerPid = p.GetPod().GetContainer().GetPid().GetValue()
	re.ParentPodContainerMaybeExecProbe = p.GetPod().GetContainer().GetMaybeExecProbe()

	if pls := p.GetPod().GetPodLabels(); len(pls) > 0 {
		re.ProcessPodPodLabels = model.PodLabels(pls)
	}
	re.ParentPodWorkload = p.GetPod().GetWorkload()
	re.ParentPodWorkloadKind = p.GetPod().GetWorkloadKind()

	re.ParentDocker = p.GetDocker()
	re.ParentParentExecID = p.GetParentExecId()
	re.ParentRefcnt = p.GetRefcnt()

	re.ParentCapPermitted = p.GetCap().GetPermitted()
	re.ParentCapEffective = p.GetCap().GetEffective()
	re.ParentCapInheritable = p.GetCap().GetInheritable()

	re.ParentNsUtsInum = p.GetNs().GetUts().GetInum()
	re.ParentNsUtsIsHost = p.GetNs().GetUts().GetIsHost()

	re.ParentNsIpcInum = p.GetNs().GetIpc().GetInum()
	re.ParentNsIpcIsHost = p.GetNs().GetIpc().GetIsHost()

	re.ParentNsMntInum = p.GetNs().GetMnt().GetInum()
	re.ParentNsMntIsHost = p.GetNs().GetMnt().GetIsHost()

	re.ParentNsPidInum = p.GetNs().GetPid().GetInum()
	re.ParentNsPidIsHost = p.GetNs().GetPid().GetIsHost()

	re.ParentNsPidForChildrenInum = p.GetNs().GetPidForChildren().GetInum()
	re.ParentNsPidForChildrenIsHost = p.GetNs().GetPidForChildren().GetIsHost()

	re.ParentNsNetInum = p.GetNs().GetNet().GetInum()
	re.ParentNsNetIsHost = p.GetNs().GetNet().GetIsHost()

	re.ParentNsTimeInum = p.GetNs().GetTime().GetInum()
	re.ParentNsTimeIsHost = p.GetNs().GetTime().GetIsHost()

	re.ParentNsTimeForChildrenInum = p.GetNs().GetTimeForChildren().GetInum()
	re.ParentNsTimeForChildrenIsHost = p.GetNs().GetTimeForChildren().GetIsHost()

	re.ParentNsCgroupInum = p.GetNs().GetCgroup().GetInum()
	re.ParentNsCgroupIsHost = p.GetNs().GetCgroup().GetIsHost()

	re.ParentNsUserInum = p.GetNs().GetUser().GetInum()
	re.ParentNsUserIsHost = p.GetNs().GetUser().GetIsHost()

	re.ParentTid = p.GetTid().GetValue()

	return nil
}
