package redis

import (
	"cinema/internal/lib/config"
	"context"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(config config.RedisConfig) (*Redis, error) {
	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	return &Redis{client: client}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) MustConnect() {
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		panic("error pinging redis: " + err.Error())
	}
}

func (r *Redis) Close() {
	_ = r.client.Close()
}
