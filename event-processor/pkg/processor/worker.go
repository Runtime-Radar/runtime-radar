//go:build !tinygo.wasm

package processor

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor/detector"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	enforcer_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

const (
	// TODO: pass Tetragon version from runtime_monitor when corresponding field is added to incoming message
	tetragonVersion = "v1.5.0"
	actionType      = "runtime-threat-detection"
)

func (wp *WorkersPool) worker(id int) {
	defer wp.wg.Done()

	log.Info().Msgf("Worker[%d] started", id)
	defer log.Info().Msgf("Worker[%d] stopped", id)

	ctx := context.Background()

	bins, rootHash := wp.Bins()
	t0 := time.Now()

	matcher, err := getMatcher(ctx, wp.plugin, bins)
	if err != nil {
		// If very first initialization failed, there are not too many options
		panic(fmt.Errorf("can't init matcher when starting worker[%d]: %w", id, err))
	}
	log.Info().Str("delay", time.Since(t0).String()).Str("root_hash", rootHash).Msgf("Matcher initialized for worker[%d]", id)

	upd := make(chan bool, 1)
	wp.registerUpdate(upd)

	for {
		select {
		case <-upd:
			bins, rootHash := wp.Bins()
			t0 := time.Now()
			matcher, err = getMatcher(ctx, wp.plugin, bins)
			if err != nil {
				log.Error().Err(err).Msgf("Can't init matcher chain for worker[%d]", id)
			} else {
				log.Info().Str("delay", time.Since(t0).String()).Str("root_hash", rootHash).Msgf("Matcher initialized for worker[%d]", id)
			}
		case j := <-wp.jobs:
			log.Debug().Interface("job", j).Int("id", id).Msgf("Worker[%d] got job", id)

			t0 := time.Now()
			result, err := wp.doJob(ctx, j, matcher) // <-- do the job
			delta := time.Since(t0)

			if wp.withReports {
				t := time.NewTimer(reportTimeout)
				select {
				case wp.reports <- &Report{
					ID:     id,
					Result: result,
					Err:    err,
					Delay:  delta,
				}:
				case <-t.C:
					log.Error().Msgf("Timeout dispatching report")
				}
				t.Stop()
			}

			if err != nil {
				log.Error().Err(err).Str("delay", delta.String()).Interface("event", j).Interface("result", result).Msgf("Worker[%d] got error", id)
			} else {
				log.Debug().Str("delay", delta.String()).Interface("event", j).Interface("result", result).Msgf("Worker[%d] done job", id)
			}
		case <-wp.fire:
			return // <-- return
		}
	}
}

