package rdb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSession(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		device := "device-abc"
		key := "shared-secret"
		id := 42

		err := storage.SetSession(ctx, device, key, id)
		require.NoError(t, err)

		val, err := mr.Get(prefixForSession + device)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s%s%d", key, splitSep, id), val)

		ttl := mr.TTL(prefixForSession + device)
		assert.True(t, ttl > 47*time.Hour && ttl <= 48*time.Hour,
			"expected TTL ~48h, got %v", ttl)
	})

	t.Run("overwrite existing session", func(t *testing.T) {
		device := "device-xyz"
		require.NoError(t, storage.SetSession(ctx, device, "old-key", 1))
		require.NoError(t, storage.SetSession(ctx, device, "new-key", 99))

		val, err := mr.Get(prefixForSession + device)
		require.NoError(t, err)
		assert.Equal(t, "new-key/99", val)
	})
}

func TestGetAuthInfo(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		device := "device-1"
		key := "secret-key"
		id := 7

		require.NoError(t, storage.SetSession(ctx, device, key, id))

		gotKey, gotID, err := storage.GetAuthInfo(ctx, device)
		require.NoError(t, err)
		assert.Equal(t, key, gotKey)
		assert.Equal(t, id, gotID)
	})

	t.Run("session not found", func(t *testing.T) {
		gotKey, gotID, err := storage.GetAuthInfo(ctx, "unknown-device")
		require.Error(t, err)
		assert.True(t, errors.Is(err, redis.Nil) || err != nil)
		assert.Equal(t, "", gotKey)
		assert.Equal(t, 0, gotID)
	})

	t.Run("invalid session format - no separator", func(t *testing.T) {
		device := "bad-format-1"
		mr.Set(prefixForSession+device, "onlykey")

		gotKey, gotID, err := storage.GetAuthInfo(ctx, device)
		require.Error(t, err)
		assert.Equal(t, "session is not valid", err.Error())
		assert.Equal(t, "", gotKey)
		assert.Equal(t, 0, gotID)
	})

	t.Run("invalid session format - too many parts", func(t *testing.T) {
		device := "bad-format-2"
		mr.Set(prefixForSession+device, "key/1/extra")

		gotKey, gotID, err := storage.GetAuthInfo(ctx, device)
		require.Error(t, err)
		assert.Equal(t, "session is not valid", err.Error())
		assert.Equal(t, "", gotKey)
		assert.Equal(t, 0, gotID)
	})

	t.Run("invalid id - not a number", func(t *testing.T) {
		device := "bad-id"
		mr.Set(prefixForSession+device, "key/notanumber")

		gotKey, gotID, err := storage.GetAuthInfo(ctx, device)
		require.Error(t, err)
		assert.Equal(t, "", gotKey)
		assert.Equal(t, 0, gotID)
	})

	t.Run("key contains separator still works if exactly 2 parts after split", func(t *testing.T) {
		// если в key есть "/", Split даст больше 2 частей → ошибка
		device := "key-with-sep"
		mr.Set(prefixForSession+device, "sec/ret/5")

		_, _, err := storage.GetAuthInfo(ctx, device)
		require.Error(t, err)
		assert.Equal(t, "session is not valid", err.Error())
	})
}
