package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

type telegramUsecase struct {
	telegramRepo model.TelegramRepository
	telebot      *gotgbot.Bot
	workerClient model.WorkerClient
}

// NewTelegramUsecase create a new telegram usecase
func NewTelegramUsecase(telegramRepo model.TelegramRepository, telebot *gotgbot.Bot, workerClient model.WorkerClient) model.TelegramUsecase {
	return &telegramUsecase{
		telegramRepo,
		telebot,
		workerClient,
	}
}

func (u *telegramUsecase) RegisterSecretMessagingService(ctx context.Context, teleUser *model.TelegramUser) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":           helper.DumpContext(ctx),
		"telegram_user": utils.Dump(teleUser),
	})

	logger.Info("start register secret messaging service usecase")

	_, err := u.telegramRepo.FindUserByID(ctx, teleUser.ID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by ID")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case nil:
		return model.UsecaseError{
			UnderlyingError: ErrAlreadyExists,
			Message:         "user already registered",
		}

	case repository.ErrNotFound:
		break
	}

	logger.Info("saving telegram user to repository...")

	if err := u.telegramRepo.CreateUser(ctx, teleUser); err != nil {
		logger.WithError(err).Error("failed to register user to telegram user repo")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}

func (u *telegramUsecase) InitateSecretMessagingSession(ctx context.Context, initiatorID, targetID int64) (*model.SecretMessagingSession, *model.TelegramUser, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":         helper.DumpContext(ctx),
		"initiatorID": initiatorID,
		"targetID":    targetID,
	})

	logger.Info("start initiating secret messaging session")

	initiator, err := u.telegramRepo.FindUserByID(context.Background(), initiatorID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find telegram user ID: ", initiatorID)
		return nil, nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         "To use this issue, you must register first.",
		}

	case nil:
		break
	}

	targetUser, err := u.telegramRepo.FindUserByID(context.Background(), targetID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find user")
		return nil, nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         fmt.Sprintf("User with ID: %d is not found. Please ask them to register this feature first.", targetID),
		}

	case nil:
		break
	}

	session := &model.SecretMessagingSession{
		ID:        helper.GenerateID(),
		SenderID:  initiator.ID,
		TargetID:  targetUser.ID,
		CreatedAt: time.Now().UTC(),
		ExpiredAt: time.Now().Add(config.TelegramBotSecretMessagingSessionExpiryHour()).UTC(),
	}

	if err := u.telegramRepo.CreateSecretMessagingSession(context.Background(), session); err != nil {
		logger.WithError(err).Error("failed to create secret messaging session")
		return nil, nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return session, targetUser, model.NilUsecaseError
}

func (u *telegramUsecase) SetMessageNodeToSecretMessagingSession(ctx context.Context, sessID string, msg *gotgbot.Message) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.Dump(ctx),
		"sessID": sessID,
		"msg":    utils.Dump(msg),
	})

	logger.Info("start setting message node to secret messaging session")

	sess, err := u.telegramRepo.FindSecretMessagingSessionByID(ctx, sessID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret messaging session")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	node := &model.SecretMessageNode{
		ID:        msg.MessageId,
		SessionID: sess.ID,
		CreatedAt: time.Now().UTC(),
		Text:      msg.Text,
	}

	if err := u.telegramRepo.CreateSecretMessagingMessageNode(ctx, node); err != nil {
		logger.WithError(err).Error("failed to create secret message node")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}

func (u *telegramUsecase) FindSecretMessageNodeByID(ctx context.Context, msgID int64) (*model.SecretMessageNode, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"msgID": msgID,
	})

	logger.Info("start finding secret message node by ID")

	msgNode, err := u.telegramRepo.FindSecretMessagingMessageNodeByID(ctx, msgID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret message node")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		return msgNode, model.NilUsecaseError
	}
}

func (u *telegramUsecase) FindSecretMessagingSessionByID(ctx context.Context, sessID string) (*model.SecretMessagingSession, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"sessID": sessID,
	})

	logger.Info("start finding secret messaging session by ID")

	session, err := u.telegramRepo.FindSecretMessagingSessionByID(ctx, sessID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find secret messaging session by ID")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		return session, model.NilUsecaseError
	}
}

func (u *telegramUsecase) FindUserByID(ctx context.Context, id int64) (*model.TelegramUser, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    helper.DumpContext(ctx),
		"userID": id,
	})

	logger.Info("start finding telegram user by ID")

	user, err := u.telegramRepo.FindUserByID(ctx, id)
	switch err {
	default:
		logger.WithError(err).Error("failed to find telegram user by ID")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		return user, model.NilUsecaseError
	}
}

