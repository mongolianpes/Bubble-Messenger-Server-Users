package rdb

import (
	"context"
	"fmt"
	"strings"
	"time"
	pb "users/proto"
)

const (
	prefixForUser  = "user_"
	timeToSaveUser = time.Hour * 4
)

func (s *RedisStorage) GetUsers(ctx context.Context, login string) (result []*pb.UserInfo, err error) {
	iter := s.rdb.Scan(ctx, 0, fmt.Sprintf("*%s*", login), 0).Iterator()

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
	return s.rdb.Set(ctx, login, name, timeToSaveSession).Err()
}
