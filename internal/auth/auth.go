package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"otp-basic/internal/database"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

// MasterToken is an alias for database.MasterToken for backward compatibility
type MasterToken = database.MasterToken

type AuthManager struct {
	db *database.DB
}

func NewAuthManager(db *database.DB) *AuthManager {
	return &AuthManager{
		db: db,
	}
}

func (am *AuthManager) RegisterMasterToken(issuer, accountName string) (*MasterToken, error) {
	// Generate a random secret for TOTP
	secret, err := am.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	token := &MasterToken{
		ID:          uuid.New().String(),
		Secret:      secret,
		CreatedAt:   time.Now(),
		IsActive:    true,
		Issuer:      &issuer,
		AccountName: &accountName,
	}

	// Save to database
	if err := am.db.CreateMasterToken(token); err != nil {
		return nil, fmt.Errorf("failed to save master token to database: %w", err)
	}

	return token, nil
}

func (am *AuthManager) ValidateOTP(userID, otpCode string) bool {
	// Get master token from database
	token, err := am.db.GetMasterToken(userID)
	if err != nil || token == nil || !token.IsActive {
		return false
	}

	return totp.Validate(otpCode, token.Secret)
}

func (am *AuthManager) GetMasterToken(userID string) (*MasterToken, bool) {
	token, err := am.db.GetMasterToken(userID)
	if err != nil || token == nil {
		return nil, false
	}
	return token, true
}

func (am *AuthManager) generateSecret() (string, error) {
	// Generate 20 random bytes
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

func (am *AuthManager) GenerateOTPCode(userID string) (string, error) {
	// Get master token from database
	token, err := am.db.GetMasterToken(userID)
	if err != nil || token == nil || !token.IsActive {
		return "", fmt.Errorf("user not found or inactive")
	}

	return totp.GenerateCode(token.Secret, time.Now())
}

func (am *AuthManager) GetQRCodeURL(userID, issuer, accountName string) (string, error) {
	// Get master token from database
	token, err := am.db.GetMasterToken(userID)
	if err != nil || token == nil || !token.IsActive {
		return "", fmt.Errorf("user not found or inactive")
	}

	// Generate QR code URL manually
	url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, accountName, token.Secret, issuer)

	return url, nil
}
