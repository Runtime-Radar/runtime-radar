package service

import (
	"context"
	"slices"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/cache"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/database"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
	"github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model/convert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EnforcerGeneric struct {
	api.UnimplementedEnforcerServer

	RuleMatcher    cache.RuleMatcher
	RuleRepository database.RuleRepository
}

func (eg *EnforcerGeneric) EvaluatePolicyRuntimeEvent(ctx context.Context, req *api.EvaluatePolicyRuntimeEventReq) (*api.EvaluatePolicyRuntimeEventReq, error) {
	if reason, ok := eg.validateRuntimeEventRequest(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	// TODO: add cache.WithCluster when we support multiple clusters
	opts := []cache.MatchOption{
		cache.WithNamespace(req.GetAction().GetArgs().GetNamespace()),
		cache.WithImageName(req.GetAction().GetArgs().GetImageName()),
		cache.WithRegistry(req.GetAction().GetArgs().GetRegistry()),
		cache.WithPod(req.GetAction().GetArgs().GetPod()),
		cache.WithContainer(req.GetAction().GetArgs().GetContainer()),
		cache.WithNode(req.GetAction().GetArgs().GetNode()),
	}
	rules, err := eg.RuleMatcher.MatchRules(ctx, model.RuleTypeRuntime, opts...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't match rules: %v", err)
	}

	rules = filterRulesByFunc(rules, func(r *model.Rule) bool {
		return whitelistedByPattern(req.GetAction().GetArgs().GetBinary(), r.Rule.Whitelist.GetBinaries())
	})

	for _, ev := range req.GetResult().GetEvents() {
		var eventSeverity model.Severity
		eventSeverity.Set(ev.GetSeverity())

		rs := filterRulesByFunc(rules, func(r *model.Rule) bool {
			return slices.Contains(r.Rule.Whitelist.GetThreats(), ev.GetDetectorId())
		})
		block, notify := filterRulesBySeverity(rs, eventSeverity)

		ev.Policy = &api.Policy{
			BlockBy:  convert.RulesToProto(block),
			NotifyBy: convert.RulesToProto(notify),
		}
	}

	return req, nil
}

func (eg *EnforcerGeneric) validateRuntimeEventRequest(req *api.EvaluatePolicyRuntimeEventReq) (reason string, ok bool) {
	if a := req.GetAction(); a == nil {
		return "no action", false
	} else if a.GetArgs() == nil {
		return "no args", false
	}

	return "", true
}

func filterRulesByFunc(rules []*model.Rule, filter func(*model.Rule) bool) []*model.Rule {
	cloned := slices.Clone(rules) // avoid modifying outer slice

	j := 0
	for _, r := range cloned {
		if !filter(r) {
			cloned[j] = r
			j++
		}
	}
	cloned = cloned[:j]

	return cloned
}

func filterRulesBySeverity(rs []*model.Rule, severity model.Severity) (block, notify []*model.Rule) {
	for _, r := range rs {
		blockSeverity, notifySeverity := model.UnsetSeverity, model.UnsetSeverity

		if r.Rule.Block != nil {
			blockSeverity.Set(r.Rule.Block.Severity)
		}
		if r.Rule.Notify != nil {
			notifySeverity.Set(r.Rule.Notify.Severity)
		}

		if severity >= blockSeverity {
			block = append(block, r)
		}

		if severity >= notifySeverity {
			notify = append(notify, r)
		}
	}

	return
}

func whitelistedByPattern(s string, patterns []string) bool {
	for _, p := range patterns {
		if g := glob.MustCompile(p); g.Match(s) {
			return true
		}
	}
	return false
}
