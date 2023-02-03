package model

import (
	"context"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

// TelegramUser represent database table and also represent telegram user information
type TelegramUser struct {
	ID           int64       `json:"id"`
	IsBot        bool        `json:"is_bot"`
	FirstName    string      `json:"first_name"`
	LastName     null.String `json:"last_name,omitempty"`
	Username     null.String `json:"username,omitempty"`
	LanguageCode null.String `json:"language_code,omitempty"`
	IsPremium    null.Bool   `json:"is_premium,omitempty"`
}

// GenerateShareSecretMessagingText returns a text for user to share on their social media
// to announce their're ready to have secret messaging
func (tu *TelegramUser) GenerateShareSecretMessagingText() string {
	return fmt.Sprintf("Hello, I'm %s. If you want to secretly have chat with me in Telegram without me knowing who you are, you can register from this bot: %s and my code is: %d. Can't wait to have chat with you!", tu.FirstName, config.TelegramBotStartLink(), tu.ID)
}

// SendMessageToThisUser helper function to send message to private chat
// be carefull, must only to the already registered users
func (tu *TelegramUser) SendMessageToThisUser(bot *gotgbot.Bot, text string, opts *gotgbot.SendMessageOpts) error {
	logger := logrus.WithFields(logrus.Fields{
		"user": utils.Dump(tu),
	})

	logger.Info("sending message to telegram user")

	chat := &gotgbot.Chat{
		Id:       tu.ID,
		Username: tu.Username.String,
	}

	msg, err := chat.SendMessage(bot, text, opts)

	if err != nil {
		logrus.WithError(err).Error("failed to send chat to user")
		return err
	}

	logger.Info("message sent: ", utils.Dump(msg))

	return nil
}

// SecretMessagingSession session created to indicate that the Sender is estabilished a
// conversation and bot will secretly forwarding the message to target
type SecretMessagingSession struct {
	ID        string    `json:"id"`
	SenderID  int64     `json:"sender_id"`
	TargetID  int64     `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
	IsBlocked bool      `json:"is_blocked"`
}

// IsExpired compared now in utc against sms.ExpiredAt
func (sms *SecretMessagingSession) IsExpired() bool {
	return time.Now().UTC().After(sms.ExpiredAt)
}

// IsOwnedByID check is sms.SenderID is same with id
func (sms *SecretMessagingSession) IsOwnedByID(id int64) bool {
	return sms.SenderID == id
}

// SecretMessageNode secret message node
type SecretMessageNode struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	Text      string    `json:"text"`

	// to indicate this message was sent for which message. must be FK to this table
	PreviousSecretMessageID null.Int `json:"previous_secret_message_id"`
}

// CreateCacheKeyForBlockedSecretMessagingSessionUser create cache key for blocked secret messaging session in cache
func CreateCacheKeyForBlockedSecretMessagingSessionUser(senderID, targetID int64) string {
	return fmt.Sprintf("blocked_secret_messaging_session_%d_%d", senderID, targetID)
}

// TelegramUsecase usecase for telegram
type TelegramUsecase interface {
	// RegisterSecretMessagingService will check is user already registered by it's ID
	// If already registered, returns err already exists.
	RegisterSecretMessagingService(ctx context.Context, teleUser *TelegramUser) UsecaseError

	ReportSecretMessage(ctx context.Context, msgID int64) UsecaseError

	InitateSecretMessagingSession(ctx context.Context, senderID, targetID int64) (*SecretMessagingSession, *TelegramUser, UsecaseError)

	BlockSecretMessagingSession(ctx context.Context, sess *SecretMessagingSession) UsecaseError

	CreateSecretMessagingMessageNode(ctx context.Context, node *SecretMessageNode) UsecaseError

	SetMessageNodeToSecretMessagingSession(ctx context.Context, sessID string, msgNode *gotgbot.Message) UsecaseError

	SendSecretMessage(ctx context.Context, sms *SecretMessagingSession, secretMsg *gotgbot.Message, parentMsgNode *SecretMessageNode) UsecaseError

	HandleReplyForSecretMessage(ctx context.Context, session *SecretMessagingSession, replyMsg *gotgbot.Message, parentMsgNode *SecretMessageNode) UsecaseError

	SentTextMessageToUser(ctx context.Context, userID int64, message string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, UsecaseError)

	FindSecretMessageNodeByID(ctx context.Context, msgID int64) (*SecretMessageNode, UsecaseError)

	FindSecretMessagingSessionByID(ctx context.Context, sessID string) (*SecretMessagingSession, UsecaseError)

	FindUserByID(ctx context.Context, id int64) (*TelegramUser, UsecaseError)

	GetSecretMessagingSession(ctx context.Context, senderID, targetID int64) (*SecretMessagingSession, UsecaseError)
}

// TelegramRepository telegram repository
type TelegramRepository interface {
	BlockSecretMessagingSessionByID(ctx context.Context, sessionID string) error
	CreateUser(ctx context.Context, user *TelegramUser) error
	CreateSecretMessagingSession(ctx context.Context, sess *SecretMessagingSession) error
	CreateSecretMessagingMessageNode(ctx context.Context, msg *SecretMessageNode) error
	FindUserByID(ctx context.Context, userID int64) (*TelegramUser, error)
	FindSecretMessagingSessionByID(ctx context.Context, sessionID string) (*SecretMessagingSession, error)
	FindSecretMessagingMessageNodeByID(ctx context.Context, msgID int64) (*SecretMessageNode, error)
	FindSecretMessagingSessionByUserID(ctx context.Context, senderID, targetID int64) (*SecretMessagingSession, error)
	GetBlockerForSecretMessagingSessionToCache(ctx context.Context, key string) error
	SetBlockerForSecretMessagingSessionToCache(ctx context.Context, key string, exp time.Duration) error
}
