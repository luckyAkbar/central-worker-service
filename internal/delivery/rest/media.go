package rest

import (
	"net/http"
	"strconv"

	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (s *Service) handleUploadImage() echo.HandlerFunc {
	return func(c echo.Context) error {
		imageName := c.FormValue("image_name")
		isPrivate, err := strconv.ParseBool(c.FormValue("is_private"))
		if err != nil {
			logrus.Info("failed to parse form value is_private: ", err)
			return ErrBadRequest
		}

		file, err := c.FormFile("file")
		if err != nil {
			logrus.Info("failed to read file from form value: ", err)
			return ErrBadRequest
		}

		input := &model.UploadImageInput{
			ImageName: imageName,
			IsPrivate: isPrivate,
		}

		image, ucErr := s.imageUsecase.Upload(c.Request().Context(), input, file)
		switch ucErr.UnderlyingError {
		default:
			logrus.WithFields(logrus.Fields{
				"ctx":   helper.DumpContext(c.Request().Context()),
				"input": utils.Dump(input),
			}).Error(err)

			return ErrInternal

		case usecase.ErrValidations:
			return sendError(http.StatusBadRequest, ucErr.Message)

		case nil:
			return c.JSON(http.StatusOK, image)
		}

	}
}
