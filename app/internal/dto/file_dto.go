package dto

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type InitUploadRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	ContentType string  `json:"contentType" validate:"required"`
	Size        int64   `json:"size" validate:"required,gt=0"`
	FolderID    *string `json:"folderId"`
}

type CompleteUploadRequest struct {
	Parts []UploadedPart `json:"parts" validate:"required,min=1,dive"`
}

type UploadedPart struct {
	PartNumber int32  `json:"partNumber" validate:"required,gt=0"`
	ETag       string `json:"etag" validate:"required"`
}

func (req *InitUploadRequest) ValidateRequest() error {
	return Validate.Struct(req)
}

func (req *CompleteUploadRequest) ValidateRequest() error {
	return Validate.Struct(req)
}
