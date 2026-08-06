package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ConfigCreate(t *testing.T) {
	ctx := context.Background()
	cd := &ConfigDatabase{DB: db}

	tests := []struct {
		name   string
		source model.Config
	}{
		{
			name: "empty",
			source: model.Config{
				Base: model.Base{ID: uuid.New()},
				Config: &model.ConfigJSON{
					Version: "1",
				},
			},
		},
		{
			name: "with deny list",
			source: model.Config{
				Base: model.Base{ID: uuid.New()},
				Config: &model.ConfigJSON{
					Version: "1",
					DenyList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"deny"},
					},
				},
			},
		},
		{
			name: "with allow list",
			source: model.Config{
				Base: model.Base{ID: uuid.New()},
				Config: &model.ConfigJSON{
					Version: "1",
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"allow"},
					},
				},
			},
		},
		{
			name: "with deny and allow list",
			source: model.Config{
				Base: model.Base{ID: uuid.New()},
				Config: &model.ConfigJSON{
					Version: "1",
					AllowList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"allow"},
					},
					DenyList: &api.Config_ConfigJSON_Filter{
						Namespaces: []string{"deny"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cd.Add(ctx, &tt.source)
			require.NoError(t, err)

			stored, err := cd.GetLast(ctx, true)
			require.NoError(t, err)
			require.NotNil(t, stored)

			assert.GreaterOrEqual(t, time.Now(), stored.CreatedAt)
			assert.Less(t, stored.CreatedAt, time.Now().Add(time.Second))
			tt.source.CreatedAt = stored.CreatedAt

			assert.GreaterOrEqual(t, time.Now(), stored.UpdatedAt)
			assert.Less(t, stored.UpdatedAt, time.Now().Add(time.Second))
			tt.source.UpdatedAt = stored.UpdatedAt

			assert.Equal(t, tt.source.CreatedAt, tt.source.UpdatedAt)
			assert.Equal(t, tt.source, *stored)
		})
	}
}
