package console

import (
	"github.com/luckyAkbar/central-worker-service/internal/client"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
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

	sibClient := client.NewSendInBlueClient()
	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	mailRepo := repository.NewMailRepository(db.PostgresDB)

	taskHandler := worker.NewTaskHandler(sibClient, mailRepo, workerClient)

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
