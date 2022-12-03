package model

import (
	validate "github.com/go-playground/validator/v10"
)

type contextKey string

// ReqIDCtxKey context key for request id
var ReqIDCtxKey contextKey = "request_id_context_key"

var validator = validate.New()

// UsecaseError error returned from usecase
type UsecaseError struct {
	// UnderlyingError is the real error
	UnderlyingError error `json:"error"`

	// Message is the error message. Cound be filled with hint to improve request
	Message string `json:"message"`
}

var (
	// NilUsecaseError returned when no error is returned
	NilUsecaseError = UsecaseError{}
)
