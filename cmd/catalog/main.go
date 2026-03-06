package main

import (
	"cinema/internal/catalog/app"
	catalogConfig "cinema/internal/catalog/config"
	libconfig "cinema/internal/lib/config"
	"cinema/internal/lib/env"
	"cinema/internal/lib/shutdown"
	"cinema/internal/lib/sl"
)

func main() {
	env.Load()

	cfg := libconfig.MustLoad[catalogConfig.Config]()

	log := sl.SetupLogger(cfg.Env)

	application := app.New(log, cfg)

	go application.MustRun()

	shutdown.WaitForShutdown()

	application.Stop()
	log.Info("Gracefully stopped")
}
