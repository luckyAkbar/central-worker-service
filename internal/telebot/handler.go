package telebot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

var startMessage = `Welcome to Central Service Telegram Bot!

If you want to use secret messaging feature, you have to register first. Just type "/register" and sent that to me.
After you are registered, you can then start secretly messaging with the person you want!

To start secret messaging feature, all you have to do is type <strong>/secret [user-id]</strong>. The 'user-id' is the ID of the person you want to start secret messaging with.
`

type handler struct {
	dispatcher   *ext.Dispatcher
	teleUsecase  model.TelegramUsecase
	telegramRepo model.TelegramRepository
	workerClient model.WorkerClient
	diaryUsecase model.DiaryUsecase
}

// NewTelegramHandler create new telegram handler
func NewTelegramHandler(dispatcher *ext.Dispatcher, teleUsecase model.TelegramUsecase, telegramRepo model.TelegramRepository, workerClient model.WorkerClient, diaryUsecase model.DiaryUsecase) model.TelegramBot {
	return &handler{
		dispatcher,
		teleUsecase,
		telegramRepo,
		workerClient,
		diaryUsecase,
	}
}

func (h *handler) RegisterHandlers() {
	h.dispatcher.AddHandler(handlers.NewCommand("start", h.startCommandHandler))
	h.dispatcher.AddHandler(handlers.NewCommand("register", h.registerCommandHandler))
	h.dispatcher.AddHandler(handlers.NewCommand("secret", h.initiateSecretMessagingHandler))
	h.dispatcher.AddHandler(handlers.NewCommand("diary", h.createDiaryCommandHandler))
	h.dispatcher.AddHandler(handlers.NewCommand("find-diary", h.findDiaryCommandHandler))

	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(string(model.RegisterSecretMessagingService)), h.registerSecretTelegramMessagingCallbackHandler))
	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(string(model.DeleteDiaryPrefix)), h.handleDeleteDiaryByID))
	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(string(model.ReportSecretMessagePrefix)), h.handleReportSecretMessage))
	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(string(model.BlockSecretMessagingUserPrefix)), h.handleBlockSecretMessagingUser))

	h.dispatcher.AddHandler(handlers.NewMessage(message.Text, h.secretMessagingHandler))
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
							CallbackData: string(model.RegisterSecretMessagingService),
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
		startMessage,
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.Message.MessageId,
			ParseMode:        "html",
		},
	)

	if err != nil {
		logger.Error("failed to send reply to start command: ", err)
		return err
	}

	return nil
}

func (h *handler) unknownCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	return helper.TelegramEffectiveMessageReplier(
		b,
		ctx.EffectiveMessage,
		"Sorry, the command / text is not known",
		&gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
}
