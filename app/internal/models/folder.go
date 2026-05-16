package models

import (
	"time"

	"github.com/google/uuid"
)

// Folder representa uma pasta no sistema de arquivos
type Folder struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	ParentID  *uuid.UUID `json:"parent_id" db:"parent_id"`   // Ponteiro para permitir null (raiz)
	CreatedAt time.Time  `json:"created_at" db:"created_at"` // data de criação
}
