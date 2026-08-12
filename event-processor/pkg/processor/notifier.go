//go:build !tinygo.wasm

package processor

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/api"
	notifier_api "github.com/runtime-radar/runtime-radar/notifier/api"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	enforcer_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (wp *WorkersPool) notify(ctx context.Context, ed *eventData, threats []*api.Threat, rules []*enforcer_api.Rule, block bool, eventID string) error {
	msgs := []*notifier_api.Message{}

	sev := enforcer_model.NoneSeverity
	nts := make([]*notifier_api.RuntimeEvent_Event_Threat, 0, len(threats))

	for _, t := range threats {
		tSev := enforcer_model.NoneSeverity
		tSev.Set(t.GetSeverity())
		if tSev > sev {
			sev = tSev
		}

		tacticsCovered := make([]*notifier_api.RuntimeEvent_Event_MitreTactic, 0, len(t.GetTacticsCovered()))
		for _, mt := range t.GetTacticsCovered() {
			tacticsCovered = append(tacticsCovered, &notifier_api.RuntimeEvent_Event_MitreTactic{
				Id:         mt.GetId(),
				Techniques: mt.GetTechniques(),
			})
		}

		nts = append(nts, &notifier_api.RuntimeEvent_Event_Threat{
			DetectorId:          t.GetDetector().GetId(),
			DetectorName:        t.GetDetector().GetName(),
			DetectorVersion:     t.GetDetector().GetVersion(),
			DetectorDescription: t.GetDetector().GetDescription(),
			Severity:            t.GetSeverity(),
			Reason:              t.GetReason(),
			TacticsCovered:      tacticsCovered,
		})

	}

	innerEvent := &notifier_api.RuntimeEvent_Event{
		Threats:             nts,
		EventType:           ed.EventType,
		PodNamespace:        ed.PodNamespace,
		PodName:             ed.PodName,
		ContainerId:         ed.ContainerID,
		ContainerName:       ed.ContainerName,
		ContainerImage:      ed.ContainerImage,
		FunctionName:        ed.FunctionName,
		FunctionArgs:        ed.FunctionArgs,
		FunctionReturn:      ed.FunctionReturn,
		ProcessPid:          ed.ProcessPID,
		ProcessUid:          ed.ProcessUID,
		ProcessBinary:       ed.ProcessBinary,
		ProcessArguments:    ed.ProcessArguments,
		ProcessCapEffective: ed.ProcessCapEffective,
		ProcessCapPermitted: ed.ProcessCapPerimtted,
		ProcessSetuid:       ed.ProcessSetuid,
		ProcessSetgid:       ed.ProcessSetgid,
		ParentPid:           ed.ParentPID,
		ParentUid:           ed.ParentUID,
		ParentBinary:        ed.ParentBinary,
		ParentArguments:     ed.ParentArguments,
		NodeName:            ed.NodeName,
	}

	for _, r := range rules {
		for _, id := range r.GetRule().GetNotify().GetTargets() {
			ev := &notifier_api.RuntimeEvent{
				Event:        innerEvent,
				Severity:     sev.String(),
				RegisteredAt: timestamppb.New(ed.RegisteredAt),
				Block:        block,
				RuleName:     r.GetName(),
				EventId:      eventID,
			}

			msgs = append(msgs, &notifier_api.Message{
				Event:          &notifier_api.Message_RuntimeEvent{RuntimeEvent: ev},
				NotificationId: id,
			})
		}
	}

	_, err := wp.notifier.Notify(ctx, &notifier_api.NotifyReq{Notifications: msgs})
	return err
}
