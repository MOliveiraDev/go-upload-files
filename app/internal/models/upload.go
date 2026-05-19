package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadStatus string

const (
	UploadStatusInitiated UploadStatus = "INITIATED"
	UploadStatusAborted   UploadStatus = "ABORTED"
	UploadStatusCompleted UploadStatus = "COMPLETED"
)

type UploadSession struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	OwnerID     uuid.UUID      `json:"owner_id" gorm:"type:uuid;index"`
	FolderID    *uuid.UUID     `json:"folder_id,omitempty" gorm:"type:uuid;index"`
	FileName    string         `json:"file_name"`
	ContentType string         `json:"content_type"`
	Size        int64          `json:"size"`
	StorageKey  string         `json:"storage_key" gorm:"uniqueIndex"`
	S3UploadID  string         `json:"s3_upload_id"`
	Status      UploadStatus   `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
