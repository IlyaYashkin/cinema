package redis

import (
	"cinema/internal/lib/config"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	log    *slog.Logger
	client *redis.Client
}

func New(log *slog.Logger, config config.RedisConfig) (*Redis, error) {
	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	return &Redis{log: log, client: client}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) MustConnect() {
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		panic("error pinging redis: " + err.Error())
	}
	r.log.Info("redis connected")
}

func (r *Redis) Close() {
	_ = r.client.Close()

	r.log.Info("redis disconnected")
}
