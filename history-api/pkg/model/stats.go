package model

import (
	enf_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

type DetectorCounter struct {
	DetectorID string
	Severity   enf_model.Severity
	Count      int
}
