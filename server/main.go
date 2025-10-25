package main

import (
	"log"
	"os"

	"otp-basic/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Starting OTP server on port %s", port)
	log.Fatal(srv.Run(":" + port))
}
