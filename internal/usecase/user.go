package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/sendinblue/APIv3-go-library/lib"
	"github.com/sirupsen/logrus"
)

type userUsecase struct {
	userRepo    model.UserRepository
	mailUsecase model.MailUsecase
}

// NewUserUsecase return a new user usecase
func NewUserUsecase(userRepo model.UserRepository, mailUsecase model.MailUsecase) model.UserUsecase {
	return &userUsecase{
		userRepo,
		mailUsecase,
	}
}

// Register register user to and send email to activate the user
func (u *userUsecase) Register(ctx context.Context, input *model.RegisterUserInput) (*model.User, model.UsecaseError) {
	log := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"input": utils.Dump(input),
	})

	log.Info("start to register user")

	if err := input.Validate(); err != nil {
		log.Info("invalid input on register user: ", err.Error())
		return nil, model.UsecaseError{
			UnderlyingError: ErrValidations,
			Message:         err.Error(),
		}
	}

	log.Info("check user by email is already registered")

	_, err := u.userRepo.FindByEmail(ctx, input.Email)
	switch err {
	default:
		log.Error(err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}

	case nil:
		log.Info("mail already used: ", input.Email)
		return nil, model.UsecaseError{
			UnderlyingError: ErrAlreadyExists,
			Message:         "email already taken",
		}

	case repository.ErrNotFound:
	}

	user := &model.User{
		ID:        helper.GenerateID(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  helper.CreateHashSHA512([]byte(input.Password)),
		CreatedAt: time.Now().UTC(),
		IsActive:  false,
	}

	log.Info("creating user to db...")

	// FIXME: use transaction rollback and commit to ensure if mail successfully registered by usecase
	// if fails, rollback don't create user to db
	if err := u.userRepo.Create(ctx, user); err != nil {
		log.Error(err)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgDatabaseError,
		}
	}

	log.Info("creating activation url")

	sig := user.GenerateActivationSignatureInput()
	signature := helper.CreateHashSHA512([]byte(sig))
	activationURL := fmt.Sprintf("%s/%s/?signature=%s", config.UserActivationBaseURL(), user.ID, signature)

	log.Info("enqueueing the email activation")

	mail, ucErr := u.mailUsecase.Enqueue(ctx, &model.MailingInput{
		To: []lib.SendSmtpEmailTo{
			{
				Email: user.Email,
			},
		},
		Subject:     "User Activation",
		HTMLContent: helper.HTMLContentForUserRegistrationEmail(user.Username, activationURL),
	})

	log.Info("received user activation mail: ", utils.Dump(mail))

	switch ucErr.UnderlyingError {
	default:
		log.Error("error from mail usecase enqueue: ", ucErr.UnderlyingError)
		return nil, model.UsecaseError{
			UnderlyingError: ErrInternal,
			Message:         MsgInternalError,
		}

	case nil:
		break
	}

	log.Info("finish registering user")

	return user, model.NilUsecaseError
}
