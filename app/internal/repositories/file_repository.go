package repositories

import (
"context"

"github.com/MOliveiraDev/go-upload-files/internal/models"
)

type FileRepository interface {
CreateFile(ctx context.Context, file *models.File) error
}
