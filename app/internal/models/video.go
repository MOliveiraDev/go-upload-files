package models

import (
	"time"

	"github.com/google/uuid"
)

type VideoStatus string

const (
	StatusUploaded   VideoStatus = "UPLOADED"   
	StatusProcessing VideoStatus = "PROCESSING" 
	StatusReady      VideoStatus = "READY"      
	StatusFailed     VideoStatus = "FAILED"    
	StatusDeleted    VideoStatus = "DELETED"    
)

// Video metadata que será salva no banco de dados
type Video struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Type      string      `json:"type" db:"type"`             
	Size      int64       `json:"size" db:"size"`             
	Path      string      `json:"path" db:"path"`            
	URL       string      `json:"url" db:"url"`               
	Status    VideoStatus `json:"status" db:"status"`         
	CreatedAt time.Time   `json:"created_at" db:"created_at"` 
}
