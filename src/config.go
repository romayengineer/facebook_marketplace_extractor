package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type UserCredentials struct {
	Username string
	Password string
}

type Config struct {
	UserCredentials UserCredentials
}

func NewConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	// Read USERNAME and PASSWORD from .env
	username := os.Getenv("FME_USERNAME")
	password := os.Getenv("FME_PASSWORD")

	if username == "" || password == "" {
		Log(LE0, "USERNAME or PASSWORD not set in .env file")
		os.Exit(1)
	}

	fmt.Printf("Loaded credentials - Username: %s\n", username)

	config := Config{
		UserCredentials: UserCredentials{
			Username: username,
			Password: password,
		},
	}

	return &config, nil
}
