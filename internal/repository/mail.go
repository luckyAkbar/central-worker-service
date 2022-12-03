package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type mailRepo struct {
	db *gorm.DB
}

// NewMailRepository creates a new MailRepository
func NewMailRepository(db *gorm.DB) model.MailRepository {
	return &mailRepo{
		db,
	}
}

// Create create mail record
func (r *mailRepo) Create(ctx context.Context, mail *model.Mail) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"mail": utils.Dump(mail),
	})

	logger.Info("start to create mail record")

	if err := r.db.WithContext(ctx).Create(mail).Error; err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("mail record has been created")

	return nil
}

// Update update mail record
func (r *mailRepo) Update(ctx context.Context, mail *model.Mail) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"mail": utils.Dump(mail),
	})

	logger.Info("updating mail record")

	err := r.db.WithContext(ctx).Save(mail).Error
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("success updating mail record")

	return nil
}
