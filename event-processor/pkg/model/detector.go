package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
)

type MitreTactics []*detector_api.MitreTactic

type Detector struct {
	ID             string    `gorm:"primaryKey"`
	CreatedAt      time.Time `gorm:"index:,sort:desc"`
	Name           string    `gorm:"index"`
	Description    string
	Version        uint `gorm:"primaryKey"`
	Author         string
	Contact        string
	License        string
	WasmBinary     []byte
	WasmHash       string       // hex-encoded SHA-512 hash of wasm binary
	TacticsCovered MitreTactics `gorm:"type:jsonb"`
}

func (mt *MitreTactics) Scan(src any) error {
	b := src.([]byte)
	return json.Unmarshal(b, mt)
}

func (mt MitreTactics) Value() (driver.Value, error) {
	if len(mt) == 0 {
		return nil, nil
	}
	return json.Marshal(mt)
}
