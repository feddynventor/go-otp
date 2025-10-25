package server

import (
	"otp-basic/internal/auth"
	"otp-basic/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	auth   *auth.AuthManager
}

func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	authManager := auth.NewAuthManager()
	handler := handlers.NewHandler(authManager)

	// Public routes
	router.POST("/register", handler.RegisterMasterToken)
	router.POST("/validate-otp", handler.ValidateOTP)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(authManager.OTPMiddleware())
	{
		protected.GET("/status", handler.GetStatus)
		protected.GET("/protected-data", handler.GetProtectedData)
	}

	return &Server{
		router: router,
		auth:   authManager,
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
