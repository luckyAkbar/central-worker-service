package telebot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
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
	h.dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("register_secret_telegram_messaging"), h.registerSecretTelegramMessagingCallbackHandler))
	h.dispatcher.AddHandler(handlers.NewMessage(message.Text, h.secretMessagingHandler))
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

func (h *handler) initiateSecretMessagingHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"message": utils.Dump(ctx.Message),
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	// TODO: ensure only one session active per target - sender id pair

	logger.Info("start initate secret messaging handler")

	_, err := h.telegramRepo.FindUserByID(context.Background(), ctx.EffectiveUser.Id)
	switch err {
	default:
		logger.WithError(err).Error("failed to find telegram user ID: ", ctx.EffectiveUser.Id)
		return helper.TelegramEffectiveMessageReplier(b, ctx.EffectiveMessage, "Sorry bot experiencing failure. Try again later", &gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		})

	case repository.ErrNotFound:
		return helper.TelegramEffectiveMessageReplier(b, ctx.EffectiveMessage, "To use this feature, you must register first. Just use \"/register\" command.", &gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		})

	case nil:
		break
	}

	_, args, err := helper.TelegramParseMessageCommandAndArgs(ctx.EffectiveMessage.Text)
	if err != nil {
		logger.Info("user sending invalid command and args formatted message")
		return helper.TelegramEffectiveMessageReplier(b, ctx.EffectiveMessage, err.Error(), &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
	}

	// only lookup the message after the command as the target user ID
	targetUserID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return helper.TelegramEffectiveMessageReplier(b, ctx.EffectiveMessage, "invalid user ID. please make sure the ID is correct", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
	}

	if targetUserID == ctx.EffectiveUser.Id {
		return helper.TelegramEffectiveMessageReplier(b, ctx.EffectiveMessage, "you can't use this feature to yourself", &gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
	}

	session, targetUser, ucErr := h.teleUsecase.InitateSecretMessagingSession(context.Background(), ctx.EffectiveUser.Id, targetUserID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(err).Error("failed to initiate secret messaging session")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing problem. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrNotFound:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Sorry there is a problem: %s", ucErr.Message),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		break
	}

	msg, err := ctx.EffectiveMessage.Reply(
		b,
		fmt.Sprintf("Success! Secret messaging session from you to %s is now active. To send your message secretly, you must reply to this message and after that, bot will forward it secretly to %s. Enjoy!", targetUser.FirstName, targetUser.FirstName),
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		},
	)

	if err != nil {
		logger.WithError(err).Error("failed to send success message in secret messaging")
		return err
	}

	if err := h.workerClient.RegisterSettingMessageNodeToSecretMessagingSessionTask(context.Background(), session.ID, msg); err != nil {
		logger.WithError(err).Error("failed to register setting message node to secret messaging session task")
		return err
	}

	return nil
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

// secretMessagingHandler will check if the ctx.EffectiveMessage.ReplyToMessageId is null or not
// if null, will be handled with unknown command handler
// otherwise, will be handled with secret messaging handler
func (h *handler) secretMessagingHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.ReplyToMessage == nil {
		return h.unknownCommandHandler(b, ctx)
	}

	logger := logrus.WithFields(logrus.Fields{
		"message": utils.Dump(ctx.Message),
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start secret messaging handler")

	// pertama, ambil msg reply to di ID nya
	// cari id itu termasuk di message node atau enggak
	// kalo ketemu, ambil session ID nya
	// cek dlu, itu bener ga punya nya si sender
	// kalo bener, ambiil data si target
	// dan terusin deh message text nya ke si target hehe
	repliedMsg := ctx.EffectiveMessage.ReplyToMessage

	msgNode, ucErr := h.teleUsecase.FindSecretMessageNodeByID(context.Background(), repliedMsg.MessageId)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find secret message node by ID")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing unexpected error. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrNotFound:
		logger.Info("dari sini")
		return h.unknownCommandHandler(b, ctx)

	case nil:
		break
	}

	logger.Info("ini msgNODE: ", utils.Dump(msgNode))

	session, ucErr := h.teleUsecase.FindSecretMessagingSessionByID(context.Background(), msgNode.SessionID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find secret messaging session by ID")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing unexpected error. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrNotFound:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, your secret messaging session is not found. Please initiate the session first",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		break
	}

	sender := ctx.EffectiveUser
	if !session.IsOwnedByID(sender.Id) {
		// untuk ngirim balik replies dari secret messaging service
		// pertama itu butuh session dan juga msgNode nya sama ambil juga effective msg nya
		// trus ntar create node dari effective msg, create node lagi untuk msg yg pas ngirim balik ke sender
		ucErr := h.teleUsecase.HandleReplyForSecretMessage(context.Background(), session, ctx.EffectiveMessage, msgNode)
		return ucErr.UnderlyingError
	}

	if session.IsExpired() {
		// TODO: increase expiry when secret messaging is replied
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, your secret messaging session is expired. Please re-initiate the session again",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)
	}

	_, ucErr = h.teleUsecase.FindUserByID(context.Background(), session.TargetID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find telegram user by ID")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing unexpected error. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrNotFound:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"bot couldn't find the user target of your secret message. Maybe you should invite them first?",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		break
	}

	ucErr = h.teleUsecase.SendSecretMessage(context.Background(), session, ctx.EffectiveMessage, msgNode)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to send secret message")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing unexpected error. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		return nil
	}
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
