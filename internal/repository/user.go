package repository

import (
	"context"

	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository returns a new user repository
func NewUserRepository(db *gorm.DB) model.UserRepository {
	return &userRepo{
		db,
	}
}

// Create create new user to database
func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	log := logrus.WithFields(logrus.Fields{
		"ctx":  helper.DumpContext(ctx),
		"user": utils.Dump(user),
	})

	log.Info("start saving user to database")

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// FindByEmail find user by email address
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	log := logrus.WithFields(logrus.Fields{
		"ctx":   helper.DumpContext(ctx),
		"email": email,
	})

	log.Info("start to find user by email")

	user := &model.User{}
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Take(user).Error
	switch err {
	default:
		log.Error(err)
		return nil, err
	case gorm.ErrRecordNotFound:
		log.Info("user not found by email: ", email)
		return nil, ErrNotFound
	case nil:
		return user, nil
	}
}

// FindByID find user by id
func (r *userRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	log := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"id":  id,
	})

	log.Info("start to find user by id")

	user := &model.User{}
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Take(user).Error
	switch err {
	default:
		log.Error(err)
		return nil, err
	case gorm.ErrRecordNotFound:
		log.Info("user not found by id: ", id)
		return nil, ErrNotFound
	case nil:
		return user, nil
	}
}

// ActivateByUserID set is_active field with true in matching user id
func (r *userRepo) ActivateByUserID(ctx context.Context, id string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": helper.DumpContext(ctx),
		"id":  id,
	})

	logger.Info("activating user by id in database")

	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
		logger.Error("failed to update user activation db: ", err)
		return err
	}

	return nil
}
