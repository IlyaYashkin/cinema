package app

import (
	"cinema/internal/lib/file/s3"
	"cinema/internal/lib/grpc"
	"cinema/internal/lib/postgres"
	"cinema/internal/media/config"
	mediaController "cinema/internal/media/grpc/content"
	"cinema/internal/media/service/content"
	mediaPostgres "cinema/internal/media/storage/postgres"
	"log/slog"
)

type Connection interface {
	Close()
	MustConnect()
}

type App struct {
	DBConnection Connection
	S3Connection Connection
	GRPCServer   *grpc.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	dbConn, err := postgres.New(log, cfg.DBConfig)
	if err != nil {
		panic(err)
	}
	op := &mediaPostgres.Original{Postgres: dbConn}

	s3Conn, err := s3.New(log, cfg.S3Config)
	if err != nil {
		panic(err)
	}

	mediaSrv := content.New(log, s3Conn, op)

	grpcApp := grpc.New(log, cfg.GRPCConfig.Port, cfg.Env)
	grpcApp.Register(mediaController.NewController(mediaSrv))

	return &App{DBConnection: dbConn, S3Connection: s3Conn, GRPCServer: grpcApp}
}

func (a *App) MustRun() {
	a.DBConnection.MustConnect()
	a.S3Connection.MustConnect()
	a.GRPCServer.MustRun()
}

func (a *App) Stop() {
	a.DBConnection.Close()
	a.GRPCServer.Stop()
}
