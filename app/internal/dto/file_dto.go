package dto

import (
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

type UploadFileRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	ContentType string `json:"content_type" validate:"required,oneof=video/mp4 video/x-matroska video/quicktime"` // Aceita mp4, mkv, mov
	Size        int64  `json:"size" validate:"required,gt=0"`
}

type Part struct {
	PartNumber int32
	ETag       string
}

func (req *UploadFileRequest) ValidateRequest() error {
	return Validate.Struct(req)
}
