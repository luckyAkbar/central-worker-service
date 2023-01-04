package usecase

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type telegramUsecase struct {
	telegramRepo model.TelegramRepository
}

// NewTelegramUsecase create a new telegram usecase
func NewTelegramUsecase(telegramRepo model.TelegramRepository) model.TelegramUsecase {
	return &telegramUsecase{
		telegramRepo,
	}
}

func (u *telegramUsecase) RegisterSecretMessagingService(ctx context.Context, teleUser *model.TelegramUser) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":           helper.DumpContext(ctx),
		"telegram_user": utils.Dump(teleUser),
	})

	logger.Info("start register secret messaging service usecase")

	_, err := u.telegramRepo.FindUserByID(ctx, teleUser.ID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by ID")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case nil:
		return model.UsecaseError{
			UnderlyingError: ErrAlreadyExists,
			Message:         "user already registered",
		}

	case repository.ErrNotFound:
		break
	}

	logger.Info("saving telegram user to repository...")

	if err := u.telegramRepo.CreateUser(ctx, teleUser); err != nil {
		logger.WithError(err).Error("failed to register user to telegram user repo")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}
