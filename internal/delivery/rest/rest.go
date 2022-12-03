package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/model"
)

type Service struct {
	apiGroup    *echo.Group
	mailUsecase model.MailUsecase
}

func Init(apiGroup *echo.Group, mailUsecase model.MailUsecase) {
	s := &Service{
		apiGroup:    apiGroup,
		mailUsecase: mailUsecase,
	}

	s.InitAPIRoutes()
}

func (s *Service) InitAPIRoutes() {
	s.apiGroup.POST("/email/enqueue/", s.handleEnqueueEmail())
}
