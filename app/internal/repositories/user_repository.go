package repositories

import (
	"context"
	"errors"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrUserEmailAlreadyExists = errors.New("email já registrado")

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	if r == nil || r.db == nil {
		return errors.New("Conexão com o banco de dados não configurada")
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrUserEmailAlreadyExists
	}

	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("Conexão com o banco de dados não configurada")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}
