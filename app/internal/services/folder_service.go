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
	ErrFolderNotFound          = errors.New("pasta não encontrada")
	ErrFolderForbidden         = errors.New("a pasta não pertence ao usuário autenticado")
	ErrFolderNameAlreadyExists = errors.New("já existe uma pasta com esse nome")
	ErrInvalidFolderName       = errors.New("o nome da pasta é obrigatório")
	ErrInvalidFolderMove       = errors.New("não é possível mover a pasta para ela mesma ou para uma de suas descendentes")
)

type FolderItems struct {
	Folder  *models.Folder   `json:"folder"`
	Folders []models.Folder  `json:"folders"`
	Files   []map[string]any `json:"files"`
}

type FolderService struct {
	repo repositories.FolderRepository
}

func NewFolderService(repo repositories.FolderRepository) *FolderService {
	return &FolderService{repo: repo}
}

func (s *FolderService) CreateFolder(ctx context.Context, ownerID uuid.UUID, name string, parentID *uuid.UUID) (*models.Folder, error) {
	normalizedName, err := normalizeFolderName(name)
	if err != nil {
		return nil, err
	}

	if parentID != nil {
		if _, err := s.requireOwnedFolder(ctx, ownerID, *parentID); err != nil {
			return nil, err
		}
	}

	newFolder := &models.Folder{
		ID:        uuid.New(),
		Name:      normalizedName,
		OwnerID:   ownerID,
		ParentID:  parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateFolder(ctx, newFolder); err != nil {
		if errors.Is(err, repositories.ErrFolderNameAlreadyExists) {
			return nil, ErrFolderNameAlreadyExists
		}

		return nil, fmt.Errorf("create folder: %w", err)
	}

	return newFolder, nil
}

func (s *FolderService) ListFolders(ctx context.Context, ownerID uuid.UUID, name string) ([]models.Folder, error) {
	folders, err := s.repo.ListFolders(ctx, ownerID, name)
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}

	return folders, nil
}

func (s *FolderService) GetFolder(ctx context.Context, ownerID, folderID uuid.UUID) (*models.Folder, error) {
	return s.requireOwnedFolder(ctx, ownerID, folderID)
}

func (s *FolderService) ListSubfolders(ctx context.Context, ownerID, folderID uuid.UUID) ([]models.Folder, error) {
	if _, err := s.requireOwnedFolder(ctx, ownerID, folderID); err != nil {
		return nil, err
	}

	folders, err := s.repo.ListSubfolders(ctx, ownerID, &folderID)
	if err != nil {
		return nil, fmt.Errorf("list subfolders: %w", err)
	}

	return folders, nil
}

func (s *FolderService) GetFolderPath(ctx context.Context, ownerID, folderID uuid.UUID) ([]models.Folder, error) {
	folder, err := s.requireOwnedFolder(ctx, ownerID, folderID)
	if err != nil {
		return nil, err
	}

	path := []models.Folder{*folder}
	current := folder

	for current.ParentID != nil {
		parent, err := s.requireOwnedFolder(ctx, ownerID, *current.ParentID)
		if err != nil {
			return nil, err
		}

		path = append([]models.Folder{*parent}, path...)
		current = parent
	}

	return path, nil
}

func (s *FolderService) GetFolderItems(ctx context.Context, ownerID, folderID uuid.UUID) (*FolderItems, error) {
	folder, err := s.requireOwnedFolder(ctx, ownerID, folderID)
	if err != nil {
		return nil, err
	}

	subfolders, err := s.repo.ListSubfolders(ctx, ownerID, &folderID)
	if err != nil {
		return nil, fmt.Errorf("list folder items: %w", err)
	}

	return &FolderItems{
		Folder:  folder,
		Folders: subfolders,
		Files:   []map[string]any{},
	}, nil
}

func (s *FolderService) RenameFolder(ctx context.Context, ownerID, folderID uuid.UUID, name string) (*models.Folder, error) {
	folder, err := s.requireOwnedFolder(ctx, ownerID, folderID)
	if err != nil {
		return nil, err
	}

	normalizedName, err := normalizeFolderName(name)
	if err != nil {
		return nil, err
	}

	folder.Name = normalizedName
	folder.UpdatedAt = time.Now()

	if err := s.repo.UpdateFolder(ctx, folder); err != nil {
		if errors.Is(err, repositories.ErrFolderNameAlreadyExists) {
			return nil, ErrFolderNameAlreadyExists
		}

		return nil, fmt.Errorf("rename folder: %w", err)
	}

	return folder, nil
}

func (s *FolderService) MoveFolder(ctx context.Context, ownerID, folderID uuid.UUID, parentID *uuid.UUID) (*models.Folder, error) {
	folder, err := s.requireOwnedFolder(ctx, ownerID, folderID)
	if err != nil {
		return nil, err
	}

	if parentID != nil {
		if *parentID == folder.ID {
			return nil, ErrInvalidFolderMove
		}

		parent, err := s.requireOwnedFolder(ctx, ownerID, *parentID)
		if err != nil {
			return nil, err
		}

		if err := s.ensureNotDescendant(ctx, ownerID, folder.ID, parent); err != nil {
			return nil, err
		}
	}

	folder.ParentID = parentID
	folder.UpdatedAt = time.Now()

	if err := s.repo.UpdateFolder(ctx, folder); err != nil {
		if errors.Is(err, repositories.ErrFolderNameAlreadyExists) {
			return nil, ErrFolderNameAlreadyExists
		}

		return nil, fmt.Errorf("move folder: %w", err)
	}

	return folder, nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, ownerID, folderID uuid.UUID) error {
	folder, err := s.requireOwnedFolder(ctx, ownerID, folderID)
	if err != nil {
		return err
	}

	if err := s.deleteFolderRecursive(ctx, ownerID, folder); err != nil {
		return fmt.Errorf("delete folder recursively: %w", err)
	}

	return nil
}

func (s *FolderService) deleteFolderRecursive(ctx context.Context, ownerID uuid.UUID, folder *models.Folder) error {
	children, err := s.repo.ListSubfolders(ctx, ownerID, &folder.ID)
	if err != nil {
		return err
	}

	for i := range children {
		child := children[i]
		if err := s.deleteFolderRecursive(ctx, ownerID, &child); err != nil {
			return err
		}
	}

	return s.repo.DeleteFolder(ctx, folder)
}

func (s *FolderService) ensureNotDescendant(ctx context.Context, ownerID, folderID uuid.UUID, destination *models.Folder) error {
	current := destination

	for current != nil {
		if current.ID == folderID {
			return ErrInvalidFolderMove
		}

		if current.ParentID == nil {
			return nil
		}

		parent, err := s.requireOwnedFolder(ctx, ownerID, *current.ParentID)
		if err != nil {
			return err
		}

		current = parent
	}

	return nil
}

func (s *FolderService) requireOwnedFolder(ctx context.Context, ownerID, folderID uuid.UUID) (*models.Folder, error) {
	folder, err := s.repo.GetFolderByIDAndOwner(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("get folder by owner: %w", err)
	}
	if folder != nil {
		return folder, nil
	}

	existingFolder, err := s.repo.GetFolderByID(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("get folder by id: %w", err)
	}
	if existingFolder == nil {
		return nil, ErrFolderNotFound
	}

	return nil, ErrFolderForbidden
}

func normalizeFolderName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", ErrInvalidFolderName
	}

	return trimmedName, nil
}
