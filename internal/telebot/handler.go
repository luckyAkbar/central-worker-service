package telebot

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

type handler struct {
	dispatcher  *ext.Dispatcher
	teleUsecase model.TelegramUsecase
}

// NewTelegramHandler create new telegram handler
func NewTelegramHandler(dispatcher *ext.Dispatcher, teleUsecase model.TelegramUsecase) model.TelegramBot {
	return &handler{
		dispatcher,
		teleUsecase,
	}
}

func (h *handler) RegisterHandlers() {
	h.dispatcher.AddHandler(handlers.NewCommand("start", h.startCommandHandler))
	h.dispatcher.AddHandler(handlers.NewCommand("register", h.registerCommandHandler))
	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("register_secret_telegram_messaging"), h.registerSecretTelegramMessagingCallbackHandler))
	h.dispatcher.AddHandler(handlers.NewMessage(message.Text, h.unknownCommandHandler))
}

func (h *handler) registerSecretTelegramMessagingCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"message":       utils.Dump(ctx.Message),
		"telegram_user": utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start callback for register secret telegram messaging")

	cb := ctx.Update.CallbackQuery
	user := ctx.EffectiveUser
	if user.IsBot {
		logger.Info("this user is bot. not allowed to register")
		return helper.TelegramCallbackAnswerer(b, cb, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "sorry, user with status 'bot' is not allowed to register",
			ShowAlert: false,
		})
	}

	teleUser := &model.TelegramUser{
		ID:           user.Id,
		IsBot:        user.IsBot,
		FirstName:    user.FirstName,
		LastName:     null.NewString(user.LastName, true),
		Username:     null.NewString(user.Username, true),
		LanguageCode: null.NewString(user.LanguageCode, true),
		IsPremium:    null.NewBool(user.IsPremium, true),
	}

	ucErr := h.teleUsecase.RegisterSecretMessagingService(context.Background(), teleUser)
	switch ucErr.UnderlyingError {
	default:
		return helper.TelegramCallbackAnswerer(b, cb, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "sorry, you're unable to register. reason: " + ucErr.Message,
			ShowAlert: false,
		})

	case usecase.ErrAlreadyExists, nil:
		break
	}

	if err := helper.TelegramCallbackAnswerer(b, cb, &gotgbot.AnswerCallbackQueryOpts{
		Text:      "Success! Bot will send your text and you can share it to people to secretly have a chat with you through this bot.",
		ShowAlert: true,
	}); err != nil {
		logrus.Error(err)
		return err
	}

	return teleUser.SendMessageToThisUser(b, teleUser.GenerateShareSecretMessagingText(), &gotgbot.SendMessageOpts{
		AllowSendingWithoutReply: true,
	})
}

func (h *handler) registerCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"message": utils.Dump(ctx.Message),
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start handler register command")

	msg, err := ctx.EffectiveMessage.Reply(
		b,
		"Please click one of these buttons to select which service you want to register to",
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.Message.MessageId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						gotgbot.InlineKeyboardButton{
							Text:         "Secret Telegram Messaging",
							CallbackData: "register_secret_telegram_messaging",
						},
					},
				},
			},
		},
	)

	logger.Info("message: ", utils.Dump(msg))

	if err != nil {
		logger.WithError(err).Error("failed to send markup register")
		return err
	}

	return nil
}

func (h *handler) startCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"message": utils.Dump(ctx.Message),
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("starting command handler")

	_, err := ctx.EffectiveMessage.Reply(
		b,
		`Welcome to Central Service Telegram Bot!

If you want to use secret messaging feature, you have to register first. Just type "/register" and sent that to me.
After you are registered, you can then start secretly messaging with the person you want!`,
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.Message.MessageId,
		},
	)

	if err != nil {
		logger.Error("failed to send reply to start command: ", err)
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