func (u *telegramUsecase) SendSecretMessage(ctx context.Context, sms *model.SecretMessagingSession, secretMsg *gotgbot.Message, parentMsgNode *model.SecretMessageNode) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":           helper.DumpContext(ctx),
		"sms":           utils.Dump(sms),
		"secretMsg":     utils.Dump(secretMsg),
		"parentMsgNode": utils.Dump(parentMsgNode),
	})

	logger.Info("start sending secret message")

	msgNode := &model.SecretMessageNode{
		ID:                      secretMsg.MessageId,
		SessionID:               sms.ID,
		CreatedAt:               time.Now().UTC(),
		Text:                    secretMsg.Text,
		PreviousSecretMessageID: null.IntFrom(parentMsgNode.ID),
	}

	if err := u.telegramRepo.CreateSecretMessagingMessageNode(ctx, msgNode); err != nil {
		logger.WithError(err).Error("failed to create secret message node")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	payload := &model.SendTelegramMessageToUserPayload{
		UserID:           sms.TargetID,
		Message:          helper.WrapSecretMessageText(secretMsg.Text),
		MessageID:        secretMsg.MessageId,
		ReplyToMessageID: parentMsgNode.PreviousSecretMessageID,
		SessionID:        sms.ID,
	}

	if err := u.workerClient.RegisterSendingTelegramMessageToUser(ctx, payload); err != nil {
		logger.WithError(err).Error("failed to register sending telegram message to user")
		return model.UsecaseError{
			UnderlyingError: err,
			Message:         MsgInternalError,
		}

	}

	return model.NilUsecaseError
}

func (u *telegramUsecase) HandleReplyForSecretMessage(ctx context.Context, sms *model.SecretMessagingSession, replyMsg *gotgbot.Message, parentMsgNode *model.SecretMessageNode) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":           helper.DumpContext(ctx),
		"sms":           utils.Dump(sms),
		"replyMsg":      utils.Dump(replyMsg),
		"parentMsgNode": utils.Dump(parentMsgNode),
	})

	logger.Info("start sending secret message")

	replier, err := u.telegramRepo.FindUserByID(ctx, sms.TargetID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find user by ID")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	msgNode := &model.SecretMessageNode{
		ID:                      replyMsg.MessageId,
		SessionID:               sms.ID,
		CreatedAt:               time.Now().UTC(),
		Text:                    replyMsg.Text,
		PreviousSecretMessageID: null.IntFrom(parentMsgNode.ID),
	}

	if err := u.telegramRepo.CreateSecretMessagingMessageNode(ctx, msgNode); err != nil {
		logger.WithError(err).Error("failed to create secret message node")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	payload := &model.SendTelegramMessageToUserPayload{
		UserID:           sms.SenderID,
		Message:          helper.WrapRepliedSecretMessageText(replyMsg.Text, replier.FirstName),
		MessageID:        replyMsg.MessageId,
		ReplyToMessageID: parentMsgNode.PreviousSecretMessageID,
		SessionID:        sms.ID,
	}

	if err := u.workerClient.RegisterSendingTelegramMessageToUser(ctx, payload); err != nil {
		logger.WithError(err).Error("failed to register sending telegram message to user")
		return model.UsecaseError{
			UnderlyingError: err,
			Message:         MsgInternalError,
		}

	}

	return model.NilUsecaseError
}

func (u *telegramUsecase) SentTextMessageToUser(ctx context.Context, userID int64, message string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, model.UsecaseError) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     helper.DumpContext(ctx),
		"userID":  userID,
		"message": message,
	})

	logger.Info("start sending text message to user")

	user, err := u.telegramRepo.FindUserByID(ctx, userID)
	switch err {
	default:
		logger.WithError(err).Error("failed to find telegram user by ID")
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case repository.ErrNotFound:
		return nil, model.UsecaseError{
			UnderlyingError: ErrNotFound,
			Message:         MsgNotFound,
		}

	case nil:
		break
	}

	chat := &gotgbot.Chat{
		Id: user.ID,
	}

	msg, err := chat.SendMessage(u.telebot, message, opts)
	if err != nil {
		logger.WithError(err).Error("failed to send text message to user")
		return nil, model.UsecaseError{
			UnderlyingError: err,
			Message:         MsgInternalError,
		}
	}

	return msg, model.NilUsecaseError
}

func (u *telegramUsecase) CreateSecretMessagingMessageNode(ctx context.Context, node *model.SecretMessageNode) model.UsecaseError {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"node": utils.Dump(node),
	})

	logger.Info("start creating secret messaging message node")

	if err := u.telegramRepo.CreateSecretMessagingMessageNode(ctx, node); err != nil {
		logger.WithError(err).Error("failed to create secret messaging message node")
		return model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	return model.NilUsecaseError
}
