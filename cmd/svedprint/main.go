package main

import (
	"github.com/PegasusMKD/svedprint-go/internal/svedprint"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Print("Initializing Svedprint service...")
	server := svedprint.NewServer()
	server.Run()
}
