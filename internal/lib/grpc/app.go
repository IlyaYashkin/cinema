package grpc

import (
	"cinema/internal/lib/env"
	"cinema/internal/lib/sl"
	"context"
	"fmt"
	"log/slog"
	"net"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidateInterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

type Registrar interface {
	RegisterGRPCServer(gRPCServer *grpc.Server)
}

func New(log *slog.Logger, port int, e env.Env) *App {
	const op = "lib.grpc.new"

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	validator, err := protovalidate.New()
	if err != nil {
		panic(sl.WrapErr(op, err))
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
		protovalidateInterceptor.UnaryServerInterceptor(validator),
	))

	if e.Is(env.Local) {
		reflection.Register(gRPCServer)
	}

	return &App{log: log, gRPCServer: gRPCServer, port: port}
}

func (a *App) Register(registrar Registrar) {
	registrar.RegisterGRPCServer(a.gRPCServer)
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "lib.grpc.run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return sl.WrapErr(op, err)
	}

	a.log.With(slog.String("op", op)).
		Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "lib.grpc.stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
