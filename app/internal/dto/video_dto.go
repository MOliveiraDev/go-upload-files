package dto

import (
	"github.com/go-playground/validator/v10"
)

// Inicia o validador para usarmos nas nossas requisições
var Validate = validator.New()

// UploadVideoRequest representa os dados que esperamos receber do Front-end (Cliente)
// antes de começar o upload de fato, ou junto com o arquivo.
type UploadVideoRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	ContentType string `json:"content_type" validate:"required,oneof=video/mp4 video/x-matroska video/quicktime"` // Aceita mp4, mkv, mov
	Size        int64  `json:"size" validate:"required,gt=0"`
}

// ValidateRequest é uma função helper para validar nossos DTOs
func (req *UploadVideoRequest) ValidateRequest() error {
	return Validate.Struct(req)
}
