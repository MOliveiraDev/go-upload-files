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
	Type      string      `json:"type" db:"type"`             // mime type (ex: video/mp4)
	Size      int64       `json:"size" db:"size"`             // tamanho em bytes
	Path      string      `json:"path" db:"path"`             // caminho original no storage
	URL       string      `json:"url" db:"url"`               // URL pública
	Status    VideoStatus `json:"status" db:"status"`         // UPLOADED, PROCESSING, READY, FAILED, DELETED
	CreatedAt time.Time   `json:"created_at" db:"created_at"` // data de criação
}
