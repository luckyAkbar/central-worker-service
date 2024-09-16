package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type sessionRepo struct {
	db *gorm.DB
}

// NewSessionRepository create new session repository
func NewSessionRepository(db *gorm.DB) model.SessionRepository {
	return &sessionRepo{
		db,
	}
}

func (r *sessionRepo) Create(ctx context.Context, session *model.Session) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"session": utils.Dump(session),
	})

	logger.Info("start creating session")

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		logger.Error("failed to create session")
		return err
	}

	return nil
}

func (r *sessionRepo) FindByAccessToken(ctx context.Context, accessToken string) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":          helper.DumpContext(ctx),
		"access_token": accessToken,
	})

	logger.Info("start finding session by access token")

	session := &model.Session{}

	err := r.db.WithContext(ctx).Model(&model.Session{}).Where("access_token = ?", accessToken).Take(session).Error
	switch err {
	default:
		logger.Error("failed to find session: ", err)
		return nil, err

	case gorm.ErrRecordNotFound:
		logger.Info("session not found")
		return nil, ErrNotFound

	case nil:
		return session, nil
	}
}
