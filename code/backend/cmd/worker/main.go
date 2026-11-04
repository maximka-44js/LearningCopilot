package main

import (
	"log"

	"github.com/maximka-44js/lc/config"
	workerruntime "github.com/maximka-44js/lc/workers/runtime"
)

func main() {
	cfg, err := config.NewWorkerConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	workerruntime.Run(cfg)
}
