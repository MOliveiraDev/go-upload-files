package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrFileNameAlreadyExists = errors.New("file name already exists")

type FileListFilter struct {
	Name         string
	Status       string
	UploadedFrom *time.Time
	UploadedTo   *time.Time
	FolderID     *uuid.UUID
	Page         int
	PageSize     int
}

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, id uuid.UUID) (*models.File, error)
	GetFileByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.File, error)
	ListFiles(ctx context.Context, ownerID uuid.UUID, filter FileListFilter) ([]models.File, error)
	UpdateFile(ctx context.Context, file *models.File) error
	DeleteFilesByIDs(ctx context.Context, ownerID uuid.UUID, fileIDs []uuid.UUID) error
	NameExists(ctx context.Context, ownerID uuid.UUID, folderID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *models.File) error {
	if r == nil || r.db == nil {
		return errors.New(ErrDatabaseNotConfigured.Error())
	}

	exists, err := r.NameExists(ctx, file.OwnerID, file.FolderID, file.Name, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrFileNameAlreadyExists
	}

	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetFileByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	if r == nil || r.db == nil {
		return nil, errors.New(ErrDatabaseNotConfigured.Error())
	}

	var file models.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &file, nil
}

func (r *fileRepository) GetFileByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.File, error) {
	if r == nil || r.db == nil {
		return nil, errors.New(ErrDatabaseNotConfigured.Error())
	}

	var file models.File
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &file, nil
}

func (r *fileRepository) ListFiles(ctx context.Context, ownerID uuid.UUID, filter FileListFilter) ([]models.File, error) {
	if r == nil || r.db == nil {
		return nil, errors.New(ErrDatabaseNotConfigured.Error())
	}

	query := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC")
	if trimmedName := strings.TrimSpace(filter.Name); trimmedName != "" {
		query = query.Where("name ILIKE ?", "%"+trimmedName+"%")
	}
	if trimmedStatus := strings.TrimSpace(filter.Status); trimmedStatus != "" {
		query = query.Where("status = ?", trimmedStatus)
	}
	if filter.UploadedFrom != nil {
		query = query.Where("created_at >= ?", *filter.UploadedFrom)
	}
	if filter.UploadedTo != nil {
		query = query.Where("created_at <= ?", *filter.UploadedTo)
	}
	if filter.FolderID != nil {
		query = query.Where("folder_id = ?", *filter.FolderID)
	}
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	var files []models.File
	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

func (r *fileRepository) UpdateFile(ctx context.Context, file *models.File) error {
	if r == nil || r.db == nil {
		return errors.New(ErrDatabaseNotConfigured.Error())
	}

	exists, err := r.NameExists(ctx, file.OwnerID, file.FolderID, file.Name, &file.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrFileNameAlreadyExists
	}

	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) DeleteFilesByIDs(ctx context.Context, ownerID uuid.UUID, fileIDs []uuid.UUID) error {
	if r == nil || r.db == nil {
		return errors.New(ErrDatabaseNotConfigured.Error())
	}

	if len(fileIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("owner_id = ? AND id IN ?", ownerID, fileIDs).Delete(&models.File{}).Error
}

func (r *fileRepository) NameExists(ctx context.Context, ownerID uuid.UUID, folderID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New(ErrDatabaseNotConfigured.Error())
	}

	query := r.db.WithContext(ctx).Model(&models.File{}).Where("owner_id = ? AND name = ?", ownerID, name)
	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", *folderID)
	}
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
