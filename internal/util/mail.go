package util

import (
	"context"
	"errors"
	"fmt"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/client"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

type mail struct {
	clients []model.MailClient
}

// NewMailUtility return new mail utility
func NewMailUtility(sendinblue *client.SIB, mailgun *client.Mailgun) model.MailUtility {
	return &mail{
		clients: []model.MailClient{
			sendinblue,
			mailgun,
		},
	}
}

// SendEmail send email using any available mailinng client. Returning metadata, client signature and error
// will retry using the next available client if the previous returning error
func (m *mail) SendEmail(ctx context.Context, mail *model.Mail) (string, model.MailClientSignature, error) {
	log := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"mail": utils.Dump(mail),
	})

	log.Info("starting send email from mail utility")

	for _, client := range m.clients {
		metadata, err := client.SendEmail(ctx, mail)

		if err == nil {
			return metadata, client.GetClientName(), nil
		}

		log.Error(fmt.Sprintf("failed to send email using client: %s: %s", client.GetClientName(), err.Error()))
	}

	return "", "", errors.New("failed sending email")
}
