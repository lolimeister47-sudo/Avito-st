package main

import (
	"log"

	"prservice/internal/app"
	"prservice/internal/config"
)

func main() {
	cfg := config.Load()

	if err := app.Run(cfg); err != nil {
		log.Fatalf("app terminated: %v", err)
	}
}
