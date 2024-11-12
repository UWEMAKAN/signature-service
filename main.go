package main

import (
	"log"

	"github.com/uwemakan/signing-service/api"
)

const (
	ListenAddress = ":8080"
	// TODO: add further configuration parameters here ...
)

func main() {
	server := api.NewServer(ListenAddress)
	log.Default().Println("Starting server on ", ListenAddress)

	if err := server.Run(); err != nil {
		log.Fatal("Could not start server on ", ListenAddress)
	}
	log.Default().Println("Starting server on ", ListenAddress)
}
