package service

import (
	enf_pb "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
)

type RuleGeneric struct {
	enf_pb.RuleControllerClient
}
