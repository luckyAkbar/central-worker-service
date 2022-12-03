package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/client"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
)

type taskHandler struct {
	sibClient    *client.SIB
	workerClient model.WorkerClient
	mailRepo     model.MailRepository
}

func NewTaskHandler(sibClient *client.SIB, mailRepo model.MailRepository, workerClient model.WorkerClient) model.TaskHandler {
	return &taskHandler{
		sibClient:    sibClient,
		mailRepo:     mailRepo,
		workerClient: workerClient,
	}
}

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

	to, err := mail.SendInBlueTo()
	if err != nil {
		logger.Error(err)
		return err
	}

	cc, err := mail.SendInBlueCc()
	if err != nil {
		logger.Error(err)
		return err
	}

	bcc, err := mail.SendInBlueBcc()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("start to call sibClient send email function")

	result, err := th.sibClient.SendEmail(ctx, lib.SendSmtpEmail{
		Sender:      config.SendInBlueSender(),
		To:          to,
		Bcc:         bcc,
		Cc:          cc,
		HtmlContent: mail.HTMLContent,
		Subject:     mail.Subject,
	})

	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("received result from sibClient send email: ", utils.Dump(result))

	if err := th.workerClient.RegisterMailUpdatingTask(ctx, mail, model.PriorityHigh); err != nil {
		// if err here, just report and forget
		logger.Error(err)
	}

	logger.Info("task send email finished")

	return nil
}

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
