package main

import (
	svedprintadmin "github.com/PegasusMKD/svedprint-go/internal/svedprint-admin"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Print("Initializing Svedprint service...")
	server := svedprintadmin.NewServer()
	server.Run()
}
