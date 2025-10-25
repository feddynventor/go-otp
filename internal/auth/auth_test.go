package auth

import (
	"testing"
)

func TestAuthManager_RegisterMasterToken(t *testing.T) {
	am := NewAuthManager()

	token, err := am.RegisterMasterToken()
	if err != nil {
		t.Fatalf("Failed to register master token: %v", err)
	}

	if token.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if token.Secret == "" {
		t.Error("Expected non-empty secret")
	}

	if !token.IsActive {
		t.Error("Expected token to be active")
	}

	// Check if token is stored
	storedToken, exists := am.GetMasterToken(token.ID)
	if !exists {
		t.Error("Expected token to be stored")
	}

	if storedToken.ID != token.ID {
		t.Error("Stored token ID mismatch")
	}
}

func TestAuthManager_ValidateOTP(t *testing.T) {
	am := NewAuthManager()

	// Register a token
	token, err := am.RegisterMasterToken()
	if err != nil {
		t.Fatalf("Failed to register master token: %v", err)
	}

	// Generate OTP
	otp, err := am.GenerateOTPCode(token.ID)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// Validate OTP
	valid := am.ValidateOTP(token.ID, otp)
	if !valid {
		t.Error("Expected OTP to be valid")
	}

	// Test invalid OTP
	invalid := am.ValidateOTP(token.ID, "000000")
	if invalid {
		t.Error("Expected invalid OTP to be rejected")
	}

	// Test non-existent user
	notFound := am.ValidateOTP("non-existent", otp)
	if notFound {
		t.Error("Expected non-existent user to be rejected")
	}
}

func TestAuthManager_GenerateOTPCode(t *testing.T) {
	am := NewAuthManager()

	// Register a token
	token, err := am.RegisterMasterToken()
	if err != nil {
		t.Fatalf("Failed to register master token: %v", err)
	}

	// Generate OTP
	otp, err := am.GenerateOTPCode(token.ID)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	if len(otp) != 6 {
		t.Errorf("Expected OTP length 6, got %d", len(otp))
	}

	// Test non-existent user
	_, err = am.GenerateOTPCode("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestAuthManager_GetQRCodeURL(t *testing.T) {
	am := NewAuthManager()

	// Register a token
	token, err := am.RegisterMasterToken()
	if err != nil {
		t.Fatalf("Failed to register master token: %v", err)
	}

	// Generate QR code URL
	url, err := am.GetQRCodeURL(token.ID, "TestApp", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate QR code URL: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty QR code URL")
	}

	// Test non-existent user
	_, err = am.GetQRCodeURL("non-existent", "TestApp", "test@example.com")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestOTPTimeWindow(t *testing.T) {
	am := NewAuthManager()

	// Register a token
	token, err := am.RegisterMasterToken()
	if err != nil {
		t.Fatalf("Failed to register master token: %v", err)
	}

	// Generate OTP
	otp, err := am.GenerateOTPCode(token.ID)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// OTP should be valid immediately
	valid := am.ValidateOTP(token.ID, otp)
	if !valid {
		t.Error("Expected OTP to be valid immediately after generation")
	}

	// Test with a different OTP (should be invalid)
	invalidOTP := "000000"
	valid = am.ValidateOTP(token.ID, invalidOTP)
	if valid {
		t.Error("Expected invalid OTP to be rejected")
	}
}
