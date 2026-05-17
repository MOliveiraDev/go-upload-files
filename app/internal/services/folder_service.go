package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
)

type FolderRepository interface {
	CreateFolder(ctx context.Context, folder *models.Folder) error
	GetFolderByID(ctx context.Context, id uuid.UUID) (*models.Folder, error)
}

type FolderService struct {
	repo FolderRepository
}

// Construtor
func NewFolderService(repo FolderRepository) *FolderService {
	return &FolderService{
		repo: repo,
	}
}

// CreateFolder contém a regra principal de iniciar uma pasta
func (s *FolderService) CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*models.Folder, error) {
	// Validação básica
	if name == "" {
		return nil, fmt.Errorf("o nome da pasta não pode ser vazio")
	}

	// o Service deve ir no banco verificar se a pasta pai realmente existe
	if parentID != nil {
		_, err := s.repo.GetFolderByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("pasta de destino não encontrada: %w", err)
		} // Se der erro, ele bloqueia a criação na hora!
	}

	// Montamos a estrutura
	newFolder := &models.Folder{
		ID:        uuid.New(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now(), // Marca o exato momento que foi criada
	}

	// Mandamos pro Repository salvar de fato na tabela do Postgres
	if err := s.repo.CreateFolder(ctx, newFolder); err != nil {
		return nil, fmt.Errorf("erro ao salvar a pasta no banco de dados: %w", err)
	}

	return newFolder, nil
}
