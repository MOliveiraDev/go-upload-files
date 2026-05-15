package models

import (
	"time"

	"github.com/google/uuid"
)

// VideoStatus representa os estados possíveis de um vídeo no sistema.
type VideoStatus string

const (
	StatusUploaded   VideoStatus = "UPLOADED"   // Upload concluído
	StatusProcessing VideoStatus = "PROCESSING" // Sendo transcodificado/comprimido
	StatusReady      VideoStatus = "READY"      // Pronto para ser assistido/stream
	StatusFailed     VideoStatus = "FAILED"     // Erro no processamento
	StatusDeleted    VideoStatus = "DELETED"    // Vídeo apagado
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
