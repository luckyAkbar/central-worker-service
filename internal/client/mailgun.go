package client

import (
	"context"
	"errors"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"
)

// Mailgun :nodoc:
type Mailgun struct {
	client      *mailgun.MailgunImpl
	isActivated bool
}

// NewMailgunClient create new mailgun client
func NewMailgunClient(client *mailgun.MailgunImpl, isActivated bool) *Mailgun {
	return &Mailgun{
		client,
		isActivated,
	}
}

// SendEmail send email using sendinblue
func (mg *Mailgun) SendEmail(ctx context.Context, mail *model.Mail) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"mail": utils.Dump(mail),
	})

	if !mg.isActivated {
		log.Info("mailgun is not activated by configuration")
		return "", errors.New("mailgun is not activated by configuration")
	}

	log.Info("sending mail using mailgun")

	to, err := mail.MailgunTo()
	if err != nil {
		log.Error("failed to get mailgun to: ", err)
		return "", err
	}

	cc, err := mail.MailgunCC()
	if err != nil {
		log.Error("failed to get mailgun cc: ", err)
		return "", err
	}

	bcc, err := mail.MailgunBCC()
	if err != nil {
		log.Error("failed to get mailgun to: ", err)
		return "", err
	}

	message := mg.client.NewMessage(config.ServerSenderEmail(), mail.Subject, "", to...)
	message.SetHtml(mail.HTMLContent)

	for _, email := range cc {
		message.AddCC(email)
	}

	for _, email := range bcc {
		message.AddBCC(email)
	}

	resp, id, err := mg.client.Send(ctx, message)
	if err != nil {
		log.Error("failed send email using mailgun: ", err)
		return "", err
	}

	log.Info("received response from mailgun: ", resp)
	log.Info("received id from mailgun: ", id)

	return id, nil
}

// GetClientName returning client name signature
func (mg *Mailgun) GetClientName() model.MailClientSignature {
	return model.MailgunClientSignature
}
