// Package repository contains all repository functionality
package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type diaryRepo struct {
	db     *gorm.DB
	cacher model.Cacher
}

// NewDiaryRepo create a new diary repository
func NewDiaryRepo(db *gorm.DB, cacher model.Cacher) model.DiaryRepository {
	return &diaryRepo{
		db,
		cacher,
	}
}

func (r *diaryRepo) Create(ctx context.Context, diary *model.Diary) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"diary": utils.Dump(diary),
	})

	logger.Info("start saving diary to database")

	if err := r.db.WithContext(ctx).Create(diary).Error; err != nil {
		logger.WithError(err).Error("failed to create diary data to database")
		return err
	}

	return nil
}

func (r *diaryRepo) FindDiaryByIDAndOwnerID(ctx context.Context, diaryID, ownerID string) (*model.Diary, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      helper.DumpContext(ctx),
		"diary_id": diaryID,
		"owner_id": ownerID,
	})

	logger.Info("start getting diary by id and owner ID from database")

	diary := &model.Diary{}
	err := r.db.WithContext(ctx).Model(&model.Diary{}).Where("id = ? AND owner_id = ?", diaryID, ownerID).Take(diary).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to get diary data from database")
		return nil, err

	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound

	case nil:
		return diary, nil
	}
}

func (r *diaryRepo) GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]model.Diary, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      helper.DumpContext(ctx),
		"start":    start,
		"end":      end,
		"owner_id": ownerID,
	})

	logger.Info("start getting diary by written date from database")

	var diaries []model.Diary
	if err := r.db.WithContext(ctx).Model(&model.Diary{}).Where("created_at >= ? AND created_at <= ? AND owner_id = ?", start, end, ownerID).Find(&diaries).Error; err != nil {
		logger.WithError(err).Error("failed to get diary data from database")
		return nil, err
	}

	if len(diaries) == 0 {
		return diaries, ErrNotFound
	}

	return diaries, nil
}

func (r *diaryRepo) DeleteByID(ctx context.Context, diaryID, ownerID string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"diaryID": diaryID,
		"ownerID": ownerID,
	})

	logger.Info("starting to delete diary by ID")

	deletedDiary := &model.Diary{}
	err := r.db.WithContext(ctx).Unscoped().
		Model(&model.Diary{}).Where("id = ? AND owner_id = ?", diaryID, ownerID).
		Delete(deletedDiary).Error

	if err != nil {
		logger.WithError(err).Error("failed to delete diary by ID")
		return err
	}

	logger.Info("deleted diary: ", utils.Dump(deletedDiary))

	return nil
}

func (r *diaryRepo) SetFrontendDiaryDataToCache(ctx context.Context, cacheKey string, exp time.Duration, data model.DiaryFrontendTemplateData) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := r.cacher.Set(ctx, cacheKey, string(val), exp); err != nil {
		return err
	}

	return nil
}

func (r *diaryRepo) GetFrontendDiaryDataFromCache(ctx context.Context, key string) (*model.DiaryFrontendTemplateData, error) {
	res, err := r.cacher.Get(ctx, key)
	switch err {
	default:
		return nil, err
	case redis.Nil:
		return nil, ErrNotFound
	case nil:
		break
	}

	data := model.DiaryFrontendTemplateData{}
	if err := json.Unmarshal([]byte(res), &data); err != nil {
		return nil, err
	}

	return &data, nil
}
