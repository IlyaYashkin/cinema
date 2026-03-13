package redis

import (
	redislib "cinema/internal/lib/redis"
	"cinema/internal/lib/sl"
	"cinema/internal/sso/storage"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Reset struct {
	*redislib.Redis
}

func (r *Reset) SaveResetToken(ctx context.Context, userId, token string, ttl time.Duration) error {
	const op = "sso.storage.reset.reset_token"

	key := fmt.Sprintf("reset:%s", token)

	err := r.Client().Set(ctx, key, userId, ttl).Err()
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (r *Reset) GetUserIdByResetToken(ctx context.Context, token string) (string, error) {
	const op = "sso.storage.reset.get_user_by_token"

	key := fmt.Sprintf("reset:%s", token)

	userId, err := r.Client().Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", sl.WrapErr(op, storage.ErrResetTokenNotFound)
		}

		return "", sl.WrapErr(op, err)
	}

	return userId, nil
}

func (r *Reset) DeleteResetToken(ctx context.Context, token string) error {
	const op = "sso.storage.reset.delete"

	key := fmt.Sprintf("reset:%s", token)

	err := r.Client().Del(ctx, key).Err()
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}
