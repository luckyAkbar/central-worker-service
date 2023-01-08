package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type siakadRepo struct {
	db     *gorm.DB
	cacher model.Cacher
}

// NewSiakadRepository create a new siakad repository
func NewSiakadRepository(db *gorm.DB, cacher model.Cacher) model.SiakaduRepository {
	return &siakadRepo{
		db,
		cacher,
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

func (r *siakadRepo) GetLastNPMFromCache(ctx context.Context) (int, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
	})

	logger.Info("start getting last npm from cache")

	res, err := r.cacher.Get(ctx, model.LastNPMCacheKey)
	switch err {
	default:
		logger.WithError(err).Error("failed to get last npm from cache")
		return 0, err

	case redis.Nil:
		return 0, ErrNotFound

	case nil:
		break
	}

	var npm int
	if err := json.Unmarshal([]byte(res), &npm); err != nil {
		logger.WithError(err).Error("failed to unmarshal cache res to npm: ", res)
		return 0, err
	}

	return npm, nil
}

func (r *siakadRepo) SetLastNPMToCache(ctx context.Context, npm int) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"npm": npm,
	})

	logger.Info("start setting last npm to cache")

	val := fmt.Sprintf("%d", npm)
	exp := time.Hour * 2400

	if err := r.cacher.Set(ctx, model.LastNPMCacheKey, val, exp); err != nil {
		logger.WithError(err).Error("failed to set LastNPMCacheKey")
		return err
	}

	return nil
}
