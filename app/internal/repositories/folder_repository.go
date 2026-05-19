package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrFolderNameAlreadyExists = errors.New("Nome da pasta já existe")
var ErrDatabaseNotConfigured = errors.New("Conexão com o banco de dados não configurada")

type FolderRepository interface {
	CreateFolder(ctx context.Context, folder *models.Folder) error
	GetFolderByID(ctx context.Context, id uuid.UUID) (*models.Folder, error)
	GetFolderByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.Folder, error)
	ListFolders(ctx context.Context, ownerID uuid.UUID, name string) ([]models.Folder, error)
	ListSubfolders(ctx context.Context, ownerID uuid.UUID, parentID *uuid.UUID) ([]models.Folder, error)
	UpdateFolder(ctx context.Context, folder *models.Folder) error
	DeleteFolder(ctx context.Context, folder *models.Folder) error
	NameExists(ctx context.Context, ownerID uuid.UUID, parentID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
}

type folderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) FolderRepository {
	return &folderRepository{db: db}
}

func (r *folderRepository) CreateFolder(ctx context.Context, folder *models.Folder) error {
	if r == nil || r.db == nil {
		return ErrDatabaseNotConfigured
	}

	exists, err := r.NameExists(ctx, folder.OwnerID, folder.ParentID, folder.Name, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrFolderNameAlreadyExists
	}

	return r.db.WithContext(ctx).Create(folder).Error
}

func (r *folderRepository) GetFolderByID(ctx context.Context, id uuid.UUID) (*models.Folder, error) {
	if r == nil || r.db == nil {
		return nil, ErrDatabaseNotConfigured
	}

	var folder models.Folder
	err := r.db.WithContext(ctx).First(&folder, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &folder, nil
}

func (r *folderRepository) GetFolderByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.Folder, error) {
	if r == nil || r.db == nil {
		return nil, ErrDatabaseNotConfigured
	}

	var folder models.Folder
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&folder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &folder, nil
}

func (r *folderRepository) ListFolders(ctx context.Context, ownerID uuid.UUID, name string) ([]models.Folder, error) {
	if r == nil || r.db == nil {
		return nil, ErrDatabaseNotConfigured
	}

	query := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at ASC")
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		query = query.Where("name ILIKE ?", "%"+trimmed+"%")
	}

	var folders []models.Folder
	if err := query.Find(&folders).Error; err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *folderRepository) ListSubfolders(ctx context.Context, ownerID uuid.UUID, parentID *uuid.UUID) ([]models.Folder, error) {
	if r == nil || r.db == nil {
		return nil, ErrDatabaseNotConfigured
	}

	query := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at ASC")
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	var folders []models.Folder
	if err := query.Find(&folders).Error; err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *folderRepository) UpdateFolder(ctx context.Context, folder *models.Folder) error {
	if r == nil || r.db == nil {
		return ErrDatabaseNotConfigured
	}

	exists, err := r.NameExists(ctx, folder.OwnerID, folder.ParentID, folder.Name, &folder.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrFolderNameAlreadyExists
	}

	return r.db.WithContext(ctx).Save(folder).Error
}

func (r *folderRepository) DeleteFolder(ctx context.Context, folder *models.Folder) error {
	if r == nil || r.db == nil {
		return ErrDatabaseNotConfigured
	}

	return r.db.WithContext(ctx).Delete(folder).Error
}

func (r *folderRepository) NameExists(ctx context.Context, ownerID uuid.UUID, parentID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	if r == nil || r.db == nil {
		return false, ErrDatabaseNotConfigured
	}

	query := r.db.WithContext(ctx).Model(&models.Folder{}).Where("owner_id = ? AND name = ?", ownerID, name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
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
