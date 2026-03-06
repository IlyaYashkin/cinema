package app

import (
	"cinema/internal/catalog/config"
	"log/slog"
)

type App struct {
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	return &App{}
}

func (a *App) MustRun() {}

func (a *App) Stop() {}
