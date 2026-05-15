package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
)

const (
	ConfigVersion Version = "1"
)

var (
	DefaultConfig = &Config{
		Base: Base{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		Config: &ConfigJSON{
			Version: string(ConfigVersion),
			// Disable saving processed events by default
			HistoryControl: api.Config_ConfigJSON_NONE,
		},
	}
)

type ConfigJSON api.Config_ConfigJSON

type Config struct {
	Base
	Config *ConfigJSON `gorm:"type:jsonb"`
}

// TableName method implements Tabler interface and makes GORM name the table of Config "event_processor_configs" instead of just "configs".
// This is done in order to keep more generic "configs" available for possible use by dynamic config mechanism. However, this is not yet
// implemented and can be done differently. One of the possible scenarios is that "event_processor_configs" will be slightly modified and become
// the base for storing dynamic configs of all components instead of having different config tables, in this case this method will be removed.
func (Config) TableName() string {
	return "event_processor_configs"
}

func (s *ConfigJSON) Scan(src interface{}) error {
	b := src.([]byte)
	return json.Unmarshal(b, s)
}

func (s *ConfigJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}
