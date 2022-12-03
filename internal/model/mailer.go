package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
)

// MailStatus is the status of a mail
type MailStatus string

var (
	MailStatusOnProgress MailStatus = "ON_PROGRESS"
	MailStatusSuccess    MailStatus = "SUCCESS"
	MailStatusFailed     MailStatus = "FAILED"
)

type MailingInput struct {
	To          []lib.SendSmtpEmailTo  `json:"to" validate:"required"`
	Cc          []lib.SendSmtpEmailCc  `json:"cc,omitempty"`
	Bcc         []lib.SendSmtpEmailBcc `json:"bcc,omitempty"`
	HTMLContent string                 `json:"html_content" validate:"required"`
	Subject     string                 `json:"subject" validate:"required"`
}

func (mi *MailingInput) Validate() error {
	return validator.Struct(mi)
}

type Mail struct {
	ID string `json:"id"`

	// To is the marshalled verion of []lib.SendSmtpEmailTo
	To string `json:"to"`
	// Cc is the marshalled verion of []lib.SendSmtpEmailCc
	Cc *sql.NullString `json:"cc,omitempty"`
	// Bcc is the marshalled verion of []lib.SendSmtpEmailBcc
	Bcc         *sql.NullString `json:"bcc,omitempty"`
	HTMLContent string          `json:"html_content"`
	Subject     string          `json:"subject"`
	CreatedAt   time.Time       `json:"created_at"`
	DeliveredAt *sql.NullTime   `json:"delivered_at,omitempty"`
	Status      MailStatus      `json:"status"`
	Metadata    *sql.NullString `json:"metadata,omitempty"`
}

func (m *Mail) SendInBlueTo() ([]lib.SendSmtpEmailTo, error) {
	logrus.Info("converting string to []lib.SendSmtpEmailTo")

	var to []lib.SendSmtpEmailTo
	if err := json.Unmarshal([]byte(m.To), &to); err != nil {
		logrus.Error(err)
		return to, err
	}

	logrus.Info("result to: ", utils.Dump(to))

	return to, nil
}

func (m *Mail) SendInBlueCc() (cc []lib.SendSmtpEmailCc, err error) {
	logrus.Info("converting string to []lib.SendSmtpEmailCc")

	if !m.Cc.Valid {
		return cc, nil
	}

	if err := json.Unmarshal([]byte(m.Cc.String), &cc); err != nil {
		logrus.Error(err)
		return cc, err
	}

	logrus.Info("result: ", utils.Dump(cc))

	return cc, nil
}

func (m *Mail) SendInBlueBcc() (bcc []lib.SendSmtpEmailBcc, err error) {
	logrus.Info("converting string to []lib.SendSmtpEmailBcc")

	if !m.Bcc.Valid {
		return bcc, nil
	}

	if err := json.Unmarshal([]byte(m.Bcc.String), &bcc); err != nil {
		logrus.Error(err)
		return bcc, err
	}

	logrus.Info("result bcc: ", utils.Dump(bcc))

	return bcc, nil
}

type MailUsecase interface {
	Enqueue(ctx context.Context, input *MailingInput) (*Mail, UsecaseError)
}

type MailRepository interface {
	Create(ctx context.Context, mail *Mail) error
	Update(ctx context.Context, mail *Mail) error
}
