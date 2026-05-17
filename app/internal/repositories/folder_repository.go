package repositories

import (
	"context"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/google/uuid"
)

type FolderRepository interface {
	CreateFolder(ctx context.Context, folder *models.Folder) error
	GetFolderByID(ctx context.Context, id uuid.UUID) (*models.Folder, error)
}
