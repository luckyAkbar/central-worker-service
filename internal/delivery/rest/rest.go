package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/model"
)

// Service rest service
type Service struct {
	apiGroup    *echo.Group
	mailUsecase model.MailUsecase
}

// Init init rest service
func Init(apiGroup *echo.Group, mailUsecase model.MailUsecase) {
	s := &Service{
		apiGroup:    apiGroup,
		mailUsecase: mailUsecase,
	}

	s.InitAPIRoutes()
}

// InitAPIRoutes initialize api routes (prefixed by 'api')
func (s *Service) InitAPIRoutes() {
	s.apiGroup.POST("/email/enqueue/", s.handleEnqueueEmail())
}
