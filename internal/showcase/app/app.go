package app

import (
	"cinema/internal/lib/s3"
	"cinema/internal/showcase/config"
	"context"
	"log/slog"
)

type MustConnectionChecker interface {
	MustConnect()
}

type App struct {
	S3Connection *s3.S3
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	s3Conn, err := s3.New(log, cfg.S3Config)
	if err != nil {
		panic("failed to create S3 connection: " + err.Error())
	}

	return &App{S3Connection: s3Conn}
}

func (a *App) MustRun() {
	ctx := context.Background()

	a.S3Connection.MustConnect()
	a.S3Connection.MustCreateBucketIfNotExists(ctx)
	a.S3Connection.MustSetupTemporaryFilesLifecycleConfigurationIfNotExists(ctx)
}

func (a *App) Stop() {}
