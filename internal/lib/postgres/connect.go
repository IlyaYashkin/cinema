package postgres

import (
	"cinema/internal/lib/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(config config.DBConfig) (*Postgres, error) {
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

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) MustConnect() {
	if err := p.pool.Ping(context.Background()); err != nil {
		panic("error pinging database: " + err.Error())
	}
}

func (p *Postgres) Close() {
	p.pool.Close()
}
