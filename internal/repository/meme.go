package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type memeRepo struct {
	db *gorm.DB
}

func NewMemeRepository(db *gorm.DB) model.GagMemeRepository {
	return &memeRepo{
		db,
	}
}

func (r *memeRepo) FindRandomGagMeme(ctx context.Context) (*model.GagMeme, error) {
	logger := logrus.WithContext(ctx)

	logger.Info("start fetching gag meme data")

	meme := &model.GagMeme{}
	if err := r.db.WithContext(ctx).Order("RANDOM()").Model(&model.GagMeme{}).First(meme).Error; err != nil {
		logger.WithError(err).Error("failed to fetch gag meme data")
		return nil, err
	}

	logger.Info(utils.Dump(meme))

	return meme, nil
}
