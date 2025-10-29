package main

import (
	"github.com/PegasusMKD/svedprint-go/internal/gateway"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Print("Initializing Svedprint Gateway service...")
	server := gateway.NewServer()
	server.Run()
}
