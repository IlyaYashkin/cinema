package app

import (
	"cinema/internal/lib/s3"
	"cinema/internal/showcase/config"
	"log/slog"
)

type MustConnectionChecker interface {
	MustConnect()
}

type App struct {
	S3Connection MustConnectionChecker
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	s3Conn, err := s3.New(log, cfg.S3Config)
	if err != nil {
		panic(err)
	}

	return &App{S3Connection: s3Conn}
}

func (a *App) MustRun() {
	a.S3Connection.MustConnect()
}

func (a *App) Stop() {}
