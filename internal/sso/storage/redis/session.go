package redis

import (
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Storage) SaveSession(ctx context.Context, userId, token, deviceId, deviceName string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

	session := domain.Session{
		RefreshToken: token,
		DeviceName:   deviceName,
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Storage) GetSession(ctx context.Context, userId, deviceId string) (domain.Session, error) {
	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.Session{}, storage.ErrSessionNotFound
		}

		return domain.Session{}, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *Storage) DeleteSession(ctx context.Context, userId, deviceId string) error {
	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

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
