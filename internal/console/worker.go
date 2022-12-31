package console

import (
	"github.com/luckyAkbar/central-worker-service/internal/client"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
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

	mailUtility := util.NewMailUtility(sibClient, mailgunClient)

	mailRepo := repository.NewMailRepository(db.PostgresDB)
	userRepo := repository.NewUserRepository(db.PostgresDB)
	siakadRepo := repository.NewSiakadRepository(db.PostgresDB)

	taskHandler := worker.NewTaskHandler(mailUtility, mailRepo, workerClient, userRepo, siakadRepo)

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
