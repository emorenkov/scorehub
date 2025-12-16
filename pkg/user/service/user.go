package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/common/models"
	"gorm.io/gorm"
)

// User describes the user business logic exposed to transports.
type User interface {
	Create(ctx context.Context, name, email string) (*models.User, error)
	Get(ctx context.Context, id int64) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, id int64, name, email string) (*models.User, error)
	Delete(ctx context.Context, id int64) error
}

type Repository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id int64) error
}

type user struct {
	repo Repository
}

func NewService(repo Repository) User {
	return &user{repo: repo}
}

func (s *user) Create(ctx context.Context, name, email string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if name == "" || email == "" {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "name and email are required")
	}

	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "check user email uniqueness")
	}

	u := &models.User{Name: name, Email: email}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "create user")
	}
	return u, nil
}

func (s *user) Get(ctx context.Context, id int64) (*models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewStatusError(http.StatusNotFound, "user not found")
		}
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "get user")
	}
	return u, nil
}

func (s *user) List(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "list users")
	}
	return users, nil
}

func (s *user) Update(ctx context.Context, id int64, name, email string) (*models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewStatusError(http.StatusNotFound, "user not found")
		}
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "get user")
	}
	if name = strings.TrimSpace(name); name != "" {
		u.Name = name
	}
	if email = strings.TrimSpace(strings.ToLower(email)); email != "" {
		if email != u.Email {
			existing, err := s.repo.GetByEmail(ctx, email)
			if err == nil && existing.ID != u.ID {
				return nil, apperrors.NewStatusError(http.StatusBadRequest, "email already exists")
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "check user email uniqueness")
			}
		}
		u.Email = email
	}
	if err := s.repo.Update(ctx, u); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewStatusError(http.StatusNotFound, "user not found")
		}
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "update user")
	}
	return u, nil
}

func (s *user) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewStatusError(http.StatusNotFound, "user not found")
		}
		return apperrors.WrapStatus(err, http.StatusInternalServerError, "delete user")
	}
	return nil
}
