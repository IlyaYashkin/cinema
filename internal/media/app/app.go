package app

import (
	"cinema/internal/media/config"
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

func (app *App) MustRun() {

}

func (app *App) Stop() {}
