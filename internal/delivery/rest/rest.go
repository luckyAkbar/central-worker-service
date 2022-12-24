package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/model"
)

// Service rest service
type Service struct {
	apiGroup    *echo.Group
	authGroup   *echo.Group
	mailUsecase model.MailUsecase
	userUsecase model.UserUsecase
}

// Init init rest service
func Init(apiGroup *echo.Group, authGroup *echo.Group, mailUsecase model.MailUsecase, userUsecase model.UserUsecase) {
	s := &Service{
		apiGroup:    apiGroup,
		authGroup:   authGroup,
		mailUsecase: mailUsecase,
		userUsecase: userUsecase,
	}

	s.initAPIRoutes()
	s.initAuthRoutes()
}

func (s *Service) initAuthRoutes() {
	s.authGroup.POST("/user/", s.handleRegisterUser())
	s.authGroup.GET("/user/activation/:userID/", s.handleUserActivation())
}

// InitAPIRoutes initialize api routes (prefixed by 'api')
func (s *Service) initAPIRoutes() {
	s.apiGroup.POST("/email/enqueue/", s.handleEnqueueEmail())
}
