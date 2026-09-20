package rdb

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	rdb *redis.Client
}

type SessionStorage interface {
	SetSession(ctx context.Context, sessionKey, userID string) error
	GetKey(ctx context.Context, device string) (string, error)
	Close() error
}

func NewRedisStorage() (*RedisStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	redisAddr := os.Getenv("REDIS_HOST")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	rdb := &RedisStorage{
		rdb: redisClient,
	}

	return rdb, nil
}

func (c *RedisStorage) Close() error {
	return c.rdb.Close()
}
