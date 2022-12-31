package usecase

import (
	"context"
	"mime"
	"mime/multipart"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type imageUsecase struct {
	imageRepo model.ImageRepository
}

// NewImageUsecase creates a new image usecase
func NewImageUsecase(imageRepo model.ImageRepository) model.ImageUsecase {
	return &imageUsecase{
		imageRepo,
	}
}

func (u *imageUsecase) Upload(ctx context.Context, input *model.UploadImageInput, file *multipart.FileHeader) (*model.Image, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
		"file":  utils.Dump(file),
	})

	logger.Info("start upload image usecase")

	if err := input.Validate(); err != nil {
		logger.Info("validation failed on upload image usecase: ", err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         MsgInvalidInput,
		}
	}

	if err := helper.FilterImageMimetype(file); err != nil {
		logger.Info("mimetype is not allowed")
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         "mimetype is not allowed",
		}
	}

	exts, err := mime.ExtensionsByType(file.Header["Content-Type"][0])
	if err != nil {
		logger.Info("failed to associate mimetype with extention: ", err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         "mimetype is not known",
		}
	}

	input.SetFormat(exts[0])

	if file.Size > config.ImageMediaMaxSizeBytes() {
		logger.Info("image size is too large")
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         "filesize is too large",
		}
	}

	if err := helper.SaveMediaImageToLocalStorage(file, config.ImageMediaLocalStorage(), input.GenerateFullFilename(helper.GenerateID())); err != nil {
		logger.Error("failed to save image to local storage: ", err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgInternalError,
		}
	}

	image := &model.Image{
		ID:            helper.GenerateID(),
		Filename:      input.GenerateFullFilename(helper.GenerateID()),
		CreatedAt:     time.Now().UTC(),
		FileSizeBytes: file.Size,
		IsPrivate:     input.IsPrivate,
		AccessKey:     helper.GenerateToken(config.ImageMediaTokenLength()),
		Location:      model.LocationLocal,
	}

	if err := u.imageRepo.Create(ctx, image); err != nil {
		logger.WithError(err).Error("failed to save image to db")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return image, model.NilUsecaseError
}
