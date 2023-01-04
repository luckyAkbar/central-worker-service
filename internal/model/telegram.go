package model

import (
	"context"
	"fmt"

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

// TelegramUsecase usecase for telegram
type TelegramUsecase interface {
	// RegisterSecretMessagingService will check is user already registered by it's ID
	// If already registered, returns err already exists.
	RegisterSecretMessagingService(ctx context.Context, teleUser *TelegramUser) UsecaseError
}

// TelegramRepository telegram repository
type TelegramRepository interface {
	CreateUser(ctx context.Context, user *TelegramUser) error
	FindUserByID(ctx context.Context, userID int64) (*TelegramUser, error)
}
