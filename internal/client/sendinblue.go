package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	sendinblue "github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
)

// SIB send in blue client
type SIB struct {
	client      *sendinblue.APIClient
	isActivated bool
}

// NewSendInBlueClient creates a new SIB client
func NewSendInBlueClient(cfg *sendinblue.Configuration, isActivated bool) *SIB {
	return &SIB{
		client:      sendinblue.NewAPIClient(cfg),
		isActivated: isActivated,
	}
}

// SendEmail sends an email. error if status code from sendinblue server is not 201
func (s *SIB) SendEmail(ctx context.Context, input *model.Mail) (string, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	if !s.isActivated {
		logger.Info("sendinblue is not activated by configuration")
		return "", errors.New("sendinblue is not activated by configuration")
	}

	logger.Info("sending email using send in blue")

	to, err := input.SendInBlueTo()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	cc, err := input.SendInBlueCc()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	bcc, err := input.SendInBlueBcc()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	body := sendinblue.SendSmtpEmail{
		Sender:      config.SendInBlueSender(),
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		HtmlContent: input.HTMLContent,
		Subject:     input.Subject,
	}

	email, res, err := s.client.TransactionalEmailsApi.SendTransacEmail(ctx, body)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	defer helper.WrapCloser(res.Body.Close)

	logrus.Info("receiving response from send in blue: ", utils.Dump(res.StatusCode), utils.Dump(res.Body))

	if res.StatusCode != 201 {
		e := fmt.Errorf("failed to send email send in blue: %v", res)
		logger.Error(e)
		return "", e
	}

	logrus.Info("received create smtp email send in blue: ", utils.Dump(email))

	metadata := utils.Dump(email)

	return metadata, nil
}

// GetClientName return client name signature sendinblue
func (s *SIB) GetClientName() model.MailClientSignature {
	return model.SendInBlueClientSignature
}
