package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type subscriptionRepo struct {
	db *gorm.DB
}

// NewSubscriptionRepository will create an object that represent the subscription Repository interface
func NewSubscriptionRepository(db *gorm.DB) model.SubscriptionRepository {
	return &subscriptionRepo{
		db,
	}
}

func (r *subscriptionRepo) Create(ctx context.Context, subs *model.Subscription) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"subs": utils.Dump(subs),
	})

	logger.Info("start saving subscription data")

	if err := r.db.WithContext(ctx).Create(subs).Error; err != nil {
		logger.WithError(err).Error("failed to save subscription data")
		return err
	}

	return nil
}

func (r *subscriptionRepo) FindSubscriptions(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"limit":  limit,
		"offset": offset,
	})

	logger.Info("start fetching subscriptions data")

	var subs []model.Subscription
	if err := r.db.WithContext(ctx).Model(&model.Subscription{}).Limit(limit).Offset(offset).Find(&subs).Error; err != nil {
		logger.WithError(err).Error("failed to fetch subscriptions data")
		return nil, err
	}

	return subs, nil
}

func (r *subscriptionRepo) FindSubscription(ctx context.Context, subType model.SubscriptionType, channel model.SubscriptionChannel, userID string) (*model.Subscription, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"subType": subType,
		"channel": channel,
		"userID":  userID,
	})

	logger.Info("start finding subscription data")

	sub := &model.Subscription{}
	err := r.db.WithContext(ctx).Where("type = ? AND channel = ? AND user_reference_id = ?", subType, channel, userID).Take(&sub).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find subscription data")
		return nil, err

	case gorm.ErrRecordNotFound:
		logger.Info("subscription data not found")
		return nil, ErrNotFound

	case nil:
		return sub, nil
	}
}

func (r *subscriptionRepo) Delete(ctx context.Context, id string) error {
	logger := logrus.WithContext(ctx).WithField("id", id)

	logger.Info("start deleting subscription data")

	sub := &model.Subscription{}
	if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(sub).Error; err != nil {
		logger.WithError(err).Error("failed to delete subscription data")
		return err
	}

	logger.Info("Deleted: ", utils.Dump(sub)

	return nil
}
