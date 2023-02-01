// Package rest is HTTP rest presenter layer for this project
package rest

import (
	"net/http"

	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (s *Service) handleLogin() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.LoginInput{}
		if err := c.Bind(input); err != nil {
			logrus.Info("failed binding input")
			return ErrBadRequest
		}

		session, err := s.authUsecase.Login(c.Request().Context(), input)
		switch err.UnderlyingError {
		default:
			logrus.WithFields(logrus.Fields{
				"ctx":   helper.DumpContext(c.Request().Context()),
				"input": utils.Dump(input),
			}).Error(err.UnderlyingError)
			return sendError(http.StatusInternalServerError, err.Message)

		case usecase.ErrForbidden:
			return sendError(http.StatusForbidden, err.Message)

		case usecase.ErrNotFound:
			return sendError(http.StatusNotFound, err.Message)

		case usecase.ErrValidations:
			return sendError(http.StatusBadRequest, err.Message)

		case nil:
			return c.JSON(http.StatusOK, session)
		}
	}
}
