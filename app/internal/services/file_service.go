package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/dto"
	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/MOliveiraDev/go-upload-files/internal/repositories"
	"github.com/MOliveiraDev/go-upload-files/internal/storage/aws"
	"github.com/google/uuid"
)

var (
	ErrFileNotFound          = errors.New("arquivo não encontrado")
	ErrFileForbidden         = errors.New("o arquivo não pertence ao usuário autenticado")
	ErrFileNameAlreadyExists = errors.New("já existe um arquivo com esse nome")
	ErrInvalidFileName       = errors.New("o nome do arquivo é obrigatório")
	ErrUploadNotFound        = errors.New("upload não encontrado")
	ErrUploadForbidden       = errors.New("o upload não pertence ao usuário autenticado")
	ErrUploadAlreadyClosed   = errors.New("o upload já foi concluído ou abortado")
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
	InitMultipartUpload(ctx context.Context, fileKey string) (string, error)
	UploadPart(ctx context.Context, fileKey, uploadID string, partNumber int32, chunk []byte) (string, error)
	CompleteMultipartUpload(ctx context.Context, fileKey, uploadID string, parts []aws.Part) error
	AbortMultipartUpload(ctx context.Context, fileKey, uploadID string) error
	GetDownloadURL(file *models.File) string
}

type FileService struct {
	storage    FileStorage
	repo       repositories.FileRepository
	folderRepo repositories.FolderRepository
	uploadRepo repositories.UploadSessionRepository
}

func NewFileService(storage FileStorage, repo repositories.FileRepository, folderRepo repositories.FolderRepository, uploadRepo repositories.UploadSessionRepository) *FileService {
	return &FileService{
		storage:    storage,
		repo:       repo,
		folderRepo: folderRepo,
		uploadRepo: uploadRepo,
	}
}

func (s *FileService) InitUpload(ctx context.Context, ownerID uuid.UUID, req dto.InitUploadRequest) (*models.UploadSession, error) {
	if err := req.ValidateRequest(); err != nil {
		return nil, fmt.Errorf("dados de upload inválidos: %w", err)
	}

	folderID, err := parseOptionalFolderID(req.FolderID)
	if err != nil {
		return nil, fmt.Errorf("folderId inválido: %w", err)
	}

	if folderID != nil {
		if _, err := s.requireOwnedFolder(ctx, ownerID, *folderID); err != nil {
			return nil, err
		}
	}

	fileName, err := normalizeFileName(req.Name)
	if err != nil {
		return nil, err
	}

	uploadID := uuid.New()
	fileKey := fmt.Sprintf("uploads/%s-%s", uploadID.String(), fileName)

	s3UploadID, err := s.storage.InitMultipartUpload(ctx, fileKey)
	if err != nil {
		return nil, fmt.Errorf("iniciar multipart upload: %w", err)
	}

	uploadSession := &models.UploadSession{
		ID:          uploadID,
		OwnerID:     ownerID,
		FolderID:    folderID,
		FileName:    fileName,
		ContentType: strings.TrimSpace(req.ContentType),
		Size:        req.Size,
		StorageKey:  fileKey,
		S3UploadID:  s3UploadID,
		Status:      models.UploadStatusInitiated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.uploadRepo.Create(ctx, uploadSession); err != nil {
		_ = s.storage.AbortMultipartUpload(ctx, fileKey, s3UploadID)
		return nil, fmt.Errorf("persistir sessão de upload: %w", err)
	}

	return uploadSession, nil
}

func (s *FileService) UploadPart(ctx context.Context, ownerID, uploadID uuid.UUID, partNumber int32, chunk []byte) (string, *models.UploadSession, error) {
	upload, err := s.requireOwnedUpload(ctx, ownerID, uploadID)
	if err != nil {
		return "", nil, err
	}
	if upload.Status != models.UploadStatusInitiated {
		return "", nil, ErrUploadAlreadyClosed
	}

	etag, err := s.storage.UploadPart(ctx, upload.StorageKey, upload.S3UploadID, partNumber, chunk)
	if err != nil {
		return "", nil, fmt.Errorf("enviar parte do upload: %w", err)
	}

	return etag, upload, nil
}

func (s *FileService) CompleteUpload(ctx context.Context, ownerID, uploadID uuid.UUID, parts []dto.UploadedPart) (*models.File, error) {
	upload, err := s.requireOwnedUpload(ctx, ownerID, uploadID)
	if err != nil {
		return nil, err
	}
	if upload.Status != models.UploadStatusInitiated {
		return nil, ErrUploadAlreadyClosed
	}

	completedParts := make([]aws.Part, 0, len(parts))
	for _, part := range parts {
		completedParts = append(completedParts, aws.Part{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		})
	}

	if err := s.storage.CompleteMultipartUpload(ctx, upload.StorageKey, upload.S3UploadID, completedParts); err != nil {
		return nil, fmt.Errorf("concluir multipart upload: %w", err)
	}

	file := &models.File{
		ID:        uuid.New(),
		Name:      upload.FileName,
		OwnerID:   upload.OwnerID,
		FolderID:  upload.FolderID,
		Type:      upload.ContentType,
		Size:      upload.Size,
		Path:      upload.StorageKey,
		Status:    models.StatusReady,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateFile(ctx, file); err != nil {
		if errors.Is(err, repositories.ErrFileNameAlreadyExists) {
			return nil, ErrFileNameAlreadyExists
		}

		return nil, fmt.Errorf("persistir arquivo concluído: %w", err)
	}

	upload.Status = models.UploadStatusCompleted
	upload.UpdatedAt = time.Now()
	if err := s.uploadRepo.Update(ctx, upload); err != nil {
		return nil, fmt.Errorf("atualizar status do upload: %w", err)
	}

	return file, nil
}

func (s *FileService) AbortUpload(ctx context.Context, ownerID, uploadID uuid.UUID) error {
	upload, err := s.requireOwnedUpload(ctx, ownerID, uploadID)
	if err != nil {
		return err
	}
	if upload.Status != models.UploadStatusInitiated {
		return ErrUploadAlreadyClosed
	}

	if err := s.storage.AbortMultipartUpload(ctx, upload.StorageKey, upload.S3UploadID); err != nil {
		return fmt.Errorf("abortar multipart upload: %w", err)
	}

	upload.Status = models.UploadStatusAborted
	upload.UpdatedAt = time.Now()
	if err := s.uploadRepo.Update(ctx, upload); err != nil {
		return fmt.Errorf("atualizar status do upload: %w", err)
	}

	return nil
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

func (s *FileService) requireOwnedUpload(ctx context.Context, ownerID, uploadID uuid.UUID) (*models.UploadSession, error) {
	if s.uploadRepo == nil {
		return nil, errors.New("repositório de uploads não configurado")
	}

	upload, err := s.uploadRepo.GetByIDAndOwner(ctx, uploadID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("buscar upload por owner: %w", err)
	}
	if upload != nil {
		return upload, nil
	}

	existingUpload, err := s.uploadRepo.GetByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("buscar upload por id: %w", err)
	}
	if existingUpload == nil {
		return nil, ErrUploadNotFound
	}

	return nil, ErrUploadForbidden
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

func parseOptionalFolderID(folderID *string) (*uuid.UUID, error) {
	if folderID == nil || strings.TrimSpace(*folderID) == "" {
		return nil, nil
	}

	parsedValue, err := uuid.Parse(strings.TrimSpace(*folderID))
	if err != nil {
		return nil, err
	}

	return &parsedValue, nil
}
