package app

import (
	"cinema/internal/lib/grpc"
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/s3"
	"cinema/internal/showcase/config"
	showcaseS3 "cinema/internal/showcase/file/s3"
	"cinema/internal/showcase/grpc/film"
	filmService "cinema/internal/showcase/services/film"
	showcasePostgres "cinema/internal/showcase/storage/postgres"
	"log/slog"
)

type Connection interface {
	Close()
	MustConnect()
}

type MustConnectionChecker interface {
	MustConnect()
}

type App struct {
	DBConnection Connection
	S3Connection MustConnectionChecker
	GRPCServer   *grpc.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	dbConn, err := postgres.New(log, cfg.DBConfig)
	if err != nil {
		panic("error creating database connection: " + err.Error())
	}
	filmStorage := &showcasePostgres.Film{Postgres: dbConn}

	s3Conn, err := s3.New(log, cfg.S3Config)
	if err != nil {
		panic("failed to create S3 connection: " + err.Error())
	}
	fileStorage := &showcaseS3.FileStorage{S3: s3Conn}

	filmSrv := filmService.New(log, filmStorage, fileStorage)

	grpcApp := grpc.New(log, cfg.GRPCConfig.Port, cfg.Env)
	grpcApp.Register(film.NewController(filmSrv))

	return &App{DBConnection: dbConn, S3Connection: s3Conn, GRPCServer: grpcApp}
}

func (a *App) MustRun() {
	a.DBConnection.MustConnect()
	a.S3Connection.MustConnect()
	a.GRPCServer.MustRun()
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.DBConnection.Close()
}
