package models

import (
	"time"

	"github.com/google/uuid"
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
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
    FolderID  *uuid.UUID  `json:"folder_id" db:"folder_id"`
	Type      string      `json:"type" db:"type"`             
	Size      int64       `json:"size" db:"size"`             
	Path      string      `json:"path" db:"path"`            
	URL       string      `json:"url" db:"url"`               
	Status    FileStatus `json:"status" db:"status"`         
	CreatedAt time.Time   `json:"created_at" db:"created_at"` 
}
