package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileStatus string

const (
	StatusUploaded   FileStatus = "UPLOADED"
	StatusProcessing FileStatus = "PROCESSING"
	StatusReady      FileStatus = "READY"
	StatusFailed     FileStatus = "FAILED"
	StatusDeleted    FileStatus = "DELETED"
)

type File struct {
	ID        uuid.UUID      `json:"id" db:"id" gorm:"type:uuid;primaryKey"`
	Name      string         `json:"name" db:"name"`
	OwnerID   uuid.UUID      `json:"owner_id" db:"owner_id" gorm:"type:uuid;index"`
	FolderID  *uuid.UUID     `json:"folder_id,omitempty" db:"folder_id" gorm:"type:uuid;index"`
	Type      string         `json:"type" db:"type"`
	Size      int64          `json:"size" db:"size"`
	Path      string         `json:"path" db:"path"`
	URL       string         `json:"url" db:"url"`
	Status    FileStatus     `json:"status" db:"status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" db:"deleted_at" gorm:"index"`
}
