package repository

import (
	"context"

	"github.com/emorenkov/scorehub/pkg/notification"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, n *notification.Notification) error
	GetByID(ctx context.Context, id int64) (*notification.Notification, error)
	List(ctx context.Context, userID int64) ([]notification.Notification, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, n *notification.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *GormRepository) GetByID(ctx context.Context, id int64) (*notification.Notification, error) {
	var n notification.Notification
	if err := r.db.WithContext(ctx).First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *GormRepository) List(ctx context.Context, userID int64) ([]notification.Notification, error) {
	var notifications []notification.Notification
	query := r.db.WithContext(ctx)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}
