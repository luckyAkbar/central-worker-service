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

func (s *Service) handleRegisterUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.RegisterUserInput{}
		if err := c.Bind(input); err != nil {
			logrus.Info("failed to bind: ", err)
			return ErrBadRequest
		}

		_, e := s.userUsecase.Register(c.Request().Context(), input)
		switch e.UnderlyingError {
		default:
			logrus.WithFields(logrus.Fields{
				"ctx":   helper.DumpContext(c.Request().Context()),
				"input": utils.Dump(input),
			}).Error(e.UnderlyingError)
			return sendError(http.StatusInternalServerError, e.Message)

		case usecase.ErrValidations:
			return sendError(http.StatusBadRequest, e.Message)

		case usecase.ErrAlreadyExists:
			return sendError(http.StatusBadRequest, e.Message)

		case nil:
			return c.NoContent(http.StatusCreated)
		}
	}
}

func (s *Service) handleUserActivation() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("userID")
		signature := c.QueryParam("signature")

		err := s.userUsecase.Activate(c.Request().Context(), userID, signature)
		switch err.UnderlyingError {
		default:
			logrus.WithFields(logrus.Fields{
				"ctx":       helper.DumpContext(c.Request().Context()),
				"userID":    userID,
				"signature": signature,
			}).Error(err)
			return sendError(http.StatusInternalServerError, err.Message)

		case usecase.ErrValidations:
			return sendError(http.StatusBadRequest, err.Message)

		case usecase.ErrNotFound:
			return sendError(http.StatusNotFound, err.Message)

		case usecase.ErrForbidden:
			return sendError(http.StatusForbidden, err.Message)

		case nil:
			return c.NoContent(http.StatusOK)
		}
	}
}
