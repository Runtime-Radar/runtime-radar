package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
	"gorm.io/gorm"
)

type AccessTokenRepository interface {
	Add(ctx context.Context, entity *model.AccessToken) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.AccessToken, error)
	GetPage(ctx context.Context, pageNum, pageSize int, filter, order any) ([]*model.AccessToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTokenHash(ctx context.Context, hash string) (*model.AccessToken, error)
	GetCount(ctx context.Context, filter interface{}) (int, error)
	InvalidateAll(ctx context.Context) error
}

type AccessTokenDatabase struct {
	*gorm.DB
}

func (db *AccessTokenDatabase) Add(ctx context.Context, accessToken *model.AccessToken) error {
	return db.WithContext(ctx).
		Create(accessToken).
		Error
}

func (db *AccessTokenDatabase) GetPage(ctx context.Context, pageNum, pageSize int, filter, order any) ([]*model.AccessToken, error) {
	res := []*model.AccessToken{}

	if filter == nil {
		filter = ""
	}

	sanitizedOrder, err := sanitizeOrder(order)
	if err != nil {
		return nil, err
	}

	err = db.WithContext(ctx).
		Where(filter).
		Order(sanitizedOrder).
		Limit(pageSize).
		Offset(pageSize * (pageNum - 1)).
		Find(&res).
		Error

	return res, err
}

func (db *AccessTokenDatabase) GetCount(ctx context.Context, filter any) (int, error) {
	if filter == nil {
		filter = ""
	}

	var count int64

	err := db.WithContext(ctx).
		Model(&model.AccessToken{}).
		Where(filter).
		Count(&count).
		Error

	return int(count), err
}

func (db *AccessTokenDatabase) GetByID(ctx context.Context, id uuid.UUID) (*model.AccessToken, error) {
	accessToken := &model.AccessToken{}
	err := db.WithContext(ctx).
		Where(model.AccessToken{Base: model.Base{ID: id}}).
		Take(accessToken).
		Error

	return accessToken, err
}

func (db *AccessTokenDatabase) Delete(ctx context.Context, id uuid.UUID) error {
	return db.WithContext(ctx).
		Delete(&model.AccessToken{Base: model.Base{ID: id}}).Error
}

func (db *AccessTokenDatabase) GetByTokenHash(ctx context.Context, hash string) (*model.AccessToken, error) {
	accessToken := &model.AccessToken{}
	err := db.WithContext(ctx).
		Where(model.AccessToken{Hash: hash}).
		Take(accessToken).
		Error

	return accessToken, err
}

func (db *AccessTokenDatabase) InvalidateAll(ctx context.Context) error {
	return db.WithContext(ctx).
		Model(&model.AccessToken{}).
		Where("invalidated_at is null").
		Updates(map[string]interface{}{
			"invalidated_at": time.Now(),
		}).Error
}
