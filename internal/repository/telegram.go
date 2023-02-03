package repository

import (
	"context"
	"errors"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type telegramRepo struct {
	db     *gorm.DB
	cacher model.Cacher
}

// NewTelegramRepository create new telegram repository
func NewTelegramRepository(db *gorm.DB, cacher model.Cacher) model.TelegramRepository {
	return &telegramRepo{
		db,
		cacher,
	}
}

func (r *telegramRepo) BlockSecretMessagingSessionByID(ctx context.Context, sessID string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"sessID": sessID,
	})

	logger.Info("start blocking secret messaging session")

	if err := r.db.WithContext(ctx).Model(&model.SecretMessagingSession{}).Where("id = ?", sessID).Update("is_blocked", true).Error; err != nil {
		logger.WithError(err).Error("failed to block secret messaging session")
		return err
	}

	return nil
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

func (r *telegramRepo) CreateSecretMessagingSession(ctx context.Context, sess *model.SecretMessagingSession) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"session": utils.Dump(sess),
	})

	logger.Info("start creating secret messaging session")

	if err := r.db.WithContext(ctx).Create(sess).Error; err != nil {
		logger.WithError(err).Error("failed to create secret messaging session")
		return err
	}

	return nil
}

func (r *telegramRepo) CreateSecretMessagingMessageNode(ctx context.Context, msg *model.SecretMessageNode) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"msg": utils.Dump(msg),
	})

	logger.Info("start creating secret messaging message node")

	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		logger.WithError(err).Error("failed to create secret messaging message node")
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

func (r *telegramRepo) FindSecretMessagingSessionByID(ctx context.Context, id string) (*model.SecretMessagingSession, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"id":  id,
	})

	logger.Info("start finding secret messaging session by ID")

	sess := &model.SecretMessagingSession{}
	err := r.db.WithContext(ctx).Model(&model.SecretMessagingSession{}).Where("id = ?", id).Take(sess).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret messaging session by ID")
		return nil, err

	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound

	case nil:
		return sess, nil
	}
}

func (r *telegramRepo) FindSecretMessagingMessageNodeByID(ctx context.Context, msgID int64) (*model.SecretMessageNode, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"msgID": msgID,
	})

	logger.Info("start finding secret messaging message node by ID")

	msg := &model.SecretMessageNode{}
	err := r.db.WithContext(ctx).Model(&model.SecretMessageNode{}).Where("id = ?", msgID).Take(msg).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret messaging message node by ID")
		return nil, err

	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound

	case nil:
		return msg, nil
	}
}

func (r *telegramRepo) FindSecretMessagingSessionByUserID(ctx context.Context, senderID, targetID int64) (*model.SecretMessagingSession, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      helper.DumpContext(ctx),
		"senderID": senderID,
		"targetID": targetID,
	})

	logger.Info("start finding secret messaging node by user ID")

	sess := &model.SecretMessagingSession{}
	err := r.db.WithContext(ctx).Model(&model.SecretMessagingSession{}).Where("sender_id = ? AND target_id = ?", senderID, targetID).Take(sess).Error
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret messaging node by user ID")
		return nil, err

	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound

	case nil:
		return sess, nil
	}
}

func (r *telegramRepo) GetBlockerForSecretMessagingSessionToCache(ctx context.Context, key string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"key": key,
	})

	logger.Info("start getting blocker for secret messaging session to cache")

	_, err := r.cacher.Get(ctx, key)
	if err == nil {
		return errors.New("blocker for secret messaging session is detected")
	}

	return nil
}

func (r *telegramRepo) SetBlockerForSecretMessagingSessionToCache(ctx context.Context, key string, exp time.Duration) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"key": key,
		"exp": exp,
	})

	logger.Info("start setting blocker for secret messaging session to cache")

	// value here is not important and should not be read
	if err := r.cacher.Set(ctx, key, "true", exp); err != nil {
		logger.WithError(err).Error("failed to set blocker for secret messaging session to cache")
		return err
	}

	logger.Debug("blocker is set")

	return nil
}
