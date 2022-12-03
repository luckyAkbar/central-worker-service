package client

import (
	"context"
	"fmt"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	sendinblue "github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
)

type SIB struct {
	client *sendinblue.APIClient
}

func NewSendInBlueClient() *SIB {
	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", config.SendinblueAPIKey())

	return &SIB{
		client: sendinblue.NewAPIClient(cfg),
	}
}

func (s *SIB) SendEmail(ctx context.Context, body sendinblue.SendSmtpEmail) (sendinblue.CreateSmtpEmail, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"body": utils.Dump(body),
	})

	logger.Info("sending email using send in blue")

	email, res, err := s.client.TransactionalEmailsApi.SendTransacEmail(ctx, body)
	if err != nil {
		logger.Error(err)
		return email, err
	}

	defer helper.WrapCloser(res.Body.Close)

	logrus.Info("receiving response from send in blue: ", utils.Dump(res.StatusCode), utils.Dump(res.Body))

	if res.StatusCode != 201 {
		e := fmt.Errorf("failed to send email send in blue: %v", res)
		logger.Error(e)
		return email, e
	}

	logrus.Info("received create smtp email send in blue: ", utils.Dump(email))

	return email, nil
}