func (wp *WorkersPool) doJob(ctx context.Context, event *tetragon.GetEventsResponse, matcher detector.Matcher) (*detector.ChainResult, error) {
	eventType, funcName := getEventAndFunc(event)
	chain := matcher.MatchChain(eventType, funcName)

	log.Debug().Str("event_type", eventType).Str("func_name", funcName).Str("chain", chain.String()).Msgf("Got detectors chain")
	metrics.DetectorsChainLength.Observe(float64(len(chain)))

	t0 := time.Now()
	result, err := chain.Detect(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("can't detect event: %w", err)
	}
	delta := time.Since(t0)

	log.Debug().Str("delay", delta.String()).Interface("event", event).Interface("result", result).Msgf("Detectors chain is done")
	metrics.OperationTimeAllDetectorsHistogram.Observe(delta.Seconds())

	eventData, err := getEventData(event)
	if err != nil {
		return nil, fmt.Errorf("can't get event data: %w", err)
	}

	blockRules, notifyRules := []*enforcer_api.Rule{}, []*enforcer_api.Rule{}
	incidentSeverity := enforcer_model.NoneSeverity // incident's severity to be passed to history API. Depends on policy enforce's response

	var threats []*api.Threat
	var detectErrors []*api.DetectError

	if result != nil {
		threats = result.Threats
		detectErrors = result.Errors
	}

	if len(threats) > 0 {
		// TODO: temporarily enabled logging of threats to INFO level, need to take a look if it adds any valuable overhead in production, and in case it does,
		// it will be a subject for removal (there is already a lot of DEBUG messages at the moment, no need to add another one)
		log.Info().Str("delay", delta.String()).Interface("event_data", eventData).Int("chain_length", len(chain)).Interface("result", result).Msg("Threats detected")

		enforcerResp, err := wp.evaluatePolicy(ctx, eventData, threats)
		if err != nil {
			return nil, fmt.Errorf("can't evaluate policy: %w", err)
		}

		for _, e := range enforcerResp.GetResult().GetEvents() {
			p := e.GetPolicy()

			// take event's severity into account only if at least one rule matched
			if len(p.GetBlockBy()) > 0 || len(p.GetNotifyBy()) > 0 {
				s := enforcer_model.NoneSeverity
				s.Set(e.GetSeverity())

				if s > incidentSeverity {
					incidentSeverity = s
				}
			}

			for _, r := range p.GetBlockBy() {
				blockRules = append(blockRules, r)
			}

			for _, r := range p.GetNotifyBy() {
				notifyRules = append(notifyRules, r)
			}
		}
	} else if len(detectErrors) > 0 {
		log.Info().Str("delay", delta.String()).Interface("event_data", eventData).Interface("result", result).Msg("Detect errors found")
	}

	var eventID string
	if cfg := wp.Config(); shouldSaveEvent(cfg.Config.HistoryControl, threats) {
		eventID = uuid.NewString()
		re := &api.RuntimeEvent{
			Id:               eventID,
			TetragonVersion:  tetragonVersion,
			Event:            event,
			Threats:          threats,
			DetectErrors:     detectErrors,
			IsIncident:       len(blockRules) > 0 || len(notifyRules) > 0,
			IncidentSeverity: incidentSeverity.String(),
			BlockBy:          uniqueRuleIDs(blockRules),
			NotifyBy:         uniqueRuleIDs(notifyRules),
		}

		if err := wp.history.Publish(ctx, re); err != nil {
			metrics.RabbitAddHistoryFailureCounter.Inc()

			return nil, fmt.Errorf("can't publish runtime event '%+v'", re)
		}

		metrics.RabbitAddHistorySuccessCounter.Inc()
	}

	block := false
	if len(blockRules) > 0 {
		block = true
		// TODO: implement blocking/isolation of pod as a response
	}

	if len(notifyRules) > 0 {
		if err := wp.notify(ctx, eventData, result.Threats, uniqueRules(notifyRules), block, eventID); err != nil {
			return nil, fmt.Errorf("can't notify: %w", err)
		}
	}

	return result, nil
}

func getMatcher(ctx context.Context, plugin *detector_api.DetectorPlugin, bins [][]byte) (detector.Matcher, error) {
	ds, err := detector.BinsToDetectors(ctx, plugin, bins)
	if err != nil {
		return nil, fmt.Errorf("can't get detectors from bins: %w", err)
	}

	m, err := detector.NewMatcher(ctx, ds...)
	if err != nil {
		return nil, fmt.Errorf("can't init matcher: %w", err)
	}

	return m, nil
}

func uniqueRules(rs []*enforcer_api.Rule) []*enforcer_api.Rule {
	m := make(map[string]*enforcer_api.Rule, len(rs))

	for _, r := range rs {
		m[r.GetId()] = r
	}

	return slices.Collect(maps.Values(m))
}

func uniqueRuleIDs(rs []*enforcer_api.Rule) []string {
	res := []string{}
	urs := uniqueRules(rs)

	for _, r := range urs {
		res = append(res, r.GetId())
	}

	return res
}

func shouldSaveEvent(hc api.Config_ConfigJSON_HistoryControl, ts []*api.Threat) bool {
	switch hc {
	case api.Config_ConfigJSON_NONE:
		return false
	case api.Config_ConfigJSON_ALL:
		return true
	case api.Config_ConfigJSON_WITH_THREATS:
		return len(ts) > 0
	default: // normally should not happen
		panic(fmt.Sprintf("invalid historyControl value: %v", hc))
	}
}
