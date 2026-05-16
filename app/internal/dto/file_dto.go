package dto

import (
	"github.com/go-playground/validator/v10"
)

// Inicia o validador para usarmos nas nossas requisições
var Validate = validator.New()

type UploadFileRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	ContentType string `json:"content_type" validate:"required,oneof=video/mp4 video/x-matroska video/quicktime"` // Aceita mp4, mkv, mov
	Size        int64  `json:"size" validate:"required,gt=0"`
}

// ValidateRequest é uma função helper para validar nossos DTOs
func (req *UploadFileRequest) ValidateRequest() error {
	return Validate.Struct(req)
}
