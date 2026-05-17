package repositories

import (
"context"

"github.com/google/uuid"
"github.com/MOliveiraDev/go-upload-files/internal/models"
)

type FolderRepository interface {
CreateFolder(ctx context.Context, folder *models.Folder) error
GetFolderByID(ctx context.Context, id uuid.UUID) (*models.Folder, error)
}
