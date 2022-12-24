package usecase

import "errors"

// list of usecase underlying error
var (
	ErrValidations   = errors.New("validation error")
	ErrInternal      = errors.New("internal error")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
)

// list of standard error message
var (
	MsgDatabaseError      = "operation failed, database error"
	MsgFailedRegisterTask = "failed to register task"
	MsgInternalError      = "internal error"
	MsgNotFound           = "not found"
	MsgForbidden          = "forbidden"
)
