package rdb

import (
	"context"
	"time"
)

const (
	timeToSaveSessionInRedis = time.Hour * 48
)

func (c *RedisStorage) SetSession(ctx context.Context, device, key string) error {
	err := c.rdb.Set(ctx, device, key, timeToSaveSessionInRedis).Err()
	return err
}

func (c *RedisStorage) GetKey(ctx context.Context, device string) (string, error) {
	return c.rdb.Get(ctx, device).Result()
}
