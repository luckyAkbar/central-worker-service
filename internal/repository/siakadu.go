package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type siakadRepo struct {
	db *gorm.DB
}

// NewSiakadRepository create a new siakad repository
func NewSiakadRepository(db *gorm.DB) model.SiakaduRepository {
	return &siakadRepo{
		db,
	}
}

func (r *siakadRepo) Create(ctx context.Context, res *model.SiakaduScrapingResult) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"res": utils.Dump(res),
	})

	logger.Info("start saving siakadu scraping result to db")

	if err := r.db.WithContext(ctx).Create(res).Error; err != nil {
		logger.WithError(err).Error("failed to save siakadu scraping result to db")
		return err
	}

	return nil
}

func (r *siakadRepo) FindByID(ctx context.Context, id string) (*model.SiakaduScrapingResult, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"id":  id,
	})

	logger.Info("start finding siakadu scraping result")

	res := &model.SiakaduScrapingResult{}
	err := r.db.WithContext(ctx).Model(&model.SiakaduScrapingResult{}).Where("id = ?", id).Take(res).Error
	switch err {
	default:
		logger.WithError(err).Error("database error when finding siakadu scraping result")
		return nil, err

	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound

	case nil:
		return res, nil
	}
}
