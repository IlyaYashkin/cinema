package redis

import (
	"cinema/internal/sso/config"
	"context"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

func NewStorage(config config.RedisConfig) (*Storage, error) {
	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	return &Storage{client: client}, nil
}

func (s *Storage) MustConnect() {
	if err := s.client.Ping(context.Background()).Err(); err != nil {
		panic("error pinging redis: " + err.Error())
	}
}

func (s *Storage) Close() {
	_ = s.client.Close()
}
