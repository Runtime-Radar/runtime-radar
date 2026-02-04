package detector

import (
	"context"
	"fmt"
	"maps"

	"github.com/cilium/tetragon/pkg/arch"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
)

// Matcher is a matrix which links events and detectors (event type -> function -> chain). It makes it possible to dynamically construct a [Chain] via [Matcher.MatchChain] method.
type Matcher map[string]map[string]Chain

func NewMatcher(ctx context.Context, ds ...detector_api.Detector) (Matcher, error) {
	m := Matcher{}

	for _, d := range ds {
		infoResp, err := d.Info(ctx, &detector_api.InfoReq{})
		if err != nil {
			return nil, fmt.Errorf("can't get detector info: %w", err)
		}

		key := Key{infoResp.GetId(), uint(infoResp.GetVersion())}

		info := &api.Detector{
			Id:          infoResp.GetId(),
			Name:        infoResp.GetName(),
			Description: infoResp.GetDescription(),
			Version:     infoResp.GetVersion(),
			Author:      infoResp.GetAuthor(),
			Contact:     infoResp.GetContact(),
			License:     infoResp.GetLicense(),
		}

		w := Wrapper{
			Detector: d,
			Info:     info,
		}

		tcResp, err := d.TriggerCriteria(ctx, &detector_api.TriggerCriteriaReq{})
		if err != nil {
			return nil, fmt.Errorf("can't get detector trigger criteria: %w", err)
		}

		if len(tcResp.GetCriteria()) == 0 {
			return nil, fmt.Errorf("empty trigger criteria for detector: %+v", key)
		}

		for k, v := range tcResp.GetCriteria() {
			eventType := k
			funcNames := v.GetFuncNames()

			if m[eventType] == nil {
				m[eventType] = map[string]Chain{}
			}

			if len(funcNames) == 0 {
				t := m[eventType]["*"]
				if t == nil {
					m[eventType]["*"] = Chain{key: w}
				} else {
					maps.Copy(t, Chain{key: w})
				}
			} else {
				for _, fn := range funcNames {
					_, fn = arch.CutSyscallPrefix(fn)
					t := m[eventType][fn]
					if t == nil {
						m[eventType][fn] = Chain{key: w}
					} else {
						maps.Copy(t, Chain{key: w})
					}
				}
			}
		}
	}

	return m, nil
}

func (m Matcher) MatchChain(eventType string, funcNames ...string) Chain {
	// Always match "*" pattern
	funcNames = append(funcNames, "*")

	var chain Chain

	for _, fn := range funcNames {
		matched, ok := m[eventType][fn]
		if ok && chain == nil {
			chain = maps.Clone(matched)
		} else if ok {
			maps.Copy(chain, matched)
		}
	}

	return chain
}
