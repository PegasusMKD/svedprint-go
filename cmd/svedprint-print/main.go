package main

import (
	svedprintprint "github.com/PegasusMKD/svedprint-go/internal/svedprint-print"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Print("Initializing Svedprint Print service...")
	server := svedprintprint.NewServer()
	server.Run()
}
