package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// list of rest service error
var (
	ErrBadRequest = echo.NewHTTPError(http.StatusBadRequest, "bad request")
	ErrInternal   = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	ErrNotFound   = echo.NewHTTPError(http.StatusNotFound, "not found")
)

func sendError(status int, msg string) *echo.HTTPError {
	return echo.NewHTTPError(status, msg)
}
