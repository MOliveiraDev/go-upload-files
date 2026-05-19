package repositories

import (
	"context"
	"errors"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadSessionRepository interface {
	Create(ctx context.Context, upload *models.UploadSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.UploadSession, error)
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.UploadSession, error)
	Update(ctx context.Context, upload *models.UploadSession) error
}

type uploadSessionRepository struct {
	db *gorm.DB
}

func NewUploadSessionRepository(db *gorm.DB) UploadSessionRepository {
	return &uploadSessionRepository{db: db}
}

func (r *uploadSessionRepository) Create(ctx context.Context, upload *models.UploadSession) error {
	if r == nil || r.db == nil {
		return errors.New(ErrDatabaseNotConfigured.Error())
	}

	return r.db.WithContext(ctx).Create(upload).Error
}

func (r *uploadSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UploadSession, error) {
	if r == nil || r.db == nil {
		return nil, errors.New(ErrDatabaseNotConfigured.Error())
	}

	var upload models.UploadSession
	err := r.db.WithContext(ctx).First(&upload, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &upload, nil
}

func (r *uploadSessionRepository) GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.UploadSession, error) {
	if r == nil || r.db == nil {
		return nil, errors.New(ErrDatabaseNotConfigured.Error())
	}

	var upload models.UploadSession
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&upload).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &upload, nil
}

func (r *uploadSessionRepository) Update(ctx context.Context, upload *models.UploadSession) error {
	if r == nil || r.db == nil {
		return errors.New(ErrDatabaseNotConfigured.Error())
	}

	return r.db.WithContext(ctx).Save(upload).Error
}
