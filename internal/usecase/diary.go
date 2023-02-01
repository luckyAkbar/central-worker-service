package usecase

import (
	"context"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type diaryUsecase struct {
	repo model.DiaryRepository
}

// NewDiaryUsecase create diary usecase
func NewDiaryUsecase(repo model.DiaryRepository) model.DiaryUsecase {
	return &diaryUsecase{
		repo,
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
