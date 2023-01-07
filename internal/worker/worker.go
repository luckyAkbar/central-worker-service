package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

var mux = asynq.NewServeMux()

func registerTask(th model.TaskHandler) {
	mux.HandleFunc(string(model.TaskMailing), th.HandleMailingTask)
	mux.HandleFunc(string(model.TaskMailRecordUpdating), th.HandleMailUpdatingTask)
	mux.HandleFunc(string(model.TaskUserActivation), th.HandleUserActivationTask)
	mux.HandleFunc(string(model.TaskSiakadProfilePictureScraping), th.HandleSiakadProfilePictureTask)
	mux.HandleFunc(string(model.TaskSettingMessageNodeToSecretMessagingSession), th.HandleSettingMessageNodeToSecretMessagingSessionTask)
	mux.HandleFunc(string(model.TaskSendTelegramMessageToUser), th.HandleSendTelegramMessageToUserTask)
	mux.HandleFunc(string(model.TaskCreatingSecretMessagingMessageNode), th.HandleCreateSecretMessagingMessageNode)
}

type worker struct {
	client *asynq.Client
	server *asynq.Server
}

// NewClient create a new worker client
func NewClient(redisHost string) (model.WorkerClient, error) {
	redisOpts, err := asynq.ParseRedisURI(redisHost)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	client := asynq.NewClient(redisOpts)

	logrus.Info("worker client created")

	return &worker{
		client: client,
	}, nil
}

// NewServer creates a new worker server
func NewServer(redisHost string, th model.TaskHandler) (model.WorkerServer, error) {
	registerTask(th)

	redisOpts, err := asynq.ParseRedisURI(redisHost)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	client := asynq.NewClient(redisOpts)
	server := asynq.NewServer(
		redisOpts,
		asynq.Config{
			Concurrency: config.WorkerConcurrency(),
			Queues: map[string]int{
				string(model.PriorityHigh):    7,
				string(model.PriorityDefault): 2,
				string(model.PriorityLow):     1,
			},
			Logger:              logrus.WithField("source", "worker"),
			HealthCheckFunc:     healthCheck,
			HealthCheckInterval: time.Minute * 3,
			StrictPriority:      true,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Minute * time.Duration(n+1)
			},
		},
	)

	return &worker{
		client: client,
		server: server,
	}, nil
}

// Start start worker server
func (w *worker) Start() error {
	logrus.Info("starting worker...")
	if err := w.server.Run(mux); err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("worker running...")

	return nil
}

// Stop stop worker server
func (w *worker) Stop() {
	logrus.Info("stopping worker...")
	if w.client != nil {
		helper.WrapCloser(w.client.Close)
	}

	if w.server != nil {
		w.server.Stop()
	}

	logrus.Info("worker stopped.")
}

// RegisterMailingTask register a mailing task
func (w *worker) RegisterMailingTask(ctx context.Context, input *model.Mail, queue model.Priority) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	logger.Info("start to enqueue mailing task")

	payload, err := json.Marshal(input)
	if err != nil {
		logger.Error(err)
		return err
	}

	task := asynq.NewTask(
		string(model.TaskMailing),
		payload,
		asynq.MaxRetry(model.MailingTaskOption.MaxRetry),
		asynq.Timeout(model.MailingTaskOption.Timeout),
		asynq.Queue(string(queue)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("success enqueue mailing task. info: ", utils.Dump(info))

	return nil
}

// RegisterMailUpdatingTask register mail update task
func (w *worker) RegisterMailUpdatingTask(ctx context.Context, mail *model.Mail, queue model.Priority) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"mail": utils.Dump(mail),
	})

	logger.Info("start to enqueue mail updating task")

	payload, err := json.Marshal(mail)
	if err != nil {
		logger.Error(err)
		return err
	}

	task := asynq.NewTask(
		string(model.TaskMailRecordUpdating),
		payload,
		asynq.MaxRetry(model.MailUpdatingTaskOption.MaxRetry),
		asynq.Timeout(model.MailUpdatingTaskOption.Timeout),
		asynq.Queue(string(queue)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("success enqueue mail updating task. info: ", utils.Dump(info))

	return nil
}

func (w *worker) RegisterUserActivationTask(ctx context.Context, userID string) error {
	log := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"id":  userID,
	})

	log.Info("start usera activation task")

	payload, err := json.Marshal(userID)
	if err != nil {
		log.Error("failed to marshall in user activation task", err)
		return err
	}

	task := asynq.NewTask(
		string(model.TaskUserActivation),
		payload,
		asynq.MaxRetry(model.UserActivationTaskOption.MaxRetry),
		asynq.Timeout(model.UserActivationTaskOption.Timeout),
		asynq.Queue(string(model.PriorityHigh)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		log.Error("failed to enqueue task user activation: ", err)
		return err
	}

	log.Info("success enqueue task user activation: ", utils.Dump(info))

	return nil
}

