package server

import (
	"otp-basic/internal/auth"
	"otp-basic/internal/database"
	"otp-basic/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	auth   *auth.AuthManager
	db     *database.DB
}

func NewServer() (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}

	authManager := auth.NewAuthManager(db)
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
		db:     db,
	}, nil
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
