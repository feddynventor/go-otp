package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type RegisterRequest struct {
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
}

type RegisterResponse struct {
	MasterToken struct {
		ID        string    `json:"id"`
		Secret    string    `json:"secret"`
		CreatedAt time.Time `json:"created_at"`
		IsActive  bool      `json:"is_active"`
	} `json:"master_token"`
	QRCodeURL string `json:"qr_code_url"`
	Secret    string `json:"secret"`
}

type ValidateOTPRequest struct {
	UserID string `json:"user_id"`
	OTP    string `json:"otp"`
}

type ValidateOTPResponse struct {
	Valid bool `json:"valid"`
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Register(issuer, accountName string) (*RegisterResponse, error) {
	req := RegisterRequest{
		Issuer:      issuer,
		AccountName: accountName,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed: %s", string(body))
	}

	var registerResp RegisterResponse
	err = json.Unmarshal(body, &registerResp)
	return &registerResp, err
}

func (c *Client) ValidateOTP(userID, otp string) (*ValidateOTPResponse, error) {
	req := ValidateOTPRequest{
		UserID: userID,
		OTP:    otp,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/validate-otp", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var validateResp ValidateOTPResponse
	err = json.Unmarshal(body, &validateResp)
	return &validateResp, err
}

func (c *Client) GetProtectedData(userID, otp string) (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/protected-data", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-OTP", otp)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) GetStatus(userID, otp string) (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/status", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-OTP", otp)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func generateOTP(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func main() {
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	client := NewClient(baseURL)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== OTP Client ===")
	fmt.Println("Commands:")
	fmt.Println("1. register - Register a new master token")
	fmt.Println("2. generate - Generate OTP for existing user")
	fmt.Println("3. validate - Validate OTP")
	fmt.Println("4. status - Get protected status")
	fmt.Println("5. data - Get protected data")
	fmt.Println("6. quit - Exit")
	fmt.Println()

	var currentUserID, currentSecret string

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(command)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "register":
			if len(parts) < 3 {
				fmt.Println("Usage: register <issuer> <account_name>")
				continue
			}

			resp, err := client.Register(parts[1], parts[2])
			if err != nil {
				fmt.Printf("Registration failed: %v\n", err)
				continue
			}

			currentUserID = resp.MasterToken.ID
			currentSecret = resp.Secret

			fmt.Printf("Registration successful!\n")
			fmt.Printf("User ID: %s\n", resp.MasterToken.ID)
			fmt.Printf("Secret: %s\n", resp.Secret)
			fmt.Printf("QR Code URL: %s\n", resp.QRCodeURL)
			fmt.Println("Save the secret and scan the QR code with your authenticator app.")

		case "generate":
			if currentSecret == "" {
				fmt.Println("No secret available. Please register first or provide secret.")
				fmt.Print("Enter secret: ")
				if !scanner.Scan() {
					continue
				}
				currentSecret = strings.TrimSpace(scanner.Text())
			}

			otp, err := generateOTP(currentSecret)
			if err != nil {
				fmt.Printf("Failed to generate OTP: %v\n", err)
				continue
			}

			fmt.Printf("Generated OTP: %s\n", otp)

		case "validate":
			if len(parts) < 3 {
				fmt.Println("Usage: validate <user_id> <otp>")
				continue
			}

			resp, err := client.ValidateOTP(parts[1], parts[2])
			if err != nil {
				fmt.Printf("Validation failed: %v\n", err)
				continue
			}

			if resp.Valid {
				fmt.Println("OTP is valid!")
			} else {
				fmt.Println("OTP is invalid!")
			}

		case "status":
			if currentUserID == "" {
				fmt.Println("No user ID available. Please register first.")
				continue
			}

			otp, err := generateOTP(currentSecret)
			if err != nil {
				fmt.Printf("Failed to generate OTP: %v\n", err)
				continue
			}

			status, err := client.GetStatus(currentUserID, otp)
			if err != nil {
				fmt.Printf("Failed to get status: %v\n", err)
				continue
			}

			fmt.Printf("Status: %s\n", status)

		case "data":
			if currentUserID == "" {
				fmt.Println("No user ID available. Please register first.")
				continue
			}

			otp, err := generateOTP(currentSecret)
			if err != nil {
				fmt.Printf("Failed to generate OTP: %v\n", err)
				continue
			}

			data, err := client.GetProtectedData(currentUserID, otp)
			if err != nil {
				fmt.Printf("Failed to get protected data: %v\n", err)
				continue
			}

			fmt.Printf("Protected data: %s\n", data)

		case "quit", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Type 'quit' to exit.")
		}
	}
}