func (w *worker) RegisterSiakadProfilePictureTask(ctx context.Context, npm string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"npm": npm,
	})

	logger.Info("registeting profile picture scraping task")

	payload, err := json.Marshal(npm)
	if err != nil {
		logger.WithError(err).Error("failed to marshal npm")
		return err
	}

	task := asynq.NewTask(
		string(model.TaskSiakadProfilePictureScraping),
		payload,
		asynq.MaxRetry(model.SiakadProfilePictureScraperTaskOption.MaxRetry),
		asynq.Timeout(model.SiakadProfilePictureScraperTaskOption.Timeout),
		asynq.Queue(string(model.PriorityHigh)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.WithError(err).Info("failed to enqueue task")
		return err
	}

	logrus.Info("successfully enqueued task: ", utils.Dump(info))

	return nil
}

func (w *worker) RegisterSettingMessageNodeToSecretMessagingSessionTask(ctx context.Context, sessID string, msg *gotgbot.Message) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"sessID": sessID,
		"msg":    utils.Dump(msg),
	})

	logger.Info("start registering sestting message node to secret messaging session task")

	payloadInput := &model.SettingMessageNodeToSecretMessagingSessionPayload{
		SessionID: sessID,
		Message:   msg,
	}

	payload, err := json.Marshal(payloadInput)
	if err != nil {
		logger.WithError(err).Error("failed to marshal message node to secret messaging session payload")
		return err
	}

	task := asynq.NewTask(
		string(model.TaskSettingMessageNodeToSecretMessagingSession),
		payload,
		asynq.MaxRetry(model.SettingMessageNodeToSecretMessagingSessionTaskOption.MaxRetry),
		asynq.Timeout(model.SettingMessageNodeToSecretMessagingSessionTaskOption.Timeout),
		asynq.Queue(string(model.PriorityHigh)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.WithError(err).Error("failed to enqueue task")
		return err
	}

	logger.Info("successfully enqueued task: ", utils.Dump(info))

	return nil
}

func (w *worker) RegisterSendingTelegramMessageToUser(ctx context.Context, input *model.SendTelegramMessageToUserPayload) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	logger.Info("start registering sending telegram message to user task")

	payload, err := json.Marshal(input)
	if err != nil {
		logger.WithError(err).Error("failed to marshal send telegram message to user payload")
		return err
	}

	task := asynq.NewTask(
		string(model.TaskSendTelegramMessageToUser),
		payload,
		asynq.MaxRetry(model.SendTelegramMessageToUserTaskOption.MaxRetry),
		asynq.Timeout(model.SendTelegramMessageToUserTaskOption.Timeout),
		asynq.Queue(string(model.PriorityHigh)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.WithError(err).Error("failed to enqueue task")
		return err
	}

	logger.Info("successfully enqueued task: ", utils.Dump(info))

	return nil
}

func (w *worker) RegisterCreateSecretMessagingMessageNode(ctx context.Context, node *model.SecretMessageNode) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"node": utils.Dump(node),
	})

	logger.Info("start registering create secret messaging message node task")

	payload, err := json.Marshal(node)
	if err != nil {
		logger.WithError(err).Error("failed to marshal create secret messaging message payload")
		return err
	}

	task := asynq.NewTask(
		string(model.TaskCreatingSecretMessagingMessageNode),
		payload,
		asynq.MaxRetry(model.CreateSecretMessagingMessageNodeOption.MaxRetry),
		asynq.Timeout(model.CreateSecretMessagingMessageNodeOption.Timeout),
		asynq.Queue(string(model.PriorityHigh)),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.WithError(err).Error("failed to enqueue task")
		return err
	}

	logger.Info("successfully enqueued task: ", utils.Dump(info))

	return nil
}
