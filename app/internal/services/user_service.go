package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/MOliveiraDev/go-upload-files/internal/dto"
	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/MOliveiraDev/go-upload-files/internal/repositories"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req *dto.CreateUserRequest) (*models.User, error) {
	if err := req.ValidateRequest(); err != nil {
		return nil, fmt.Errorf("dados de entrada inválidos: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("falha ao processar a senha de segurança: %w", err)
	}

	newUser := &models.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, fmt.Errorf("falha ao salvar usuário no banco de dados: %w", err)
	}

	return newUser, nil
}

func (s *UserService) LoginUser(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	if err := req.ValidateRequest(); err != nil {
		return nil, fmt.Errorf("dados de login inválidos: %w", err)
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("email ou senha inválidos")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		return nil, fmt.Errorf("email ou senha inválidos")
	}

	token := "jwt-token-aqui"

	return &dto.LoginResponse{
		JWT: token,
	}, nil
}
