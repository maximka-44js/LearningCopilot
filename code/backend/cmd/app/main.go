package main

import (
	"log"

	"github.com/maximka-44js/lc/config"
	"github.com/maximka-44js/lc/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
