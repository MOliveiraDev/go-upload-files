package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID        uuid.UUID      `json:"id" db:"id" gorm:"type:uuid;primaryKey"`
	Name      string         `json:"name" db:"name"`
	OwnerID   uuid.UUID      `json:"owner_id" db:"owner_id" gorm:"type:uuid;index"`
	ParentID  *uuid.UUID     `json:"parent_id,omitempty" db:"parent_id" gorm:"type:uuid;index"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" db:"deleted_at" gorm:"index"`
}
