package redis

import (
	"context"
	"fmt"
	"time"
)

func (s *Storage) SaveResetToken(ctx context.Context, userId, token string, ttl time.Duration) error {
	key := fmt.Sprintf("reset:%s", token)

	return s.client.Set(ctx, key, userId, ttl).Err()
}

func (s *Storage) GetUserIdByResetToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("reset:%s", token)

	userId, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *Storage) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("reset:%s", token)

	return s.client.Del(ctx, key).Err()
}
