package updater

import (
	"context"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_LoadConfig(t *testing.T) {
	ctx := context.Background()
	mc := minimock.NewController(t)

	configMock := database.NewConfigRepositoryMock(mc)

	s := Service{
		configRepository: configMock,
		config: &model.Config{
			Base: model.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Config: &model.ConfigJSON{
				Version: "1",
				AllowList: &api.Config_ConfigJSON_Filter{
					Namespaces: []string{"ns1", "ns2", "ns3"},
				},
				DenyList: &api.Config_ConfigJSON_Filter{
					Namespaces: []string{"ns4", "ns5"},
				},
			},
		},
	}

	expected := &model.Config{
		Base: model.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Config: &model.ConfigJSON{
			Version: "1",
			AllowList: &api.Config_ConfigJSON_Filter{
				Namespaces: []string{"ns1", "ns3"},
			},
			DenyList: &api.Config_ConfigJSON_Filter{
				Namespaces: []string{"ns6"},
			},
		},
	}

	configMock.GetLastMock.Set(func(_ context.Context, _ bool) (cp1 *model.Config, err error) {
		return expected, nil
	})

	err := s.loadConfig(ctx)
	require.NoError(t, err)

	applied := s.Config()
	assert.Equal(t, expected, applied)

	// without changes
	err = s.loadConfig(ctx)
	require.NoError(t, err)

	assert.Equal(t, expected, applied)
}

func TestService_Equal(t *testing.T) {
	tests := []struct {
		name   string
		oldCfg *model.Config
		newCfg *model.Config
		want   bool
	}{
		{
			name: "equal",
			oldCfg: &model.Config{
				Config: &model.ConfigJSON{
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns1", "ns2", "ns3"},
					},
					DenyList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns4", "ns5"},
					},
				},
			},
			newCfg: &model.Config{
				Config: &model.ConfigJSON{
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns1", "ns2", "ns3"},
					},
					DenyList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns4", "ns5"},
					},
				},
			},
			want: true,
		},
		{
			name: "not equal",
			oldCfg: &model.Config{
				Config: &model.ConfigJSON{
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns1", "ns2", "ns3"},
					},
					DenyList: &api.Config_ConfigJSON_Filter{},
				},
			},
			newCfg: &model.Config{
				Config: &model.ConfigJSON{
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns1", "ns2", "ns3"},
					},
					DenyList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"ns4", "ns5"},
					},
				},
			},
			want: false,
		},
		{
			name:   "default config",
			oldCfg: model.DefaultConfig,
			newCfg: nil,
			want:   true,
		},
		{
			name: "empty lists",
			oldCfg: &model.Config{
				Config: &model.ConfigJSON{
					AllowList: &api.Config_ConfigJSON_Filter{Namespaces: []string{}},
					DenyList:  &api.Config_ConfigJSON_Filter{},
				},
			},
			newCfg: &model.Config{},
			want:   true,
		},
		{
			name: "empty config JSON",
			oldCfg: &model.Config{
				Config: &model.ConfigJSON{},
			},
			newCfg: &model.Config{},
			want:   true,
		},
		{
			name:   "empty config",
			oldCfg: &model.Config{},
			newCfg: &model.Config{},
			want:   true,
		},
		{
			name:   "nil",
			oldCfg: nil,
			newCfg: nil,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}

			assert.Equalf(t, tt.want, s.Equal(tt.oldCfg, tt.newCfg), "Equal(%v, %v)", tt.oldCfg, tt.newCfg)
		})
	}
}
