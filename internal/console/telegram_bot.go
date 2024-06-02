package console

import (
	"log"
	"net/http"
	"os"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/go-redis/redis/v9"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/telebot"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var bot *gotgbot.Bot

var telegramBotCmd = &cobra.Command{
	Use:   "telegram-bot",
	Short: "start telegram bot handler",
	Run:   telegramBot,
}

func init() {
	RootCmd.AddCommand(telegramBotCmd)
}

func init() {
	b, err := gotgbot.NewBot(config.TelegramBotToken(), &gotgbot.BotOpts{
		Client:             http.Client{},
		DisableTokenCheck:  false,
		UseTestEnvironment: config.TelegramBotUseTestEnv(),
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: config.TelegramBotTimeoutDuration(),
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})

	bot = b

	if err != nil {
		logrus.Panic("failed to run telegram bot: ", err)
	}
}

func telegramBot(cmd *cobra.Command, args []string) {
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

	db.InitializePostgresConn()
	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr(),
		Password:     config.RedisPassword(),
		DB:           config.RedisCacheDB(),
		MinIdleConns: config.RedisMinIdleConn(),
		MaxIdleConns: config.RedisMaxIdleConn(),
	})
	cacher := db.NewCacher(redisClient)

	teleRepo := repository.NewTelegramRepository(db.PostgresDB, cacher)
	diaryRepo := repository.NewDiaryRepo(db.PostgresDB, cacher)
	mailRepo := repository.NewMailRepository(db.PostgresDB)

	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	yourlsUtil := helper.NewYourlsUtil(config.YourlsBaseUrl(), config.YourlsSignature(), &http.Client{})

	mailUsecase := usecase.NewMailUsecase(mailRepo, workerClient)
	teleUsecase := usecase.NewTelegramUsecase(teleRepo, bot, workerClient, mailUsecase)
	diaryUsecase := usecase.NewDiaryUsecase(diaryRepo, yourlsUtil)

	telebotHandler := telebot.NewTelegramHandler(updater.Dispatcher, teleUsecase, teleRepo, workerClient, diaryUsecase, yourlsUtil)
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
