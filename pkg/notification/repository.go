package notification

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, n *Notification) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, n *Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}
