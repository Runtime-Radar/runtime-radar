package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
)

const ConfigVersion Version = "1"

type ConfigJSON api.Config_ConfigJSON

type Config struct {
	Base
	Config *ConfigJSON `gorm:"type:jsonb"`
}

// TableName method implements Tabler interface and makes GORM name the table of Config "kube_manager_configs" instead of just "configs".
// This is done in order to keep more generic "configs" available for possible use by dynamic config mechanism. However, this is not yet
// implemented and can be done differently. One of the possible scenarios is that "kube_manager_configs" will be slightly modified and become
// the base for storing dynamic configs of all components instead of having different config tables, in this case this method will be removed.
func (c *Config) TableName() string {
	return "kube_manager_configs"
}

func (c *Config) AllowNamespaces() []string {
	if c == nil || c.Config == nil || c.Config.AllowList == nil || c.Config.AllowList.Namespaces == nil {
		return []string{}
	}

	return c.Config.AllowList.Namespaces
}

func (c *Config) DenyNamespaces() []string {
	if c == nil || c.Config == nil || c.Config.DenyList == nil || c.Config.DenyList.Namespaces == nil {
		return []string{}
	}

	return c.Config.DenyList.Namespaces
}

func (s *ConfigJSON) Scan(src interface{}) error {
	b := src.([]byte)
	return json.Unmarshal(b, s)
}

func (s *ConfigJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

var (
	DefaultConfig = &Config{
		Base: Base{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		Config: &ConfigJSON{
			Version: string(ConfigVersion),
			AllowList: &api.Config_ConfigJSON_Filter{
				Namespaces: []string{},
			},
			DenyList: &api.Config_ConfigJSON_Filter{
				Namespaces: []string{},
			},
		},
	}
)
