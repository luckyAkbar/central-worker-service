// Package telebot hold handler for interaction with telegram bot
package telebot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

var findDiaryHelpMessage = `Sorry, if you want to find your diary, you need to provide the arguments. The correct command format are 
<strong>/find-diary id diary-id</strong>
<strong>/find-diary date YYYY-MM-DD</strong>
<strong>/find-diary date today</strong>
<strong>/find-diary date yesterday</strong>
<strong>/find-diary range YYYY-MM-DD YYYY-MM-DD</strong>
<strong>/find-diary range yesterday today</strong>
Please make sure you type the command correctly`

func (h *handler) createDiaryCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"command": "create_diary",
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start create diary command handler")

	_, notes, err := helper.TelegramParseMessageCommandAndArgs(ctx.EffectiveMessage.Text)
	if err != nil || len(notes) == 0 || notes[0] == "" {
		logger.WithError(err).Error("invalid command & args for create diary from telegram")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, this command is invalid to create diary. The correct command format is <strong>/diary just type all your diary here, all the value will be read after the diary command </strong>",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}

	// TODO: get timezone from user
	defaultTimeZone := "Asia/Jakarta"

	createDiaryInput := &model.CreateDiaryInput{
		OwnerID:  strconv.FormatInt(ctx.EffectiveUser.Id, 10),
		Note:     strings.Join(notes, " "),
		TimeZone: defaultTimeZone,
		Source:   string(model.DiarySourceTelegram),

		// TODO: make sure the time here is created based also from user's timezone
		CreatedAt: time.Now(),
	}

	diary, ucErr := h.diaryUsecase.Create(context.Background(), createDiaryInput)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(err).Error("failed to create diary")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing problems. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		logger.WithField("created diary: ", utils.Dump(diary)).Info("successfully created diary")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprint("Diary saved. ID: ", diary.ID),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)
	}
}

func (h *handler) findDiaryCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	logger := logrus.WithFields(logrus.Fields{
		"command": "find_diary",
		"user":    utils.Dump(ctx.EffectiveUser),
	})

	logger.Info("start find diary command handler")

	_, args, err := helper.TelegramParseMessageCommandAndArgs(ctx.EffectiveMessage.Text)
	if err != nil || len(args) == 0 || args[0] == "" {
		logger.WithError(err).Error("invalid command & args for find diary from telegram")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			findDiaryHelpMessage,
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}

	if len(args) < 2 {
		logger.Info("invalid number of arguments to find diary")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			findDiaryHelpMessage,
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}

	findDiaryMethod := strings.ToUpper(args[0])

	switch findDiaryMethod {
	default:
		logger.Info("invalid command & args for find diary from telegram")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Sorry, method: <i>%s</i> is not a valid method to find your diary. Some of the correct values are <strong>id, date</strong>", args[0]),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)

	case "ID":
		diaryID := args[1]
		return h.handleFindDiaryByID(b, ctx, diaryID)

	case "DATE":
		date := args[1]
		return h.handleFindDiaryByDateRange(b, ctx, date, date)

	case "RANGE":
		if len(args) < 3 {
			return helper.TelegramEffectiveMessageReplier(
				b,
				ctx.EffectiveMessage,
				findDiaryHelpMessage,
				&gotgbot.SendMessageOpts{
					ReplyToMessageId: ctx.EffectiveMessage.MessageId,
					ParseMode:        "html",
				},
			)
		}

		startDate := args[1]
		endDate := args[2]
		return h.handleFindDiaryByDateRange(b, ctx, startDate, endDate)
	}
}

func (h *handler) handleFindDiaryByID(b *gotgbot.Bot, ctx *ext.Context, id string) error {
	logger := logrus.WithFields(logrus.Fields{
		"command": "find_diary_by_id",
		"user":    utils.Dump(ctx.EffectiveUser),
		"msg":     utils.Dump(ctx.EffectiveMessage),
	})

	userID := strconv.FormatInt(ctx.EffectiveUser.Id, 10)
	diary, ucErr := h.diaryUsecase.GetDiaryByID(context.Background(), id, userID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find diary by id")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, bot experiencing problems. Please try again later",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrForbidden:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			"Sorry, you're not the owner of this diary",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case usecase.ErrNotFound:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Sorry, diary with id: <strong>%s</strong> is not found", id),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)

	case nil:
		logger.WithField("found diary: ", utils.Dump(diary)).Info("successfully found diary")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("found: <strong>%s</strong>", diary.Note),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}
}

func (h *handler) handleFindDiaryByDateRange(b *gotgbot.Bot, ctx *ext.Context, startDate, endDate string) error {
	logger := logrus.WithFields(logrus.Fields{
		"command": "find_diary_by_date",
		"user":    utils.Dump(ctx.EffectiveUser),
		"msg":     utils.Dump(ctx.EffectiveMessage),
	})

	// TODO: ensure timezone is correct
	// TODO: detect maximum character length in one message
	start, _, err := helper.GenerateStartAndEndDate(startDate)
	if err != nil {
		logger.WithError(err).Info("failed to parse date")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Sorry, your date format: %s is invalid. Please use format: <strong>YYYY-MM-DD</strong>", startDate),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}

	_, end, err := helper.GenerateStartAndEndDate(endDate)
	if err != nil {
		logger.WithError(err).Info("failed to parse date")
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Sorry, your date format: %s is invalid. Please use format: <strong>YYYY-MM-DD</strong>", endDate),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}

	userID := strconv.FormatInt(ctx.EffectiveUser.Id, 10)
	diaries, ucErr := h.diaryUsecase.GetDiariesByWrittenDateRange(context.Background(), start, end, userID)
	switch ucErr.UnderlyingError {
	default:
		logger.WithError(ucErr.UnderlyingError).Error("failed to find diary by written date")
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
			"Sorry, bot could't find the diary you want :(",
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			},
		)

	case nil:
		return helper.TelegramEffectiveMessageReplier(
			b,
			ctx.EffectiveMessage,
			fmt.Sprintf("Diary found: \n%s", helper.FlattenAndFormatDiaries(diaries)),
			&gotgbot.SendMessageOpts{
				ReplyToMessageId: ctx.EffectiveMessage.MessageId,
				ParseMode:        "html",
			},
		)
	}
}
