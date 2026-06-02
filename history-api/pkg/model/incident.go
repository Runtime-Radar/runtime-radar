package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	enf_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

type Incident struct {
	Base
	BlockBy   BlockBy            `gorm:"type:jsonb"`
	NotifyBy  NotifyBy           `gorm:"type:jsonb"`
	Severity  enf_model.Severity `gorm:"index"`
	EventType string             `gorm:"index"`
	EventID   *uuid.UUID         `gorm:"index"`
}

type BlockBy []string // atm it's just a slice, but can be extended with use of JSONB capabilities later on

func (s *BlockBy) Scan(src interface{}) error {
	b := src.([]byte)
	return json.Unmarshal(b, s)
}

func (s BlockBy) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

type NotifyBy []string // atm it's just a slice, but can be extended with use of JSONB capabilities later on

func (s *NotifyBy) Scan(src interface{}) error {
	b := src.([]byte)
	return json.Unmarshal(b, s)
}

func (s NotifyBy) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}
