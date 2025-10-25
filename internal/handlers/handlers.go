package handlers

import (
	"net/http"
	"time"

	"otp-basic/internal/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	auth *auth.AuthManager
}

func NewHandler(auth *auth.AuthManager) *Handler {
	return &Handler{
		auth: auth,
	}
}

type RegisterRequest struct {
	Issuer      string `json:"issuer" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
}

type RegisterResponse struct {
	MasterToken *auth.MasterToken `json:"master_token"`
	QRCodeURL   string            `json:"qr_code_url"`
	Secret      string            `json:"secret"`
}

type ValidateOTPRequest struct {
	UserID string `json:"user_id" binding:"required"`
	OTP    string `json:"otp" binding:"required"`
}

type ValidateOTPResponse struct {
	Valid bool `json:"valid"`
}

// RegisterMasterToken registers a new master token and returns OTP setup info
func (h *Handler) RegisterMasterToken(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Register new master token
	token, err := h.auth.RegisterMasterToken(req.Issuer, req.AccountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register master token",
		})
		return
	}

	// Generate QR code URL
	qrURL, err := h.auth.GetQRCodeURL(token.ID, req.Issuer, req.AccountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate QR code",
		})
		return
	}

	response := RegisterResponse{
		MasterToken: token,
		QRCodeURL:   qrURL,
		Secret:      token.Secret,
	}

	c.JSON(http.StatusCreated, response)
}

// ValidateOTP validates an OTP code for a user
func (h *Handler) ValidateOTP(c *gin.Context) {
	var req ValidateOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	valid := h.auth.ValidateOTP(req.UserID, req.OTP)
	response := ValidateOTPResponse{
		Valid: valid,
	}

	status := http.StatusOK
	if !valid {
		status = http.StatusUnauthorized
	}

	c.JSON(status, response)
}

// GetStatus returns the current server status (protected endpoint)
func (h *Handler) GetStatus(c *gin.Context) {
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	token, exists := h.auth.GetMasterToken(userID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Master token not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "authenticated",
		"user_id":    userID,
		"created_at": token.CreatedAt,
		"is_active":  token.IsActive,
		"timestamp":  time.Now(),
	})
}

// GetProtectedData returns some protected data (example endpoint)
func (h *Handler) GetProtectedData(c *gin.Context) {
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "This is protected data",
		"user_id": userID,
		"data": gin.H{
			"secret_info": "This information is only accessible with valid OTP",
			"timestamp":   time.Now(),
		},
	})
}
