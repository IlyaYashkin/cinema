package postgres

import (
	"cinema/internal/lib/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(config config.DBConfig) (*Storage, error) {
	pgxconf, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, err
	}

	pgxconf.MaxConns = config.MaxConns
	pgxconf.MinConns = config.MinConns
	pgxconf.MaxConnLifetime = config.MaxConnLifetime
	pgxconf.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxconf)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) MustConnect() {
	if err := s.pool.Ping(context.Background()); err != nil {
		panic("error pinging database: " + err.Error())
	}
}

func (s *Storage) Close() {
	s.pool.Close()
}
