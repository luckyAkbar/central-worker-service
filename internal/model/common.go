package model

import (
	validate "github.com/go-playground/validator/v10"
)

type contextKey string

// context key for request id
var ReqIDCtxKey contextKey = "request_id_context_key"

var validator = validate.New()

// UsecaseError error returned from usecase
type UsecaseError struct {
	UnderlyingError error  `json:"error"`
	Message         string `json:"message"`
}

var (
	// NilUsecaseError returned when no error is returned
	NilUsecaseError UsecaseError = UsecaseError{}
)
