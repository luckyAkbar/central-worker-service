package usecase

import "errors"

var (
	ErrValidations = errors.New("validation error")
	ErrInternal    = errors.New("internal error")
)

var (
	MsgDatabaseError      = "operation failed, database error"
	MsgFailedRegisterTask = "failed to register task"
)
