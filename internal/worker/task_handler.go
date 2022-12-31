package worker

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type taskHandler struct {
	mailUtility  model.MailUtility
	workerClient model.WorkerClient
	mailRepo     model.MailRepository
	userRepo     model.UserRepository
	siakadRepo   model.SiakaduRepository
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(mailUtility model.MailUtility, mailRepo model.MailRepository, workerClient model.WorkerClient, userRepo model.UserRepository, siakadRepo model.SiakaduRepository) model.TaskHandler {
	return &taskHandler{
		mailUtility:  mailUtility,
		mailRepo:     mailRepo,
		workerClient: workerClient,
		userRepo:     userRepo,
		siakadRepo:   siakadRepo,
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

func (th *taskHandler) HandleUserActivationTask(ctx context.Context, t *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    t.Type(),
		"payload": string(t.Payload()),
	})

	logger.Info("start handle user activation task")

	var id string
	if err := json.Unmarshal(t.Payload(), &id); err != nil {
		logger.Error("failed to unmarshal user activation task: ", err)
		return err
	}

	if err := th.userRepo.ActivateByUserID(ctx, id); err != nil {
		logger.Error("task failed to activate user: ", err)
		return err
	}

	logger.Info("finish activating user")

	return nil
}

func (th *taskHandler) HandleSiakadProfilePictureTask(ctx context.Context, task *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    task.Type(),
		"payload": string(task.Payload()),
	})

	logger.Info("start handling siakad profile picture task")

	var npm string
	if err := json.Unmarshal(task.Payload(), &npm); err != nil {
		logger.WithError(err).Error("failed to unmarshal npm in siakad profile picture task")
		return err
	}

	scrapingResult, err := th.siakadRepo.FindByID(ctx, npm)
	switch err {
	default:
		logger.WithError(err).Error("failed to find scraping result from db")
		return err

	case nil:
		logger.Info("already found / scraped: ", utils.Dump(scrapingResult))
		return nil

	case repository.ErrNotFound:
		break
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	baseURL := "https://siakadu.unila.ac.id/uploads/fotomhs/"
	link := baseURL + npm

	res, err := client.Get(link)
	if err != nil {
		logger.WithError(err).Error("failed to get to the link: ", link)
		return err
	}

	defer helper.WrapCloser(res.Body.Close)

	if res.StatusCode == http.StatusNotFound {
		logger.Info("got 404 from link. Considering success: ", link)
		return nil
	}

	if res.StatusCode != http.StatusOK {
		logger.Error("got non 200 and non 404 response from: ", link)
		return errors.New("got non 200 and 404 response: " + link)
	}

	contentType := res.Header.Get("Content-Type")
	if err := helper.FilterImageMimetype(contentType); err != nil {
		logger.Error("received forbidden content type: ", contentType)
		return errors.New("received forbidden content type")
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err != nil {
		logger.WithError(err).Error("invalid content type: ", contentType)
		return err
	}

	filename := npm + exts[0]

	dst, err := os.Create(config.ImageMediaLocalStorage() + "/" + filename)
	if err != nil {
		logger.WithError(err).Error("failed to create destination image file")
		return err
	}

	written, err := io.Copy(dst, res.Body)
	if err != nil {
		logger.WithError(err).Error("failed to copy file source to destination")
		return err
	}

	result := &model.SiakaduScrapingResult{
		ID:        npm,
		CreatedAt: time.Now().UTC(),
		Filename:  filename,
		Location:  model.LocationLocal,
	}

	if err := th.siakadRepo.Create(ctx, result); err != nil {
		logger.WithError(err).Error("failed to write siakad scraping result to db")
		return err
	}

	logger.Info("success saving profile foto with written: ", written)

	return nil
}
