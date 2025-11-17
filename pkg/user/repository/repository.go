package repository

import (
	"context"

	"github.com/emorenkov/scorehub/pkg/common/models"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *GormRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).Where("deleted = FALSE").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).Where("email = ? AND deleted = FALSE", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("deleted = FALSE").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *GormRepository) Update(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *GormRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND deleted = FALSE", id).
		Updates(map[string]any{"deleted": true})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
