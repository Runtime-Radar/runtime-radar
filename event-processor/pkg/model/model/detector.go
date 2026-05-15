package model

import (
	"time"
)

type Detector struct {
	ID          string    `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"index:,sort:desc"`
	Name        string    `gorm:"index"`
	Description string
	Version     uint `gorm:"primaryKey"`
	Author      string
	Contact     string
	License     string
	WasmBinary  []byte
	WasmHash    string // hex-encoded SHA-512 hash of wasm binary
}
