package rdb

import (
	"context"
	"fmt"
	"strings"
	"time"
	pb "users/proto"

	"github.com/redis/go-redis/v9"
)

const (
	prefixForUser     = "user_"
	timeToSaveUser    = time.Hour * 2
	addTimeToSaveUser = time.Hour
	maxTimeToSaveUser = time.Hour * 12
)

func (s *RedisStorage) GetUsers(ctx context.Context, login string) (result []*pb.UserInfo, err error) {
	iter := s.rdb.Scan(ctx, 0, fmt.Sprintf("*%s*", login), 0).Iterator()

	go AddTTL(ctx, s.rdb, prefixForSession+login)

	for iter.Next(ctx) {
		key := iter.Val()

		if strings.HasPrefix(key, prefixForUser) {
			name := ""
			name, err = s.rdb.Get(ctx, key).Result()
			if err != nil {
				return
			}

			result = append(result, &pb.UserInfo{
				Login: key,
				Name:  name,
			})
		}
	}

	if err = iter.Err(); err != nil {
		return
	}

	return
}

func (s *RedisStorage) SaveUser(ctx context.Context, login, name string) error {
	return s.rdb.Set(ctx, prefixForUser+login, name, timeToSaveUser).Err()
}

func (s *RedisStorage) GetUser(ctx context.Context, login string) (string, error) {
	go AddTTL(ctx, s.rdb, login)
	return s.rdb.Get(ctx, login).Result()
}

func AddTTL(ctx context.Context, rdb *redis.Client, key string) {
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		return
	}

	if ttl >= maxTimeToSaveUser {
		return
	}

	minutes := int(ttl.Minutes()) % 60
	if minutes <= 40 {
		rdb.Expire(ctx, key, ttl+addTimeToSaveUser)
	}
}
