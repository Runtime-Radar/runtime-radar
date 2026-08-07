package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/updater"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type ConfigGeneric struct {
	api.UnimplementedConfigControllerServer

	database.ConfigRepository
	updater.ConfigUpdater
}

func (cg *ConfigGeneric) Add(ctx context.Context, req *api.Config) (*emptypb.Empty, error) {
	if reason, ok := cg.validateConfig(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	idStr := req.GetId()
	var id uuid.UUID
	var err error

	if idStr != "" {
		id, err = uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
		}
	}

	cfg := &model.Config{
		Base:   model.Base{ID: id},
		Config: (*model.ConfigJSON)(req.GetConfig()),
	}

	if cfg.Config == nil {
		cfg = model.DefaultConfig
	}
	oldCfg := cg.ConfigUpdater.Config()

	log.Debug().Interface("old_config", oldCfg).Msgf("Old kube manager config")
	log.Debug().Interface("new_config", cfg).Msgf("New kube manager config")

	if cg.ConfigUpdater.Equal(oldCfg, cfg) {
		log.Debug().Msgf("Monitor config didn't change")

		return &emptypb.Empty{}, nil
	}

	if err := cg.ConfigRepository.Add(ctx, cfg); err != nil {
		return nil, status.Errorf(codes.Internal, "can't add config: %v", err)
	}

	cg.ConfigUpdater.SetConfig(cfg)

	return &emptypb.Empty{}, nil
}

func (cg *ConfigGeneric) Read(ctx context.Context, _ *emptypb.Empty) (*api.Config, error) {
	cfg, err := cg.ConfigRepository.GetLast(ctx, false)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, status.Errorf(codes.NotFound, "config not found")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "can't read config: %v", err)
	}

	resp := &api.Config{
		Id:     cfg.ID.String(),
		Config: (*api.Config_ConfigJSON)(cfg.Config),
	}

	return resp, nil
}

func (cg *ConfigGeneric) validateConfig(req *api.Config) (string, bool) {
	if req.Config == nil {
		return "no config", false
	}

	if req.Config.GetVersion() == "" {
		return "empty or missing config version", false
	} else if ver := req.Config.GetVersion(); ver != string(model.ConfigVersion) {
		return fmt.Sprintf("config version mismatch: expected %s, got %s", model.ConfigVersion, ver), false
	}

	if req.Config.DenyList == nil || req.Config.AllowList == nil {
		return "", true
	}

	m := make(map[string]interface{})
	for _, d := range req.Config.DenyList.Namespaces {
		m[d] = struct{}{}
	}

	for _, a := range req.Config.AllowList.Namespaces {
		if _, ok := m[a]; ok {
			return "allow and deny list contains the same namespaces", false
		}
	}

	return "", true
}
