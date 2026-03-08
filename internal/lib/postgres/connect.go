package postgres

import (
	"cinema/internal/lib/config"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func New(log *slog.Logger, config config.DBConfig) (*Postgres, error) {
	pgxCfg, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, err
	}

	pgxCfg.MaxConns = config.MaxConns
	pgxCfg.MinConns = config.MinConns
	pgxCfg.MaxConnLifetime = config.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		return nil, err
	}

	return &Postgres{log: log, pool: pool}, nil
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) MustConnect() {
	if err := p.pool.Ping(context.Background()); err != nil {
		panic("error pinging database: " + err.Error())
	}
	p.log.Info("postgres connected")
}

func (p *Postgres) Close() {
	p.pool.Close()

	p.log.Info("postgres disconnected")
}
