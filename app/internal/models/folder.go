package models

import (
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	ParentID  *uuid.UUID `json:"parent_id" db:"parent_id"`   
	CreatedAt time.Time  `json:"created_at" db:"created_at"` 
}
