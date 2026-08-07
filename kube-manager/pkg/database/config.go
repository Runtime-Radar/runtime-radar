package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate minimock -i ConfigRepository -o . -s _mock.go -g
type ConfigRepository interface {
	Add(ctx context.Context, ls ...*model.Config) error
	GetLast(ctx context.Context, preloadData bool) (*model.Config, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ConfigDatabase struct {
	*gorm.DB
}

// Add adds new entry to the database, it can add multiple instances at once.
func (cd *ConfigDatabase) Add(ctx context.Context, cs ...*model.Config) error {
	if len(cs) == 0 {
		return nil
	}

	return cd.WithContext(ctx).Create(cs).Error
}

// GetLast returns last config.
func (cd *ConfigDatabase) GetLast(ctx context.Context, preloadData bool) (*model.Config, error) {
	c := &model.Config{}

	err := cd.preloadData(ctx, preloadData).
		Order("created_at desc").
		Take(&c).
		Error

	return c, err
}

func (cd *ConfigDatabase) Delete(ctx context.Context, id uuid.UUID) error {
	return cd.WithContext(ctx).
		Delete(&model.Config{Base: model.Base{ID: id}}).
		Error
}

func (cd *ConfigDatabase) preloadData(ctx context.Context, preloadData bool) *gorm.DB {
	if preloadData {
		return cd.WithContext(ctx).
			// This should load all associations without nested: https://gorm.io/docs/preload.html#Preload-All
			Preload(clause.Associations)
	}
	return cd.WithContext(ctx)
}
