package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type imageRepo struct {
	db *gorm.DB
}

// NewImageRepository crete a repository for image
func NewImageRepository(db *gorm.DB) model.ImageRepository {
	return &imageRepo{db: db}
}

func (r *imageRepo) Create(ctx context.Context, image *model.Image) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"image": utils.Dump(image),
	})

	logger.Info("start saving image to db")

	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		logger.Error("failed to save image to db: ", err)
		return err
	}

	return nil
}
