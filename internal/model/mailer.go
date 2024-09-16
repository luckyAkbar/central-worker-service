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

// list mail status
var (
	MailStatusOnProgress MailStatus = "ON_PROGRESS"
	MailStatusSuccess    MailStatus = "SUCCESS"
	MailStatusFailed     MailStatus = "FAILED"
)

// MailClientSignature signature for every registered mailing client
type MailClientSignature string

// list of registered mailing clients
const (
	SendInBlueClientSignature MailClientSignature = "sendinblue"
	MailgunClientSignature    MailClientSignature = "mailgun"
)

// MailingInput input to create mailing list
type MailingInput struct {
	To          []lib.SendSmtpEmailTo  `json:"to" validate:"required"`
	Cc          []lib.SendSmtpEmailCc  `json:"cc,omitempty"`
	Bcc         []lib.SendSmtpEmailBcc `json:"bcc,omitempty"`
	HTMLContent string                 `json:"html_content" validate:"required"`
	Subject     string                 `json:"subject" validate:"required"`
}

// Validate validate struct
func (mi *MailingInput) Validate() error {
	return validator.Struct(mi)
}

// Mail represents a database table mails
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

// SendInBlueTo get send in blue SendSmtpEmailTo
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

// SendInBlueCc get send in blue SendSmtpEmailCc
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

// SendInBlueBcc get send in blue SendSmtpEmailBcc
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

// MailgunTo convert to to mailgun compatible to
func (m *Mail) MailgunTo() ([]string, error) {
	to := []lib.SendSmtpEmailTo{}

	if err := json.Unmarshal([]byte(m.To), &to); err != nil {
		logrus.Error("failed to unmarshal to mailgun to: ", err)
		return []string{""}, err
	}

	var res []string
	for _, t := range to {
		res = append(res, t.Email)
	}

	logrus.Info("res to mailgun: ", res)

	return res, nil
}

// MailgunCC convert cc to mailgun compatible cc
func (m *Mail) MailgunCC() ([]string, error) {
	logrus.Info("start to convert mailgun cc")

	if m.Cc.String == "" {
		return []string{}, nil
	}

	sibCC := []lib.SendSmtpEmailCc{}
	if err := json.Unmarshal([]byte(m.Cc.String), &sibCC); err != nil {
		logrus.Error("failed to unmarhsall mailgun to lib smtp cc: ", err)
		return []string{}, err
	}

	var cc []string
	for _, c := range sibCC {
		cc = append(cc, c.Email)
	}

	logrus.Info("finish converting to mailgun cc: ", cc)

	return cc, nil
}

// MailgunBCC convert bcc to mailgun compatible bcc
func (m *Mail) MailgunBCC() ([]string, error) {
	logrus.Info("start converting mailgun bcc")

	if m.Bcc.String == "" {
		return []string{}, nil
	}

	sibBCC := []lib.SendSmtpEmailBcc{}
	if err := json.Unmarshal([]byte(m.Cc.String), &sibBCC); err != nil {
		logrus.Error("failed to unmarhsall mailgun to lib smtp cc: ", err)
		return []string{}, err
	}

	var bcc []string
	for _, c := range sibBCC {
		bcc = append(bcc, c.Email)
	}

	logrus.Info("finish converting to mailgun bcc: ", bcc)

	return bcc, nil
}

// MailResultMetadata intended to generally represent metadata from mail. Buil to easily know which mailing client used
// when sending email, and can easily know which struct to be unmarshall the detail field
type MailResultMetadata struct {
	Detail    string              `json:"detail"`
	Signature MailClientSignature `json:"signature"`
}

// MailClient must be implemented by any client for mailing purposes
type MailClient interface {
	// SendEmail send email, returning the metadata, and error if any
	SendEmail(ctx context.Context, mail *Mail) (string, error)

	// GetClientName client name for mail client
	GetClientName() MailClientSignature
}

// MailUtility mail utility interface
type MailUtility interface {
	SendEmail(ctx context.Context, mail *Mail) (string, MailClientSignature, error)
}

// MailUsecase usecase for mail
type MailUsecase interface {
	Enqueue(ctx context.Context, input *MailingInput) (*Mail, UsecaseError)
}

// MailRepository repository for mail
type MailRepository interface {
	Create(ctx context.Context, mail *Mail) error
	Update(ctx context.Context, mail *Mail) error
}
