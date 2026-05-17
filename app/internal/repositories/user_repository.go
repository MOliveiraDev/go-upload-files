package repositories

import (
"context"

"github.com/MOliveiraDev/go-upload-files/internal/models"
)

type UserRepository interface {
CreateUser(ctx context.Context, user *models.User) error
FindByEmail(ctx context.Context, email string) (*models.User, error)
}
