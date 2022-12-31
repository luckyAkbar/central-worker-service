package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/model"
)

// Service rest service
type Service struct {
	apiGroup     *echo.Group
	authGroup    *echo.Group
	mailUsecase  model.MailUsecase
	userUsecase  model.UserUsecase
	authUsecase  model.AuthUsecase
	imageUsecase model.ImageUsecase
}

// Init init rest service
func Init(apiGroup *echo.Group, authGroup *echo.Group, mailUsecase model.MailUsecase, userUsecase model.UserUsecase, authUsecase model.AuthUsecase, imageUsecase model.ImageUsecase) {
	s := &Service{
		apiGroup:     apiGroup,
		authGroup:    authGroup,
		mailUsecase:  mailUsecase,
		userUsecase:  userUsecase,
		authUsecase:  authUsecase,
		imageUsecase: imageUsecase,
	}

	s.initAPIRoutes()
	s.initAuthRoutes()
}

func (s *Service) initAuthRoutes() {
	s.authGroup.POST("/login/", s.handleLogin())
	s.authGroup.POST("/user/", s.handleRegisterUser())
	s.authGroup.GET("/user/activation/:userID/", s.handleUserActivation())
}

// InitAPIRoutes initialize api routes (prefixed by 'api')
func (s *Service) initAPIRoutes() {
	s.apiGroup.Use(s.authUsecase.AuthMiddleware(true))
	s.apiGroup.POST("/email/enqueue/", s.handleEnqueueEmail())
	s.apiGroup.POST("/media/image/", s.handleUploadImage())
}
