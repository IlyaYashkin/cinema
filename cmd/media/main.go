package main

import (
	libconfig "cinema/internal/lib/config"
	"cinema/internal/lib/env"
	"cinema/internal/lib/shutdown"
	"cinema/internal/lib/sl"
	"cinema/internal/media/app"
	mediaConfig "cinema/internal/media/config"
)

func main() {
	env.Load()

	cfg := libconfig.MustLoad[mediaConfig.Config]()

	log := sl.SetupLogger(cfg.Env)

	application := app.New(log, cfg)

	go application.MustRun()

	shutdown.WaitForShutdown()

	application.Stop()
	log.Info("gracefully stopped")
}
