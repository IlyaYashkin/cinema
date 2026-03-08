package redis

import (
	redislib "cinema/internal/lib/redis"
	"context"
	"fmt"
	"time"
)

type Reset struct {
	*redislib.Redis
}

func (r *Reset) SaveResetToken(ctx context.Context, userId, token string, ttl time.Duration) error {
	key := fmt.Sprintf("reset:%s", token)

	return r.Client().Set(ctx, key, userId, ttl).Err()
}

func (r *Reset) GetUserIdByResetToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("reset:%s", token)

	userId, err := r.Client().Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (r *Reset) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("reset:%s", token)

	return r.Client().Del(ctx, key).Err()
}
