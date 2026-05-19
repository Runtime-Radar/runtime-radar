package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/filters"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/api"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/database"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/model"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// namespaceRegex is the regular expression to match namespaces passed in filters.
// see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
var namespaceRegex = regexp.MustCompile("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$")

type ConfigGeneric struct {
	api.UnimplementedConfigControllerServer

	ConfigRepository database.ConfigRepository
	Monitor          monitor.Monitor
	NodeName         string
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
		model.Base{ID: id},
		(*model.ConfigJSON)(req.GetConfig()),
	}
	if err := cg.ConfigRepository.Add(ctx, cfg); err != nil {
		return nil, status.Errorf(codes.Internal, "can't add config: %v", err)
	}

	// Changes applied instantly when requested. However, there can be multiple nodes in cluster,
	// and consequently multiple runtime-monitor instances, each of which can update config.
	// To handle this there will be background worker doing same check periodically and calling Reinit when needed.
	_, oldCfg := cg.Monitor.Config()

	log.Debug().Interface("old_config", oldCfg).Msgf("Old monitor config")
	log.Debug().Interface("new_config", cfg).Msgf("New monitor config")

	sel, changed := config.Diff(oldCfg, cfg)
	if changed {
		log.Info().
			Interface("config", cfg).
			Interface("selector", sel).
			Msgf("Monitor config changed, re-initializing")

		cg.Monitor.Update(sel, cfg)
	} else {
		log.Debug().Msgf("Monitor config didn't change")
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ConfigGeneric) Read(ctx context.Context, _ *emptypb.Empty) (*api.Config, error) {
	cfg, err := cg.ConfigRepository.GetLast(ctx, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "config not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "can't read config: %v", err)
	}

	resp := &api.Config{
		Id:     cfg.ID.String(),
		Config: (*api.Config_ConfigJSON)(cfg.Config),
	}

	return resp, nil
}

func (cg *ConfigGeneric) ResetToDefault(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	m := map[string]any{
		"updated_at": time.Now(),
	}
	if err := cg.ConfigRepository.UpdateWithMap(ctx, model.DefaultConfig.ID, m); err != nil {
		return nil, status.Errorf(codes.Internal, "can't update config: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ConfigGeneric) Status(ctx context.Context, _ *emptypb.Empty) (*api.ConfigStatus, error) {
	cfg, err := cg.ConfigRepository.GetLast(ctx, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "config not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "can't read config: %v", err)
	}

	sel, changed := config.Diff(model.DefaultConfig, cfg)

	lastInitErr := ""
	if err := cg.Monitor.LastInitErr(); err != nil {
		lastInitErr = err.Error()
	}

	resp := &api.ConfigStatus{
		Default:                cfg.ID == model.DefaultConfig.ID,
		DefaultTracingPolicies: !changed || !sel.TracingPolicies,
		LastInitError:          lastInitErr,
		NodeName:               cg.NodeName,
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

	for i, f := range req.Config.GetAllowList() {
		if reason, ok := cg.validateFilter(f); !ok {
			return fmt.Sprintf("AllowList[%d] is invalid: %s", i, reason), false
		}
	}
	for i, f := range req.Config.GetDenyList() {
		if reason, ok := cg.validateFilter(f); !ok {
			return fmt.Sprintf("DenyList[%d] is invaild: %s", i, reason), false
		}
	}

	return "", true
}

func (cg *ConfigGeneric) validateFilter(f *tetragon.Filter) (string, bool) {
	for i, br := range f.GetBinaryRegex() {
		if _, err := regexp.Compile(br); err != nil {
			return fmt.Sprintf("BinaryRegex[%d] is invalid regex: %v", i, err), false
		}
	}

	for i, pr := range f.GetPodRegex() {
		if _, err := regexp.Compile(pr); err != nil {
			return fmt.Sprintf("PodRegex[%d] is invalid regex: %v", i, err), false
		}
	}

	for i, ar := range f.GetArgumentsRegex() {
		if _, err := regexp.Compile(ar); err != nil {
			return fmt.Sprintf("ArgumentsRegex[%d] is invalid regex: %v", i, err), false
		}
	}

	for i, ns := range f.GetNamespace() {
		if ns != "" && !namespaceRegex.MatchString(ns) {
			return fmt.Sprintf("Namespace[%d] is invalid DNS label name", i), false
		}
	}

	// check if filter contains correct labels selectors
	if _, err := filters.BuildFilter(context.Background(), f, []filters.OnBuildFilter{&filters.LabelsFilter{}}); err != nil {
		return fmt.Sprintf("invalid labels: %v", err), false
	}

	return "", true
}
