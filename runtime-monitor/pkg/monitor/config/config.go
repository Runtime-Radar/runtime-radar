package config

import (
	"github.com/google/go-cmp/cmp"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/api"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/model"
	"google.golang.org/protobuf/testing/protocmp"
)

type Selector struct {
	EventsClient, TracingPolicies, TracingPolicyStates bool
}

func Diff(oldCfg, newCfg *model.Config) (sel Selector, changed bool) {
	// TODO: disabled for debugging purposes, anyways it's a small optimization
	// if !newCfg.CreatedAt.After(oldCfg.CreatedAt) {
	// 	return
	// }

	if !cmp.Equal(oldCfg.Config.AllowList, newCfg.Config.AllowList, protocmp.Transform()) ||
		!cmp.Equal(oldCfg.Config.DenyList, newCfg.Config.DenyList, protocmp.Transform()) ||
		!cmp.Equal(oldCfg.Config.AggregationOptions, newCfg.Config.AggregationOptions, protocmp.Transform()) {
		sel.EventsClient, changed = true, true
	}

	if !cmp.Equal(oldCfg.Config.TracingPolicies, newCfg.Config.TracingPolicies, protocmp.Transform(), protocmp.IgnoreFields(&api.TracingPolicy{}, "enabled")) {
		sel.TracingPolicies, changed = true, true
	}

	if !cmp.Equal(oldCfg.Config.TracingPolicies, newCfg.Config.TracingPolicies, protocmp.Transform(), protocmp.IgnoreFields(&api.TracingPolicy{}, "name", "description", "yaml")) {
		sel.TracingPolicyStates, changed = true, true
	}

	return
}
