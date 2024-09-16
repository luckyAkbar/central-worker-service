package rest

import (
	"html/template"
	"net/url"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/sirupsen/logrus"
)

func (s *Service) handleRenderDiary() echo.HandlerFunc {
	return func(c echo.Context) error {
		input := c.QueryParam("key")
		if input == "" {
			return ErrNotFound
		}

		key, err := url.QueryUnescape(input)
		if err != nil {
			return ErrBadRequest
		}

		data, err := s.diaryUsecase.GetDiariesOnFrontendRenderData(c.Request().Context(), key)
		switch err {
		default:
			return ErrInternal
		case usecase.ErrNotFound:
			return ErrNotFound
		case nil:
			break
		}

		htmlFile := path.Join("views", "diary.html")
		tmpl, err := template.ParseFiles(htmlFile)
		if err != nil {
			logrus.WithError(err).Error("failed to parse diary html file")
			return ErrInternal
		}

		if err := tmpl.Execute(c.Response().Writer, data); err != nil {
			logrus.WithError(err).Error("failed to execute diary html template file")
			return ErrInternal
		}

		return nil
	}
}
