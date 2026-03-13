package postgres

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/sl"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func New(log *slog.Logger, config config.DBConfig) (*Postgres, error) {
	const op = "lib.postgres.new"

	pgxCfg, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	pgxCfg.MaxConns = config.MaxConns
	pgxCfg.MinConns = config.MinConns
	pgxCfg.MaxConnLifetime = config.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	return &Postgres{log: log, pool: pool}, nil
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) MustConnect() {
	const op = "lib.postgres.must_connect"

	if err := p.pool.Ping(context.Background()); err != nil {
		panic(sl.WrapErr(op, err))
	}

	p.log.With(slog.String("op", op)).Info("postgres connected")
}

func (p *Postgres) Close() {
	const op = "lib.postgres.close"

	p.pool.Close()

	p.log.With(slog.String("op", op)).Info("postgres disconnected")
}
