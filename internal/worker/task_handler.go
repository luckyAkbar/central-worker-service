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

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

type taskHandler struct {
	mailUtility     model.MailUtility
	workerClient    model.WorkerClient
	mailRepo        model.MailRepository
	userRepo        model.UserRepository
	siakadRepo      model.SiakaduRepository
	telegramUsecase model.TelegramUsecase
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(mailUtility model.MailUtility, mailRepo model.MailRepository, workerClient model.WorkerClient, userRepo model.UserRepository, siakadRepo model.SiakaduRepository, telegramUsecase model.TelegramUsecase) model.TaskHandler {
	return &taskHandler{
		mailUtility:     mailUtility,
		mailRepo:        mailRepo,
		workerClient:    workerClient,
		userRepo:        userRepo,
		siakadRepo:      siakadRepo,
		telegramUsecase: telegramUsecase,
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

func (th *taskHandler) HandleSettingMessageNodeToSecretMessagingSessionTask(ctx context.Context, t *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    t.Type(),
		"payload": string(t.Payload()),
	})

	logger.Info("start handling setting message node to secret messaging session")

	payload := &model.SettingMessageNodeToSecretMessagingSessionPayload{}
	if err := json.Unmarshal(t.Payload(), payload); err != nil {
		logger.WithError(err).Error("failed to unmarshal setting message node to secret messaging session payload")
		return err
	}

	ucErr := th.telegramUsecase.SetMessageNodeToSecretMessagingSession(ctx, payload.SessionID, payload.Message)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to set message node to secret messaging session")
		return ucErr.UnderlyingError

	case usecase.ErrNotFound:
		logger.WithError(ucErr.UnderlyingError).Error("unexpectedly not found from SetMessageNodeToSecretMessagingSession function in worker")
		return ucErr.UnderlyingError

	case nil:
		logger.Info("success SetMessageNodeToSecretMessagingSession")
		return nil
	}
}

func (th *taskHandler) HandleSendTelegramMessageToUserTask(ctx context.Context, task *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    task.Type(),
		"payload": string(task.Payload()),
	})

	logger.Info("start handling send telegram message to user")

	payload := &model.SendTelegramMessageToUserPayload{}
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		logger.WithError(err).Error("failed to unmarshal task payload")
		return err
	}

	user, ucErr := th.telegramUsecase.FindUserByID(ctx, payload.UserID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find user by ID")
		return ucErr.UnderlyingError

	case usecase.ErrNotFound:
		logger.WithError(ucErr.UnderlyingError).Error("handling send telegram called but get not found in user")
		return ucErr.UnderlyingError

	case nil:
		break
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: "html",
	}

	if payload.ReplyToMessageID.Valid {
		opts.ReplyToMessageId = payload.ReplyToMessageID.Int64
	}

	msg, ucErr := th.telegramUsecase.SentTextMessageToUser(ctx, user.ID, payload.Message, opts)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to send text message to user")
		return ucErr.UnderlyingError

	case usecase.ErrNotFound:
		logger.WithError(ucErr.UnderlyingError).Error("failed to send text message to user because user is not found")
		return ucErr.UnderlyingError

	case nil:
		break
	}

	if err := th.workerClient.RegisterCreateSecretMessagingMessageNode(ctx, &model.SecretMessageNode{
		ID:                      msg.MessageId,
		SessionID:               payload.SessionID,
		CreatedAt:               time.Now().UTC(),
		Text:                    msg.Text,
		PreviousSecretMessageID: null.IntFrom(payload.MessageID),
	}); err != nil {
		logger.WithError(err).Error("failed to register create secret message node")
		return err
	}

	return nil
}

func (th *taskHandler) HandleCreateSecretMessagingMessageNode(ctx context.Context, t *asynq.Task) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"type":    t.Type(),
		"payload": string(t.Payload()),
	})

	logger.Info("start handling create secret messaging message node")

	node := &model.SecretMessageNode{}
	if err := json.Unmarshal(t.Payload(), node); err != nil {
		logger.WithError(err).Error("failed to unmarshal task payload")
		return err
	}

	if err := th.telegramUsecase.CreateSecretMessagingMessageNode(ctx, node); err.UnderlyingError != nil {
		logger.WithError(err.UnderlyingError).Error("failed to create secret messaging message node")
		return err.UnderlyingError
	}

	return nil
}
