package sl

import (
	"cinema/internal/lib/env"
	"fmt"
	"log/slog"
	"os"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func WrapErr(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}

func SetupLogger(e env.Env) *slog.Logger {
	var log *slog.Logger

	switch e {
	case env.Local:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case env.Dev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case env.Prod:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
