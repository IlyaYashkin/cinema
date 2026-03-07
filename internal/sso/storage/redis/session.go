package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Session struct {
	UserId    string
	DeviceId  string
	CreatedAt time.Time
}

func (s *Storage) SaveSession(ctx context.Context, userId, token string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s:%s", userId, token)

	session := Session{
		UserId:    userId,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Storage) IsSessionExists(ctx context.Context, userId, token string) (bool, error) {
	key := fmt.Sprintf("session:%s:%s", userId, token)

	result, err := s.client.Exists(ctx, key).Result()

	return result > 0, err
}

func (s *Storage) DeleteSession(ctx context.Context, userId, token string) error {
	key := fmt.Sprintf("session:%s:%s", userId, token)

	return s.client.Del(ctx, key).Err()
}

func (s *Storage) DeleteAllSessions(ctx context.Context, userId string) error {
	keys, err := s.scanSessionKeys(ctx, userId)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

func (s *Storage) scanSessionKeys(ctx context.Context, userId string) ([]string, error) {
	var keys []string
	pattern := fmt.Sprintf("session:%s:*", userId)
	cursor := uint64(0)

	for {
		result, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, result...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
