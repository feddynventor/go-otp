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

	srv := server.NewServer()
	log.Printf("Starting OTP server on port %s", port)
	log.Fatal(srv.Run(":" + port))
}
