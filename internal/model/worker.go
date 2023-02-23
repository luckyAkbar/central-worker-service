package model

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"gopkg.in/guregu/null.v4"
)

// Task is worker task type
type Task string

// list of available tasks
var (
	TaskMailing                                    Task = "task:mailing"
	TaskMailRecordUpdating                         Task = "task:mailing:update_record"
	TaskUserActivation                             Task = "task:user_activation"
	TaskSiakadProfilePictureScraping               Task = "task:siakad_profile_picture_scraping"
	TaskSettingMessageNodeToSecretMessagingSession Task = "task:setting_message_node_to_secret_messaging_session"
	TaskSendTelegramMessageToUser                  Task = "task:send_telegram_message_to_user"
	TaskCreatingSecretMessagingMessageNode         Task = "task:creating_secret_message_node"
	TaskMemeSubscription                           Task = "task:meme_subscription"
)

// Priority is worker priority
type Priority string

// list worker priority
var (
	PriorityHigh    Priority = "high"
	PriorityDefault Priority = "default"
	PriorityLow     Priority = "low"
)

// TaskOption is worker task option
type TaskOption struct {
	MaxRetry int
	Timeout  time.Duration
}

// defined option for each task. any new task option must be defined here
var (
	MailingTaskOption = &TaskOption{
		MaxRetry: config.MailingTaskMaxRetry(),
		Timeout:  config.MailingTaskTimeoutSeconds(),
	}

	MailUpdatingTaskOption = &TaskOption{
		MaxRetry: config.MailUpdatingTaskMaxRetry(),
		Timeout:  config.MailUpdatingTaskTimeoutSeconds(),
	}

	UserActivationTaskOption = &TaskOption{
		MaxRetry: config.UserActivationTaskMaxRetry(),
		Timeout:  config.UserActivationTaskTimeoutSeconds(),
	}

	SiakadProfilePictureScraperTaskOption = &TaskOption{
		MaxRetry: 10,
		Timeout:  time.Second * 20,
	}

	SettingMessageNodeToSecretMessagingSessionTaskOption = &TaskOption{
		MaxRetry: config.SettingMessageNodeToSecretMessagingSessionMaxRetry(),
		Timeout:  config.SettingMessageNodeToSecretMessagingSessionTimeoutSeconds(),
	}

	SendTelegramMessageToUserTaskOption = &TaskOption{
		MaxRetry: config.SendTelegramMessageToUserMaxRetry(),
		Timeout:  config.SendTelegramMessageToUserTimeoutSeconds(),
	}

	CreateSecretMessagingMessageNodeOption = &TaskOption{
		MaxRetry: 100,
		Timeout:  time.Second * 10,
	}

	MemeSubscriptionTaskOption = &TaskOption{
		MaxRetry: config.MemeSubscriptionMaxRetry(),
		Timeout:  config.MemeSubscriptionTimeoutSeconds(),
	}
)

// SettingMessageNodeToSecretMessagingSessionPayload payload
type SettingMessageNodeToSecretMessagingSessionPayload struct {
	SessionID string
	Message   *gotgbot.Message
}

// SendTelegramMessageToUserPayload payload
type SendTelegramMessageToUserPayload struct {
	UserID               int64
	Message              string
	MessageID            int64
	ReplyToMessageID     null.Int
	SessionID            string
	ParseMode            string
	InlineKeybordButtons []gotgbot.InlineKeyboardButton
}

// WorkerClient interface to enqueue task to worker
type WorkerClient interface {
	RegisterMailingTask(ctx context.Context, input *Mail, priority Priority) error
	RegisterMailUpdatingTask(ctx context.Context, mail *Mail, priority Priority) error
	RegisterUserActivationTask(ctx context.Context, id string) error
	RegisterSiakadProfilePictureTask(ctx context.Context, npm string) error
	RegisterSettingMessageNodeToSecretMessagingSessionTask(ctx context.Context, sessID string, msg *gotgbot.Message) error
	RegisterSendingTelegramMessageToUser(ctx context.Context, payload *SendTelegramMessageToUserPayload) error
	RegisterCreateSecretMessagingMessageNode(ctx context.Context, node *SecretMessageNode) error
}

// TaskHandler worker task handler
type TaskHandler interface {
	HandleMailingTask(ctx context.Context, t *asynq.Task) error
	HandleMailUpdatingTask(ctx context.Context, t *asynq.Task) error
	HandleUserActivationTask(ctx context.Context, task *asynq.Task) error
	HandleSiakadProfilePictureTask(ctx context.Context, task *asynq.Task) error
	HandleSettingMessageNodeToSecretMessagingSessionTask(ctx context.Context, task *asynq.Task) error
	HandleSendTelegramMessageToUserTask(ctx context.Context, task *asynq.Task) error
	HandleCreateSecretMessagingMessageNode(ctx context.Context, task *asynq.Task) error
	HandleMemeSubscriptionTask(ctx context.Context, task *asynq.Task) error
}

// WorkerServer interface for worker server
type WorkerServer interface {
	Start() error
	Stop()
}
