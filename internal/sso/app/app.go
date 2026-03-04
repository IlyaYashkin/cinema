package app

import (
	"cinema/internal/lib/jwt"
	"cinema/internal/sso/app/grpc"
	"cinema/internal/sso/config"
	"cinema/internal/sso/services/auth"
	"cinema/internal/sso/storage/postgres"
	"cinema/internal/sso/storage/redis"
	"log/slog"
)

type Connection interface {
	Close()
	MustConnect()
}

type App struct {
	GRPCServer     *grpc.App
	DBConnection   Connection
	SessionStorage Connection
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	conn, err := postgres.New(cfg.DBConfig)
	if err != nil {
		panic("error creating database connection: " + err.Error())
	}

	sessionStorage, err := redis.NewSessionStorage(cfg.RedisConfig)
	if err != nil {
		panic("error creating session storage: " + err.Error())
	}

	jwtGenerator, err := jwt.NewGenerator(cfg.JWTConfig)
	if err != nil {
		panic("error creating jwt generator: " + err.Error())
	}

	authService := auth.New(log, conn, conn, sessionStorage, jwtGenerator)

	grpcApp := grpc.New(log, authService, cfg.GRPCConfig.Port, cfg.Env)

	return &App{
		GRPCServer:     grpcApp,
		DBConnection:   conn,
		SessionStorage: sessionStorage,
	}
}

func (a *App) MustRun() {
	a.DBConnection.MustConnect()
	a.SessionStorage.MustConnect()
	a.GRPCServer.MustRun()
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.SessionStorage.Close()
	a.DBConnection.Close()
}
