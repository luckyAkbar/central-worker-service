package console

import (
	"github.com/go-redis/redis/v9"
	"github.com/luckyAkbar/central-worker-service/internal/client"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/luckyAkbar/central-worker-service/internal/util"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
	"github.com/mailgun/mailgun-go/v4"
	sendinblue "github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "worker",
	Long:  `This subcommand used to run worker`,
	Run:   runWorker,
}

func init() {
	RootCmd.AddCommand(workerCmd)
}

func runWorker(_ *cobra.Command, _ []string) {
	db.InitializePostgresConn()

	sibConfig := sendinblue.NewConfiguration()
	sibConfig.AddDefaultHeader("api-key", config.SendinblueAPIKey())
	sibClient := client.NewSendInBlueClient(sibConfig, config.SendInBlueIsActivated())

	mg := mailgun.NewMailgun(config.MailgunDomain(), config.MailgunPrivateAPIKey())
	mailgunClient := client.NewMailgunClient(mg, config.MailgunIsActivated())

	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr(),
		Password:     config.RedisPassword(),
		DB:           config.RedisCacheDB(),
		MinIdleConns: config.RedisMinIdleConn(),
		MaxIdleConns: config.RedisMaxIdleConn(),
	})
	cacher := db.NewCacher(redisClient)

	mailUtility := util.NewMailUtility(sibClient, mailgunClient)

	mailRepo := repository.NewMailRepository(db.PostgresDB)
	userRepo := repository.NewUserRepository(db.PostgresDB)
	siakadRepo := repository.NewSiakadRepository(db.PostgresDB, cacher)
	telegramRepo := repository.NewTelegramRepository(db.PostgresDB, cacher)

	mailUsecase := usecase.NewMailUsecase(mailRepo, workerClient)

	telegramUsecase := usecase.NewTelegramUsecase(telegramRepo, bot, workerClient, mailUsecase)

	taskHandler := worker.NewTaskHandler(mailUtility, mailRepo, workerClient, userRepo, siakadRepo, telegramUsecase)

	wrk, err := worker.NewServer(config.WorkerBrokerRedisHost(), taskHandler)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("starting worker...")

	if err := wrk.Start(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("worker started")
}
