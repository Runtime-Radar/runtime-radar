//go:build !tinygo.wasm

package detector

import (
	"context"
	"fmt"

	tetragon_api "github.com/cilium/tetragon/api/v1/tetragon"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	detector_tetragon_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/tetratelabs/wazero"
	"google.golang.org/protobuf/proto"
)

func NewPlugin(ctx context.Context) (*detector_api.DetectorPlugin, error) {
	mc := wazero.NewModuleConfig().WithStartFunctions("_initialize")

	p, err := detector_api.NewDetectorPlugin(ctx, detector_api.WazeroModuleConfig(mc))
	if err != nil {
		return nil, fmt.Errorf("can't instantiate plugin: %w", err)
	}

	return p, nil
}

func BinsToDetectors(ctx context.Context, plugin *detector_api.DetectorPlugin, bins [][]byte) ([]detector_api.Detector, error) {
	ds := make([]detector_api.Detector, 0, len(bins))

	for _, bin := range bins {
		d, err := plugin.LoadBinary(ctx, bin)
		if err != nil {
			return nil, fmt.Errorf("can't load binary: %w", err)
		}
		ds = append(ds, d)
	}

	return ds, nil
}

func BinToModel(ctx context.Context, plugin *detector_api.DetectorPlugin, bin []byte) (*model.Detector, error) {
	detector, err := plugin.LoadBinary(ctx, bin)
	if err != nil {
		return nil, fmt.Errorf("can't load wasm module: %w", err)
	}

	infoResp, err := detector.Info(ctx, &detector_api.InfoReq{})
	if err != nil {
		return nil, fmt.Errorf("can't get detector info: %w", err)
	}

	d := &model.Detector{
		ID:             infoResp.GetId(),
		Name:           infoResp.GetName(),
		Description:    infoResp.GetDescription(),
		Version:        uint(infoResp.GetVersion()),
		Author:         infoResp.GetAuthor(),
		Contact:        infoResp.GetContact(),
		License:        infoResp.GetLicense(),
		WasmBinary:     bin,
		WasmHash:       security.HashSHA512AsHex(bin),
		TacticsCovered: infoResp.GetTacticsCovered(),
	}

	return d, err
}

func convertEvent(in *tetragon_api.GetEventsResponse) (*detector_tetragon_api.GetEventsResponse, error) {
	b, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	out := &detector_tetragon_api.GetEventsResponse{}
	if err := out.UnmarshalVT(b); err != nil {
		return nil, err
	}

	return out, nil
}
