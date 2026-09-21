package rdb

import (
	"context"
	"time"
)

const (
	prefixForSession  = "session_"
	timeToSaveSession = time.Hour * 48
)

func (s *RedisStorage) SetSession(ctx context.Context, device, key string) error {
	err := s.rdb.Set(ctx, prefixForSession+device, key, timeToSaveSession).Err()
	return err
}

func (s *RedisStorage) GetKey(ctx context.Context, device string) (string, error) {
	return s.rdb.Get(ctx, prefixForSession+device).Result()
}
