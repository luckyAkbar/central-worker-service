package model

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/central-worker-service/internal/config"
)

// Task is worker task type
type Task string

// list of available tasks
var (
	TaskMailing            Task = "task:mailing"
	TaskMailRecordUpdating Task = "task:mailing:update_record"
	TaskUserActivation     Task = "task:user_activation"
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
)

// WorkerClient interface to enqueue task to worker
type WorkerClient interface {
	RegisterMailingTask(ctx context.Context, input *Mail, priority Priority) error
	RegisterMailUpdatingTask(ctx context.Context, mail *Mail, priority Priority) error
	RegisterUserActivationTask(ctx context.Context, id string) error
}

// TaskHandler worker task handler
type TaskHandler interface {
	HandleMailingTask(ctx context.Context, t *asynq.Task) error
	HandleMailUpdatingTask(ctx context.Context, t *asynq.Task) error
	HandleUserActivationTask(ctx context.Context, task *asynq.Task) error
}

// WorkerServer interface for worker server
type WorkerServer interface {
	Start() error
	Stop()
}
