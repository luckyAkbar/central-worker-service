package usecase

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type subscriptionUsecase struct {
	repo model.SubscriptionRepository
}

// NewSubsriptionUsecase will create an object that represent the subscription Usecase interface
func NewSubsriptionUsecase(repo model.SubscriptionRepository) model.SubscriptionUsecase {
	return &subscriptionUsecase{
		repo,
	}
}

func (u *subscriptionUsecase) Create(ctx context.Context, subs *model.Subscription) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"subs": utils.Dump(subs),
	})

	logger.Info("start saving subscription info")

	if err := u.repo.Create(ctx, subs); err != nil {
		logger.WithError(err).Error("failed to save subscription info")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}

func (u *subscriptionUsecase) FindSubscriptions(ctx context.Context, limit, offset int) ([]model.Subscription, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"limit":  limit,
		"offset": offset,
	})

	logger.Info("start fetching subscriptions")

	subs, err := u.repo.FindSubscriptions(ctx, limit, offset)
	if err != nil {
		logger.WithError(err).Error("failed to fetch subscriptions")
		return []model.Subscription{}, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	if len(subs) == 0 {
		return []model.Subscription{}, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}
	}

	return subs, model.NilUsecaseError
}

func (u *subscriptionUsecase) FindSubscription(ctx context.Context, subType model.SubscriptionType, channel model.SubscriptionChannel, userID string) (*model.Subscription, model.UsecaseError) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"subType": subType,
		"channel": channel,
		"userID":  userID,
	})

	logger.Info("start fetching subscription")

	subs, err := u.repo.FindSubscription(ctx, subType, channel, userID)
	switch err {
	default:
		logger.WithError(err).Error("failed to fetch subscription")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}
	case nil:
		return subs, model.NilUsecaseError
	}

}
