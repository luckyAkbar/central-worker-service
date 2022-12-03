package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (s *Service) handleEnqueueEmail() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := &model.MailingInput{}
		if err := c.Bind(input); err != nil {
			logrus.Info(err)
			return ErrBadRequest
		}

		mail, err := s.mailUsecase.Enqueue(c.Request().Context(), input)
		switch err.UnderlyingError {
		default:
			logrus.WithFields(logrus.Fields{
				"ctx": helper.DumpContext(c.Request().Context()),
			}).Error(err)
			return ErrInternal

		case usecase.ErrValidations:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"err":     err.UnderlyingError.Error(),
				"message": err.Message,
			})
		case nil:
			return c.JSON(http.StatusCreated, mail)
		}
	}
}
