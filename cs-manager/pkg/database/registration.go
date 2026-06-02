package database

import (
	"context"

	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RegistrationRepository interface {
	Add(ctx context.Context, rs ...*model.Registration) error
	GetLastSuccessful(ctx context.Context, hash string, preloadData bool) (*model.Registration, error)
}

type RegistrationDatabase struct {
	*gorm.DB
}

func (rd *RegistrationDatabase) Add(ctx context.Context, rs ...*model.Registration) error {
	if len(rs) == 0 {
		return nil
	}

	return rd.WithContext(ctx).Create(rs).Error
}

func (rd *RegistrationDatabase) GetLastSuccessful(ctx context.Context, hash string, preloadData bool) (*model.Registration, error) {
	r := &model.Registration{}

	err := rd.preloadData(ctx, preloadData).
		Where(&model.Registration{Status: model.RegistrationStatusOK, TokenHash: hash}).
		Order("created_at desc").
		Take(&r).
		Error

	return r, err
}

func (rd *RegistrationDatabase) preloadData(ctx context.Context, preloadData bool) *gorm.DB {
	if preloadData {
		return rd.WithContext(ctx).
			// This should load all associations without nested: https://gorm.io/docs/preload.html#Preload-All
			Preload(clause.Associations)
	}
	return rd.WithContext(ctx)
}
