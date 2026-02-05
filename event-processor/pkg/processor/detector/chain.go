package detector

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tetragon_api "github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	enf_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

// Chain is a chain of detectors wrapped with corresponding info.
type Chain map[Key]Wrapper

// Key is used to identify particular detector. We do not allow same combination of ID and version, but we do allow same ID with different versions.
// This is in line with model.Detector, where primary key is also set to ID and version.
type Key struct {
	ID      string
	Version uint
}

// Wrapper represents a detector itself with info about it alongside, so that it can be used without calling Detector.Info.
type Wrapper struct {
	detector_api.Detector
	Info *api.Detector
}

// ChainResult represents a result of execution of detectors' chain.
// If at least one of detectors returned an error, it's appended to Errors
// and it doesn't prevent remaining chain from being executed.
type ChainResult struct {
	Threats []*api.Threat
	Errors  []*api.DetectError
}

// Detect executes chain of detectors and returns a result, which can contain found threats or detect errors.
// Error in particular detector won't prevent other detectors in chain from being executed.
// Detect will stop with error if it occurs anywhere else. If len(c) == 0, it will return nil, nil.
func (c Chain) Detect(ctx context.Context, event *tetragon_api.GetEventsResponse) (*ChainResult, error) {
	if len(c) == 0 {
		return nil, nil
	}

	res := &ChainResult{}

	dEvent, err := convertEvent(event)
	if err != nil {
		return nil, fmt.Errorf("can't convert event: %w", err)
	}

	for _, d := range c {
		t0 := time.Now()
		resp, err := d.Detect(ctx, &detector_api.DetectReq{Event: dEvent})
		if err != nil {
			res.Errors = append(res.Errors, &api.DetectError{
				Detector: d.Info,
				Error:    err.Error(),
			})

			continue
		}
		delta := time.Since(t0)

		log.Debug().Str("delay", delta.String()).Interface("detector_event", dEvent).Interface("detector_resp", resp).Msgf("Detector[%s] is done", d.Info.GetId())

		if resp.GetSeverity() > detector_api.DetectResp_NONE {
			s := enf_model.Severity(resp.GetSeverity())

			res.Threats = append(res.Threats, &api.Threat{
				Detector: &api.Detector{
					Id:          d.Info.GetId(),
					Name:        d.Info.GetName(),
					Description: d.Info.GetDescription(),
					Version:     d.Info.GetVersion(),
					Author:      d.Info.GetAuthor(),
					Contact:     d.Info.GetContact(),
				},
				Severity: s.String(),
			})
		}
	}

	return res, nil
}

// String implements fmt.Stringer for logging purposes.
func (c Chain) String() string {
	wb := &strings.Builder{}

	i := 0
	for k := range c {
		wb.WriteString("[")
		wb.WriteString(k.ID + ";")
		wb.WriteString(strconv.Itoa(int(k.Version)))
		wb.WriteString("]")
		if i < len(c)-1 {
			wb.WriteString(", ")
		}
		i++
	}

	return wb.String()
}
