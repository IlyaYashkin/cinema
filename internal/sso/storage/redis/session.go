package redis

import (
	redislib "cinema/internal/lib/redis"
	"cinema/internal/lib/sl"
	"cinema/internal/sso/domain"
	"cinema/internal/sso/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Session struct {
	*redislib.Redis
}

func (s *Session) SaveSession(ctx context.Context, userId, token, deviceId, deviceName string, ttl time.Duration) error {
	const op = "sso.storage.session.save"

	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

	session := domain.Session{
		RefreshToken: token,
		DeviceName:   deviceName,
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	err = s.Client().Set(ctx, key, data, ttl).Err()
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *Session) GetSession(ctx context.Context, userId, deviceId string) (domain.Session, error) {
	const op = "sso.storage.session.get"

	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

	data, err := s.Client().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.Session{}, sl.WrapErr(op, storage.ErrSessionNotFound)
		}

		return domain.Session{}, sl.WrapErr(op, err)
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return domain.Session{}, sl.WrapErr(op, err)
	}

	return session, nil
}

func (s *Session) DeleteSession(ctx context.Context, userId, deviceId string) error {
	const op = "sso.storage.session.delete"

	key := fmt.Sprintf("session:%s:%s", userId, deviceId)

	err := s.Client().Del(ctx, key).Err()
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *Session) DeleteAllSessions(ctx context.Context, userId string) error {
	const op = "sso.storage.session.delete_all"

	keys, err := s.scanSessionKeys(ctx, userId)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	if len(keys) == 0 {
		return nil
	}
	err = s.Client().Del(ctx, keys...).Err()
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *Session) scanSessionKeys(ctx context.Context, userId string) ([]string, error) {
	const op = "sso.storage.session.scan"

	var keys []string
	pattern := fmt.Sprintf("session:%s:*", userId)
	cursor := uint64(0)

	for {
		result, nextCursor, err := s.Client().Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, sl.WrapErr(op, err)
		}

		keys = append(keys, result...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
