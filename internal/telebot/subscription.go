package telebot

import (
	"context"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (h *handler) subscriptionHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"handler": "subscription",
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start subscription handler")

	err := helper.TelegramEffectiveMessageReplier(
		b,
		ctx.EffectiveMessage,
		"Please click one of these buttons to subscribe to any of our subscription services",
		&gotgbot.SendMessageOpts{
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						gotgbot.InlineKeyboardButton{
							Text:         "9Gag Meme",
							CallbackData: string(model.GagMemeServiceSubscription),
						},
					},
				},
			},
		},
	)

	if err != nil {
		logger.WithError(err).Error("failed to send subscription message")
		return err
	}

	return nil
}

func (h *handler) gagMemeSubscriptionHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"handler": "gagMemeSubscription",
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start gag meme subscription handler")

	cb := ctx.Update.CallbackQuery

	user, ucErr := h.teleUsecase.FindUserByID(context.Background(), ctx.EffectiveUser.Id)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find user")
		return helper.TelegramCallbackAnswerer(
			bot,
			cb,
			&gotgbot.AnswerCallbackQueryOpts{
				Text:      "Sorry, bot experiencing error. Please try again later",
				ShowAlert: true,
			},
		)

	case usecase.ErrNotFound:
		return helper.TelegramCallbackAnswerer(
			bot,
			cb,
			&gotgbot.AnswerCallbackQueryOpts{
				Text:      "Sorry, you are not registered yet. Please register first. Just type /register and sent it to me",
				ShowAlert: true,
			},
		)

	case nil:
		break
	}

	subs := &model.Subscription{
		ID:              helper.GenerateID(),
		Type:            model.SubscriptionTypeMeme,
		Channel:         model.SubscriptionChannelTelegram,
		UserReferenceID: fmt.Sprintf("%d", user.ID),
	}

	_, ucErr = h.subscriptionUsecase.FindSubscription(context.Background(), subs.Type, subs.Channel, subs.UserReferenceID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to create subscription")
		return helper.TelegramCallbackAnswerer(
			bot,
			cb,
			&gotgbot.AnswerCallbackQueryOpts{
				Text:      "Sorry, bot experiencing error. Please try again later",
				ShowAlert: true,
			},
		)

	case usecase.ErrNotFound:
		if ucErr := h.subscriptionUsecase.Create(context.Background(), subs); ucErr.UnderlyingError != nil {
			logger.WithError(ucErr.UnderlyingError).Error("failed to create subscription")
			return helper.TelegramCallbackAnswerer(
				bot,
				cb,
				&gotgbot.AnswerCallbackQueryOpts{
					Text:      "Sorry, bot experiencing error. Please try again later",
					ShowAlert: true,
				},
			)
		}
	case nil:
		break
	}

	return helper.TelegramCallbackAnswerer(
		bot,
		cb,
		&gotgbot.AnswerCallbackQueryOpts{
			Text:      "You have successfully subscribed to Meme subscription. Bot will periodically sent you funny meme!",
			ShowAlert: true,
		},
	)
}
