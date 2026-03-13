package redis

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/sl"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	log    *slog.Logger
	client *redis.Client
}

func New(log *slog.Logger, config config.RedisConfig) (*Redis, error) {
	const op = "lib.redis.new"

	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	client := redis.NewClient(opts)

	return &Redis{log: log, client: client}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) MustConnect() {
	const op = "lib.redis.must_connect"

	if err := r.client.Ping(context.Background()).Err(); err != nil {
		panic(sl.WrapErr(op, err))
	}
	r.log.With(slog.String("op", op)).Info("redis connected")
}

func (r *Redis) Close() {
	const op = "lib.redis.close"

	_ = r.client.Close()

	r.log.With(slog.String("op", op)).Info("redis disconnected")
}
