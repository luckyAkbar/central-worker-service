package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type mailUsecase struct {
	repo         model.MailRepository
	workerClient model.WorkerClient
}

// NewMailUsecase creates a new MailUsecase
func NewMailUsecase(repo model.MailRepository, workerClient model.WorkerClient) model.MailUsecase {
	return &mailUsecase{
		repo,
		workerClient,
	}
}

// Enqueue create mail record and send task to worker
func (u *mailUsecase) Enqueue(ctx context.Context, input *model.MailingInput) (*model.Mail, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	logger.Info("start enqueuing mail process")

	if err := input.Validate(); err != nil {
		logger.Info(err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         err.Error(),
		}
	}

	mail := &model.Mail{
		ID:          helper.GenerateID(),
		To:          utils.Dump(input.To),
		HTMLContent: input.HTMLContent,
		Subject:     input.Subject,
		CreatedAt:   time.Now().UTC(),
		Status:      model.MailStatusOnProgress,
		Cc: &sql.NullString{
			String: utils.Dump(input.Cc),
			Valid:  true,
		},
		Bcc: &sql.NullString{
			String: utils.Dump(input.Bcc),
			Valid:  true,
		},
	}

	logger.Info("saving to database...")

	if err := u.repo.Create(ctx, mail); err != nil {
		logger.Error(err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	logger.Info("registering task to mail worker...")

	if err := u.workerClient.RegisterMailingTask(ctx, mail, model.PriorityDefault); err != nil {
		logger.Error(err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgFailedRegisterTask,
		}
	}

	logger.Info("success create mailing task")

	return mail, model.NilUsecaseError
}
