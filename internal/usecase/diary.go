package usecase

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type diaryUsecase struct {
	repo   model.DiaryRepository
	yourls *helper.YourlsUtil
}

// NewDiaryUsecase create diary usecase
func NewDiaryUsecase(repo model.DiaryRepository, yourls *helper.YourlsUtil) model.DiaryUsecase {
	return &diaryUsecase{
		repo,
		yourls,
	}
}

func (u *diaryUsecase) Create(ctx context.Context, input *model.CreateDiaryInput) (*model.Diary, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	logger.Info("start create diary usecase")

	if err := input.Validate(); err != nil {
		logger.WithError(err).Info("validation error on creating diary")
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         MsgInvalidInput,
		}
	}

	diarySource, err := model.ParseStringToDiarySource(input.Source)
	if err != nil {
		logger.WithError(err).Info("invalid diary source")
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         err.Error(),
		}
	}

	diary := &model.Diary{
		ID:        helper.GenerateID(),
		OwnerID:   input.OwnerID,
		Note:      input.Note,
		CreatedAt: input.CreatedAt,
		TimeZone:  input.TimeZone,
		Source:    diarySource,
	}

	if err := u.repo.Create(ctx, diary); err != nil {
		logger.WithError(err).Error("failed to insert diary to database")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	logger.Info("finish create diary")

	return diary, model.NilUsecaseError
}

func (u *diaryUsecase) GetDiaryByID(ctx context.Context, diaryID, ownerID string) (*model.Diary, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      helper.DumpContext(ctx),
		"diary_id": diaryID,
		"owner_id": ownerID,
	})

	logger.Info("start getting diary by id and owner ID")

	diary, err := u.repo.FindDiaryByIDAndOwnerID(ctx, diaryID, ownerID)
	switch err {
	default:
		logger.WithError(err).Error("failed to get diary data from database")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		logger.WithError(err).Error("diary not found")
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	if diary.OwnerID != ownerID {
		logger.Info("diary searched is not owned by this user")
		return nil, model.UsecaseError{
			UnderlyingError: ErrForbidden,
			Message:         MsgForbidden,
		}
	}

	return diary, model.NilUsecaseError
}

func (u *diaryUsecase) GetDiariesByWrittenDateRange(ctx context.Context, start, end time.Time, ownerID string) ([]model.Diary, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      helper.DumpContext(ctx),
		"start":    start,
		"end":      end,
		"owner_id": ownerID,
	})

	logger.Info("start getting diary by written date and owner ID")

	diaries, err := u.repo.GetDiariesByWrittenDateRange(ctx, start, end, ownerID)
	switch err {
	default:
		logger.WithError(err).Error("failed to get diary data from database")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		logger.WithError(err).Error("diary not found")
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		return diaries, model.NilUsecaseError
	}

}

func (u *diaryUsecase) DeleteByID(ctx context.Context, diaryID, ownerID string) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"diaryID": diaryID,
		"ownerID": ownerID,
	})

	logger.Info("start deleteing diary by ID")

	diary, err := u.repo.FindDiaryByIDAndOwnerID(ctx, diaryID, ownerID)
	switch err {
	default:
		logger.WithError(err).Error("faild to find diary by id and owner id")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	// safety check
	if diary.OwnerID != ownerID {
		return model.UsecaseError{
			UnderlyingError: ErrForbidden,
			Message:         MsgForbidden,
		}
	}

	if err := u.repo.DeleteByID(ctx, diaryID, ownerID); err != nil {
		logger.WithError(err).Error("failed to delete diary")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}

func (u *diaryUsecase) PrepareRenderDiariesOnFrontend(ctx context.Context, cmd []string, diaries model.DiaryList) (string, error) {
	leadText := "basing dulu aja"
	diaryData := diaries.ToDiaryFrontendTemplateData(leadText)
	cacheKey := fmt.Sprintf("%s-%s", diaries[0].OwnerID, strings.Join(cmd, "-"))
	defaultDiaryOnFrontendExpiry := time.Hour * 1

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"func":     "diaryUsecase.PrepareRenderDiariesOnFrontend",
		"cacheKey": cacheKey,
	})

	err := u.repo.SetFrontendDiaryDataToCache(ctx, cacheKey, defaultDiaryOnFrontendExpiry, diaryData)
	if err != nil {
		logger.WithError(err).Error("failed to set frontend diary data to cache")
		return "", ErrInternal
	}

	url := fmt.Sprintf("%s?key=%s", config.DiaryFrontendBaseURL(), url.QueryEscape(cacheKey))
	shortURL, err := u.yourls.Shorten(ctx, &helper.ShortingInput{
		URL: url,
	})

	if err != nil {
		logger.WithError(err).Error("failed to shorten URL, may cause failed to response properly to telegram bot")
		shortURL = url
	}

	_, err = u.yourls.SetExpiry(ctx, &helper.ActionSetExpiryInput{
		ShortURL: shortURL,
		Expiry:   helper.Clock,
		AgeMod:   helper.Minutes,
		Age:      int64(defaultDiaryOnFrontendExpiry.Minutes()),
	})

	if err != nil {
		logger.WithError(err).Error("failed to set expiry to short URL, may cause failed to response properly to telegram bot")
		shortURL = url
	}

	return shortURL, nil
}

func (u *diaryUsecase) GetDiariesOnFrontendRenderData(ctx context.Context, key string) (*model.DiaryFrontendTemplateData, error) {
	res, err := u.repo.GetFrontendDiaryDataFromCache(ctx, key)
	switch err {
	default:
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"func":     "diaryUsecase.GetDiariesOnFrontendRenderData",
			"cacheKey": key,
		}).WithError(err).Error("failed to get frontend diary data from cache")
		return nil, ErrInternal
	case repository.ErrNotFound:
		return nil, ErrNotFound
	case nil:
		return res, nil
	}
}
