package app

import (
	"cinema/internal/lib/grpc"
	"cinema/internal/lib/jwt"
	"cinema/internal/lib/postgres"
	"cinema/internal/lib/redis"
	"cinema/internal/lib/sl"
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

type ConnectionChecker interface {
	Connect()
}

type App struct {
	GRPCServer      *grpc.App
	DBConnection    Connection
	RedisConnection Connection
	SMTPConnection  ConnectionChecker
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	const op = "sso.app.new"

	dbConn, err := postgres.New(log, cfg.DBConfig)
	if err != nil {
		panic(sl.WrapErr(op, err))
	}
	userProvider := &ssoPostgres.User{Postgres: dbConn}

	redisConn, err := redis.New(log, cfg.RedisConfig)
	if err != nil {
		panic(sl.WrapErr(op, err))
	}
	sessionStorage := &ssoRedis.Session{Redis: redisConn}
	resetTokenStorage := &ssoRedis.Reset{Redis: redisConn}

	jwtGenerator, err := jwt.NewGenerator(cfg.JWTConfig)
	if err != nil {
		panic(sl.WrapErr(op, err))
	}

	smtpClient, err := smtp.New(log, cfg.SMTPConfig, cfg.Env)
	if err != nil {
		panic(sl.WrapErr(op, err))
	}

	emailSender := ssoSmtp.NewEmailSender(smtpClient, cfg.ResetPasswordBaseUrl)

	authService := auth.New(log, userProvider, sessionStorage, resetTokenStorage, jwtGenerator, emailSender, cfg.ResetTokenTTL)

	grpcApp := grpc.New(log, cfg.GRPCConfig.Port, cfg.Env)
	grpcApp.Register(grpcAuth.NewController(authService))

	return &App{
		GRPCServer:      grpcApp,
		DBConnection:    dbConn,
		RedisConnection: redisConn,
		SMTPConnection:  smtpClient,
	}
}

func (a *App) MustRun() {
	a.DBConnection.MustConnect()
	a.RedisConnection.MustConnect()
	a.SMTPConnection.Connect()
	a.GRPCServer.MustRun()
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.RedisConnection.Close()
	a.DBConnection.Close()
}
