package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	ErrBadRequest = echo.NewHTTPError(http.StatusBadRequest, "bad request")
	ErrInternal   = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
)
