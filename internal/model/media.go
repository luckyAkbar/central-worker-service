package model

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"
)

// list of storage location
const (
	LocationLocal = "LOCAL"
)

// UploadImageInput input to upload media image
type UploadImageInput struct {
	ImageName string `json:"image_name" validate:"required"`
	IsPrivate bool   `json:"is_private"`

	format string
}

// SetFormat set format file
func (uii *UploadImageInput) SetFormat(format string) {
	uii.format = format
}

// Validate validate struct
func (uii *UploadImageInput) Validate() error {
	return validator.Struct(uii)
}

// GenerateFullFilename generates full filename
func (uii *UploadImageInput) GenerateFullFilename(randName string) string {
	return fmt.Sprintf("%s%s%s", uii.ImageName, randName, uii.format)
}

// Image represent database table
type Image struct {
	ID            string    `json:"id"`
	Filename      string    `json:"filename"`
	CreatedAt     time.Time `json:"created_at"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	IsPrivate     bool      `json:"is_private"`
	AccessKey     string    `json:"access_key"`
	Location      string    `json:"location"`
}

// ImageUsecase usecase
type ImageUsecase interface {
	Upload(ctx context.Context, input *UploadImageInput, file *multipart.FileHeader) (*Image, UsecaseError)
}

// ImageRepository repository
type ImageRepository interface {
	Create(ctx context.Context, image *Image) error
}
