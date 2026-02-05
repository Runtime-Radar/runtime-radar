//go:build !tinygo.wasm

package service

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/runtime-radar/runtime-radar/event-processor/api"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/database"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model/convert"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/processor/detector"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DetectorGeneric struct {
	api.UnimplementedDetectorControllerServer

	Processor          processor.Processor
	DetectorRepository database.DetectorRepository
	DetectorPlugin     *detector_api.DetectorPlugin
}

func (dg *DetectorGeneric) Create(ctx context.Context, req *api.CreateDetectorReq) (*api.CreateDetectorResp, error) {
	if reason, ok := dg.validateCreateDetectorReq(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	wasmBase64 := req.GetWasmBase64()
	wasmBinary, err := base64.StdEncoding.DecodeString(wasmBase64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse base64 binary: %v", err)
	}

	d, err := detector.BinToModel(ctx, dg.DetectorPlugin, wasmBinary)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get detector info from binary: %v", err)
	}

	if err := dg.DetectorRepository.Add(ctx, d); err != nil {
		return nil, status.Errorf(codes.Internal, "can't add detector: %v", err)
	}

	if err := dg.updateDetectors(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "can't update processor: %v", err)
	}

	resp := &api.CreateDetectorResp{
		Detector: convert.DetectorToProto(d),
	}

	return resp, nil
}

func (dg *DetectorGeneric) Delete(ctx context.Context, req *api.DeleteDetectorReq) (*emptypb.Empty, error) {
	if reason, ok := dg.validateDeleteDetectorReq(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	if err := dg.DetectorRepository.Delete(ctx, req.GetId(), uint(req.GetVersion())); err != nil {
		return nil, status.Errorf(codes.Internal, "can't delete detector: %v", err)
	}

	if err := dg.updateDetectors(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "can't update processor: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (dg *DetectorGeneric) ListPage(ctx context.Context, req *api.ListDetectorPageReq) (*api.ListDetectorPageResp, error) {
	total, err := dg.DetectorRepository.GetCount(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get detector count: %v", err)
	}

	order, pageSize := req.GetOrder(), req.GetPageSize()
	if order == "" {
		order = defaultOrder
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	ds, err := dg.DetectorRepository.GetPage(ctx, nil, order, int(pageSize), int(req.GetPageNum()), true) // preload is on
	if err != nil {
		if errors.Is(err, database.ErrInvalidOrder) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "can't get detector page: %v", err)
	}

	resp := &api.ListDetectorPageResp{
		Total:     uint32(total),
		Detectors: convert.DetectorsToProto(ds),
	}

	return resp, nil
}

func (dg *DetectorGeneric) updateDetectors(ctx context.Context) error {
	bins, err := dg.DetectorRepository.GetAllBins(ctx, nil)
	if err != nil {
		return err
	}

	dg.Processor.UpdateDetectors(bins)

	return nil
}

func (dg *DetectorGeneric) validateCreateDetectorReq(req *api.CreateDetectorReq) (reason string, ok bool) {
	if req.GetWasmBase64() == "" {
		return "empty or missing base64 encoded wasm binary", false
	}

	return "", true
}

func (dg *DetectorGeneric) validateDeleteDetectorReq(req *api.DeleteDetectorReq) (reason string, ok bool) {
	// Not checking version because it can theoretically be set to zero, and we do not want to enforce it to always be > 0
	if req.GetId() == "" {
		return "empty or missing id", false
	}

	return "", true
}
