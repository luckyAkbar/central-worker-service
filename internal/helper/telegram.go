package helper

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
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
