package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type MasterToken struct {
	ID        string    `json:"id"`
	Secret    string    `json:"secret"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

type AuthManager struct {
	masterTokens map[string]*MasterToken
	otpSecrets   map[string]string // user_id -> secret
	mutex        sync.RWMutex
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		masterTokens: make(map[string]*MasterToken),
		otpSecrets:   make(map[string]string),
	}
}

func (am *AuthManager) RegisterMasterToken() (*MasterToken, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Generate a random secret for TOTP
	secret, err := am.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	token := &MasterToken{
		ID:        uuid.New().String(),
		Secret:    secret,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	am.masterTokens[token.ID] = token
	am.otpSecrets[token.ID] = secret

	return token, nil
}

func (am *AuthManager) ValidateOTP(userID, otpCode string) bool {
	am.mutex.RLock()
	secret, exists := am.otpSecrets[userID]
	am.mutex.RUnlock()

	if !exists {
		return false
	}

	return totp.Validate(otpCode, secret)
}

func (am *AuthManager) GetMasterToken(userID string) (*MasterToken, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	token, exists := am.masterTokens[userID]
	return token, exists
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
	am.mutex.RLock()
	secret, exists := am.otpSecrets[userID]
	am.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("user not found")
	}

	return totp.GenerateCode(secret, time.Now())
}

func (am *AuthManager) GetQRCodeURL(userID, issuer, accountName string) (string, error) {
	am.mutex.RLock()
	secret, exists := am.otpSecrets[userID]
	am.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("user not found")
	}

	// Generate QR code URL manually
	url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, accountName, secret, issuer)

	return url, nil
}
