package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type telegramRepo struct {
	db *gorm.DB
}

// NewTelegramRepository create new telegram repository
func NewTelegramRepository(db *gorm.DB) model.TelegramRepository {
	return &telegramRepo{
		db,
	}
}

func (r *telegramRepo) CreateUser(ctx context.Context, user *model.TelegramUser) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  utils.DumpIncomingContext(ctx),
		"user": utils.Dump(user),
	})

	logger.Info("start creating user to telegram database")

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.WithError(err).Error("failed to create telegram user")
		return err
	}

	return nil
}

func (r *telegramRepo) FindUserByID(ctx context.Context, userID int64) (*model.TelegramUser, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.DumpIncomingContext(ctx),
		"userID": userID,
	})

	logger.Info("start finding telegram user by ID")

	user := &model.TelegramUser{}
	err := r.db.WithContext(ctx).Model(&model.TelegramUser{}).Where("id = ?", userID).Take(user).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find telegram user by ID")
		return nil, err

	case gorm.ErrRecordNotFound:
		logger.Info("telegram user not found")
		return nil, ErrNotFound

	case nil:
		return user, nil
	}
}
