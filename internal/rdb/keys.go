package rdb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	prefixForSession  = "session_"
	timeToSaveSession = time.Hour * 48
	splitSep          = "/"
)

func (s *RedisStorage) SetSession(ctx context.Context, device, key string, id int) error {
	err := s.rdb.Set(ctx, prefixForSession+device, fmt.Sprintf("%v%v%v", key, splitSep, id), timeToSaveSession).Err()
	return err
}

func (s *RedisStorage) GetAuthInfo(ctx context.Context, device string) (string, int, error) {
	info, err := s.rdb.Get(ctx, prefixForSession+device).Result()
	if err != nil {
		return "", 0, err
	}

	splittedInfo := strings.Split(info, splitSep)
	if len(splittedInfo) != 2 {
		return "", 0, errors.New("session is not valid")
	}

	id, err := strconv.Atoi(splittedInfo[1])
	if err != nil {
		return "", 0, err
	}

	return splittedInfo[0], id, nil
}
