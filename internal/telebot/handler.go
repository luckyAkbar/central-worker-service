package telebot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type handler struct {
	dispatcher *ext.Dispatcher
}

// NewTelegramHandler create new telegram handler
func NewTelegramHandler(dispatcher *ext.Dispatcher) model.TelegramBot {
	return &handler{
		dispatcher,
	}
}

func (h *handler) RegisterHandlers() {
	h.dispatcher.AddHandler(handlers.NewCommand("start", h.startCommandHandler))
	h.dispatcher.AddHandler(handlers.NewMessage(message.Text, h.unknownCommandHandler))
}

func (h *handler) startCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		b,
		"Welcome to Central Service Telegram Bot!",
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.Message.MessageId,
		},
	)

	if err != nil {
		logrus.Error("failed to send reply to start command: ", err)
		return err
	}

	return nil
}

func (h *handler) unknownCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		b,
		"Sorry, the command / text is not known",
		nil,
	)

	if err != nil {
		logrus.Error("failed to send reply to start command: ", err)
		return err
	}

	return nil
}
