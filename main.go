package main

import (
	"log"

	"github.com/KristinaAlekseeva11/finalgo/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
