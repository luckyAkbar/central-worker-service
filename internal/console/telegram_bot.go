package console

import (
	"log"
	"net/http"
	"os"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/telebot"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var telegramBotCmd = &cobra.Command{
	Use:   "telegram-bot",
	Short: "start telegram bot handler",
	Run:   telegramBot,
}

func init() {
	RootCmd.AddCommand(telegramBotCmd)
}

func telegramBot(cmd *cobra.Command, args []string) {
	bot, err := gotgbot.NewBot(config.TelegramBotToken(), &gotgbot.BotOpts{
		Client:             http.Client{},
		DisableTokenCheck:  false,
		UseTestEnvironment: config.TelegramBotUseTestEnv(),
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: config.TelegramBotTimeoutDuration(),
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})

	if err != nil {
		logrus.Panic("failed to run telegram bot: ", err)
	}

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: log.New(os.Stdout, "telegram_bot: ", log.LUTC),
		DispatcherOpts: ext.DispatcherOpts{
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				logrus.Error("experiencing error from dispatcher error handler: err", err)
				_, err = ctx.EffectiveMessage.Reply(
					b,
					"bot experiencing unexpected error. Please try again later.",
					&gotgbot.SendMessageOpts{
						ReplyToMessageId: ctx.Message.MessageId,
					},
				)

				if err != nil {
					logrus.Error("failed to send error message from dispatcher error handler: ", err)
				}

				return ext.DispatcherActionNoop
			},
		},
	})

	telebotHandler := telebot.NewTelegramHandler(updater.Dispatcher)
	telebotHandler.RegisterHandlers()

	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: config.TelegramBotDropPendingUpdate(),
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: config.TelegramBotTimeout(),
		},
	})

	if err != nil {
		logrus.Panic("unable to start polling by updater: ", err)
	}

	logrus.Info("telegram bot has been running...")

	updater.Idle()
}
