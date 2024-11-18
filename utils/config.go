package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AESKey        []byte
	ServerAddress string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg := &Config{}
	aesKey := os.Getenv("AES_KEY")
	// Ensure the key is 16 bytes for AES-128
	if len(aesKey) != 16 {
		log.Fatalf("Invalid AES key length: %d", len(aesKey))
	}
	cfg.AESKey = []byte(aesKey)
	serverAddress := os.Getenv("SERVER_ADDRESS")
	// Default to ":8080" if not set
	if len(serverAddress) == 0 {
		serverAddress = ":8080"
	}
	cfg.ServerAddress = serverAddress
	return cfg
}
