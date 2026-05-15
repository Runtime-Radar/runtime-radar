package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Version string

// Base struct represents basic model to be used by all data structs.
type Base struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time `gorm:"index:,sort:desc"`
	UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate callback sets ID as newly generated UUID. This makes it possible to not install 'uuid-ossp' extension
// with PostgreSQL. If ID was set already, just go ahead and do nothing.
func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	return nil
}
