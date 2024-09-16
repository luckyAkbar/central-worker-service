package helper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/kumparan/go-utils"
	"github.com/sirupsen/logrus"
)

// TelegramCallbackAnswerer helper to answer telegram callback. If err accoures, it will call logrus.Error
// any error accoured will be retuned
func TelegramCallbackAnswerer(bot *gotgbot.Bot, cb *gotgbot.CallbackQuery, opts *gotgbot.AnswerCallbackQueryOpts) error {
	_, err := cb.Answer(
		bot,
		opts,
	)

	if err != nil {
		logrus.WithError(err).Error("failed to answer telegram callback")
		return err
	}

	return nil
}

// TelegramEffectiveMessageReplier wrapper for replying to effective message
// any error is returned, and the sent message will be logged
func TelegramEffectiveMessageReplier(bot *gotgbot.Bot, msg *gotgbot.Message, text string, opts *gotgbot.SendMessageOpts) error {
	res, err := msg.Reply(bot, text, opts)
	if err != nil {
		logrus.WithError(err).Error("failed to send message reply")
		return err
	}

	logrus.Info("message sent from replier: ", utils.Dump(res))

	return nil
}

// TelegramParseMessageCommandAndArgs parsing the command and the argument
func TelegramParseMessageCommandAndArgs(msg string) (string, []string, error) {
	if !strings.HasPrefix(msg, "/") {
		return "", []string{}, errors.New("not a valid command")
	}

	parts := strings.Split(msg, " ")
	switch len(parts) {
	default:
		return parts[0], parts[1:], nil

	case 0:
		return "", []string{}, errors.New("not a valid command")

	case 1:
		return parts[0], []string{}, nil
	}
}

// WrapSecretMessageText will wrap real with formatted: <strong>Someone secretly said</strong>: real
func WrapSecretMessageText(real string) string {
	return fmt.Sprintf("<strong>Someone secretly said</strong>: %s", real)
}

// WrapRepliedSecretMessageText will return real with formatted message: <strong>replierName replies</strong>
func WrapRepliedSecretMessageText(real string, replierName string) string {
	return fmt.Sprintf("<strong>%s replies</strong>: %s", replierName, real)
}
