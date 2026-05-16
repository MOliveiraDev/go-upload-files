package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MOliveiraDev/go-upload-files/internal/dto"
	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/MOliveiraDev/go-upload-files/internal/storage/aws"
)

type StorageManager interface {
	InitMultipartUpload(ctx context.Context, fileKey string) (string, error)
	UploadPart(ctx context.Context, fileKey, uploadID string, partNumber int32, chunk []byte) (string, error)
	CompleteMultipartUpload(ctx context.Context, fileKey, uploadID string, parts []aws.Part) error
	AbortMultipartUpload(ctx context.Context, fileKey, uploadID string) error
}

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
}

type FileService struct {
	storage StorageManager
	repo    FileRepository
}

func NewFileService(storage StorageManager, repo FileRepository) *FileService {
	return &FileService{
		storage: storage,
		repo:    repo,
	}
}

// Função chamada quando o usuário quer iniciar o Upload
func (s *FileService) StartUpload(ctx context.Context, req dto.UploadFileRequest) (*models.File, string, error) {

	// Validar a requisição usando o nosso DTO
	if err := req.ValidateRequest(); err != nil {
		return nil, "", fmt.Errorf("dados de entrada inválidos: %w", err)
	}

	// Gerar informações base do arquivo
	fileID := uuid.New()

	// Avisar a AWS/S3 que um arquivo grande está a caminho
	fileKey := fmt.Sprintf("uploads/%s-%s", fileID.String(), req.Name)

	uploadID, err := s.storage.InitMultipartUpload(ctx, fileKey)
	if err != nil {
		return nil, "", fmt.Errorf("falha ao abrir sessão de upload no storage: %w", err)
	}

	// Criar a entidade para o Banco de Dados
	newFile := &models.File{ 
		ID:        fileID,
		Name:      req.Name,
		Type:      req.ContentType,
		Size:      req.Size,
		Path:      fileKey,                 
		Status:    models.StatusProcessing,
		CreatedAt: time.Now(),
	}

	// Mandar nosso Repository salvar no Postgres
	if err := s.repo.CreateFile(ctx, newFile); err != nil {
		// Se der erro de banco de dados, devemos mandar a AWS cancelar a sessão (Abort) para não poluir
		return nil, "", fmt.Errorf("falha ao salvar registro no banco de dados: %w", err)
	}

	return newFile, uploadID, nil
}
