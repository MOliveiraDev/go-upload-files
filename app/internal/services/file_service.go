package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/MOliveiraDev/go-upload-files/internal/repositories"
	"github.com/google/uuid"
)

var (
	ErrFileNotFound          = errors.New("arquivo não encontrado")
	ErrFileForbidden         = errors.New("o arquivo não pertence ao usuário autenticado")
	ErrFileNameAlreadyExists = errors.New("já existe um arquivo com esse nome")
	ErrInvalidFileName       = errors.New("o nome do arquivo é obrigatório")
)

type FileFilter struct {
	Name         string
	Status       string
	UploadedFrom *time.Time
	UploadedTo   *time.Time
	FolderID     *uuid.UUID
	Page         int
	PageSize     int
}

type FileStorage interface {
	GetDownloadURL(file *models.File) string
}

type FileService struct {
	storage    FileStorage
	repo       repositories.FileRepository
	folderRepo repositories.FolderRepository
}

func NewFileService(storage FileStorage, repo repositories.FileRepository, folderRepo repositories.FolderRepository) *FileService {
	return &FileService{
		storage:    storage,
		repo:       repo,
		folderRepo: folderRepo,
	}
}

func (s *FileService) ListFiles(ctx context.Context, ownerID uuid.UUID, filter FileFilter) ([]models.File, error) {
	repoFilter := repositories.FileListFilter{
		Name:         filter.Name,
		Status:       filter.Status,
		UploadedFrom: filter.UploadedFrom,
		UploadedTo:   filter.UploadedTo,
		FolderID:     filter.FolderID,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	}

	files, err := s.repo.ListFiles(ctx, ownerID, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("listar arquivos: %w", err)
	}

	return files, nil
}

func (s *FileService) GetFile(ctx context.Context, ownerID, fileID uuid.UUID) (*models.File, error) {
	return s.requireOwnedFile(ctx, ownerID, fileID)
}

func (s *FileService) ListFilesInFolder(ctx context.Context, ownerID, folderID uuid.UUID) ([]models.File, error) {
	if _, err := s.requireOwnedFolder(ctx, ownerID, folderID); err != nil {
		return nil, err
	}

	files, err := s.repo.ListFiles(ctx, ownerID, repositories.FileListFilter{
		FolderID: &folderID,
	})
	if err != nil {
		return nil, fmt.Errorf("listar arquivos da pasta: %w", err)
	}

	return files, nil
}

func (s *FileService) RenameFile(ctx context.Context, ownerID, fileID uuid.UUID, name string) (*models.File, error) {
	file, err := s.requireOwnedFile(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}

	normalizedName, err := normalizeFileName(name)
	if err != nil {
		return nil, err
	}

	file.Name = normalizedName
	file.UpdatedAt = time.Now()

	if err := s.repo.UpdateFile(ctx, file); err != nil {
		if errors.Is(err, repositories.ErrFileNameAlreadyExists) {
			return nil, ErrFileNameAlreadyExists
		}

		return nil, fmt.Errorf("renomear arquivo: %w", err)
	}

	return file, nil
}

func (s *FileService) EditMetadata(ctx context.Context, ownerID, fileID uuid.UUID, contentType string) (*models.File, error) {
	file, err := s.requireOwnedFile(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}

	trimmedContentType := strings.TrimSpace(contentType)
	if trimmedContentType != "" {
		file.Type = trimmedContentType
	}
	file.UpdatedAt = time.Now()

	if err := s.repo.UpdateFile(ctx, file); err != nil {
		if errors.Is(err, repositories.ErrFileNameAlreadyExists) {
			return nil, ErrFileNameAlreadyExists
		}

		return nil, fmt.Errorf("editar metadados do arquivo: %w", err)
	}

	return file, nil
}

func (s *FileService) MoveFile(ctx context.Context, ownerID, fileID uuid.UUID, folderID *uuid.UUID) (*models.File, error) {
	file, err := s.requireOwnedFile(ctx, ownerID, fileID)
	if err != nil {
		return nil, err
	}

	if folderID != nil {
		if _, err := s.requireOwnedFolder(ctx, ownerID, *folderID); err != nil {
			return nil, err
		}
	}

	file.FolderID = folderID
	file.UpdatedAt = time.Now()

	if err := s.repo.UpdateFile(ctx, file); err != nil {
		if errors.Is(err, repositories.ErrFileNameAlreadyExists) {
			return nil, ErrFileNameAlreadyExists
		}

		return nil, fmt.Errorf("mover arquivo: %w", err)
	}

	return file, nil
}

func (s *FileService) DeleteFiles(ctx context.Context, ownerID uuid.UUID, fileIDs []uuid.UUID) error {
	if err := s.repo.DeleteFilesByIDs(ctx, ownerID, fileIDs); err != nil {
		return fmt.Errorf("deletar arquivos: %w", err)
	}

	return nil
}

func (s *FileService) GetDownloadURL(ctx context.Context, ownerID, fileID uuid.UUID) (string, *models.File, error) {
	file, err := s.requireOwnedFile(ctx, ownerID, fileID)
	if err != nil {
		return "", nil, err
	}

	if file.URL != "" {
		return file.URL, file, nil
	}

	if s.storage == nil {
		return "", nil, errors.New("storage não configurado")
	}

	downloadURL := s.storage.GetDownloadURL(file)
	if downloadURL == "" {
		return "", nil, errors.New("não foi possível gerar a url de download")
	}

	return downloadURL, file, nil
}

func (s *FileService) requireOwnedFile(ctx context.Context, ownerID, fileID uuid.UUID) (*models.File, error) {
	file, err := s.repo.GetFileByIDAndOwner(ctx, fileID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("buscar arquivo por owner: %w", err)
	}
	if file != nil {
		return file, nil
	}

	existingFile, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("buscar arquivo por id: %w", err)
	}
	if existingFile == nil {
		return nil, ErrFileNotFound
	}

	return nil, ErrFileForbidden
}

func (s *FileService) requireOwnedFolder(ctx context.Context, ownerID, folderID uuid.UUID) (*models.Folder, error) {
	if s.folderRepo == nil {
		return nil, errors.New("repositório de pastas não configurado")
	}

	folder, err := s.folderRepo.GetFolderByIDAndOwner(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("buscar pasta por owner: %w", err)
	}
	if folder != nil {
		return folder, nil
	}

	existingFolder, err := s.folderRepo.GetFolderByID(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("buscar pasta por id: %w", err)
	}
	if existingFolder == nil {
		return nil, ErrFolderNotFound
	}

	return nil, ErrFolderForbidden
}

func normalizeFileName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", ErrInvalidFileName
	}

	return trimmedName, nil
}
