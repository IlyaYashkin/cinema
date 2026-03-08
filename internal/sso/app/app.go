package app

import (
	"cinema/internal/lib/grpc"
	"cinema/internal/lib/jwt"
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/redis"
	"cinema/internal/lib/smtp"
	"cinema/internal/sso/config"
	grpcAuth "cinema/internal/sso/grpc/auth"
	ssoSmtp "cinema/internal/sso/notification/smtp"
	"cinema/internal/sso/services/auth"
	ssoPostgres "cinema/internal/sso/storage/postgres"
	ssoRedis "cinema/internal/sso/storage/redis"
	"log/slog"
)

type Connection interface {
	Close()
	MustConnect()
}

type App struct {
	GRPCServer      *grpc.App
	DBConnection    Connection
	RedisConnection Connection
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	dbConn, err := postgres.New(cfg.DBConfig)
	if err != nil {
		panic("error creating database connection: " + err.Error())
	}
	userProvider := &ssoPostgres.User{Postgres: dbConn}

	redisConn, err := redis.New(cfg.RedisConfig)
	if err != nil {
		panic("error creating redis userProvider: " + err.Error())
	}
	sessionStorage := &ssoRedis.Session{Redis: redisConn}
	resetTokenStorage := &ssoRedis.Reset{Redis: redisConn}

	jwtGenerator, err := jwt.NewGenerator(cfg.JWTConfig)
	if err != nil {
		panic("error creating jwt generator: " + err.Error())
	}

	smtpClient, err := smtp.New(cfg.SMTPConfig)
	if err != nil {
		panic("error creating smtp client: " + err.Error())
	}

	emailSender := ssoSmtp.NewEmailSender(smtpClient, cfg.ResetPasswordBaseUrl)

	authService := auth.New(log, userProvider, sessionStorage, resetTokenStorage, jwtGenerator, emailSender, cfg.ResetTokenTTL)

	grpcApp := grpc.New(log, grpcAuth.NewController(authService), cfg.GRPCConfig.Port, cfg.Env)

	return &App{
		GRPCServer:      grpcApp,
		DBConnection:    dbConn,
		RedisConnection: redisConn,
	}
}

func (a *App) MustRun() {
	a.DBConnection.MustConnect()
	a.RedisConnection.MustConnect()
	a.GRPCServer.MustRun()
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.RedisConnection.Close()
	a.DBConnection.Close()
}
