package worker

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type taskHandler struct {
	mailUtility  model.MailUtility
	workerClient model.WorkerClient
	mailRepo     model.MailRepository
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(mailUtility model.MailUtility, mailRepo model.MailRepository, workerClient model.WorkerClient) model.TaskHandler {
	return &taskHandler{
		mailUtility:  mailUtility,
		mailRepo:     mailRepo,
		workerClient: workerClient,
	}
}

// HandleMailingTask send email using sendinblue client. If success, will register
// task to update mail record
func (th *taskHandler) HandleMailingTask(ctx context.Context, t *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    t.Type(),
		"payload": string(t.Payload()),
	})

	logger.Info("starting to handle mailing task...")

	mail := &model.Mail{}
	if err := json.Unmarshal(t.Payload(), mail); err != nil {
		logger.Error(err)
		return err
	}

	res, client, err := th.mailUtility.SendEmail(ctx, mail)

	if err != nil {
		logger.Error(err)
		return err
	}

	metadata := &model.MailResultMetadata{
		Detail:    res,
		Signature: client,
	}

	logger.Info("received result from send email: ", metadata)

	mail.Status = model.MailStatusSuccess
	mail.Metadata = &sql.NullString{
		String: utils.Dump(metadata),
		Valid:  true,
	}

	if err := th.workerClient.RegisterMailUpdatingTask(ctx, mail, model.PriorityHigh); err != nil {
		// if err here, just report and forget
		logger.Error(err)
	}

	logger.Info("task send email finished")

	return nil
}

// HandleMailUpdatingTask handle mail updating task
func (th *taskHandler) HandleMailUpdatingTask(ctx context.Context, t *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    t.Type(),
		"payload": string(t.Payload()),
	})

	logger.Info("start handle mail updating task")

	mail := &model.Mail{}
	if err := json.Unmarshal(t.Payload(), mail); err != nil {
		logger.Error(err)
		return err
	}

	if err := th.mailRepo.Update(ctx, mail); err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("finish handle mail update")

	return nil
}
