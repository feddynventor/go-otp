package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OTPRequest struct {
	UserID string `json:"user_id" binding:"required"`
	OTP    string `json:"otp" binding:"required"`
}

func (am *AuthManager) OTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for OTP in header or body
		var otpReq OTPRequest

		// Try to get from header first
		userID := c.GetHeader("X-User-ID")
		otpCode := c.GetHeader("X-OTP")

		if userID == "" || otpCode == "" {
			// Try to get from body
			if err := c.ShouldBindJSON(&otpReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Missing OTP credentials. Provide X-User-ID and X-OTP headers or JSON body with user_id and otp",
				})
				c.Abort()
				return
			}
			userID = otpReq.UserID
			otpCode = otpReq.OTP
		}

		// Validate OTP
		if !am.ValidateOTP(userID, otpCode) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid OTP",
			})
			c.Abort()
			return
		}

		// Store user ID in context for use in handlers
		c.Set("user_id", userID)
		c.Next()
	}
}

// Helper function to extract user ID from context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}
