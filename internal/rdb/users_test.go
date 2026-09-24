package rdb

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*RedisStorage, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx).Err())

	return &RedisStorage{rdb: client}, mr
}

func TestSaveUser(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		login := "ivan"
		name := "Иван Иванов"

		err := storage.SaveUser(ctx, login, name)
		require.NoError(t, err)

		val, err := mr.Get(prefixForUser + login)
		require.NoError(t, err)
		assert.Equal(t, name, val)

		ttl := mr.TTL(prefixForUser + login)
		assert.True(t, ttl > time.Hour && ttl <= timeToSaveUser,
			"expected TTL ~2h, got %v", ttl)
	})

	t.Run("overwrite existing", func(t *testing.T) {
		login := "petr"
		require.NoError(t, storage.SaveUser(ctx, login, "Пётр"))
		require.NoError(t, storage.SaveUser(ctx, login, "Пётр Петров"))

		val, err := mr.Get(prefixForUser + login)
		require.NoError(t, err)
		assert.Equal(t, "Пётр Петров", val)
	})
}

func TestGetUsers(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("empty result", func(t *testing.T) {
		users, err := storage.GetUsers(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("find users by partial login", func(t *testing.T) {
		require.NoError(t, storage.SaveUser(ctx, "ivan", "Иван"))
		require.NoError(t, storage.SaveUser(ctx, "ivanov", "Иванов"))
		require.NoError(t, storage.SaveUser(ctx, "petr", "Пётр"))

		// ключ без префикса user_ — не должен попасть
		mr.Set("session_ivan", "something")

		users, err := storage.GetUsers(ctx, "ivan")
		require.NoError(t, err)
		require.Len(t, users, 2)

		got := make(map[string]string)
		for _, u := range users {
			got[u.Login] = u.Name
		}

		assert.Equal(t, "Иван", got[prefixForUser+"ivan"])
		assert.Equal(t, "Иванов", got[prefixForUser+"ivanov"])
		assert.NotContains(t, got, prefixForUser+"petr")
		assert.NotContains(t, got, "session_ivan")
	})

	t.Run("only keys with user_ prefix", func(t *testing.T) {
		mr.FlushAll()

		mr.Set("ivan", "без префикса")
		require.NoError(t, storage.SaveUser(ctx, "ivan", "с префиксом"))

		users, err := storage.GetUsers(ctx, "ivan")
		require.NoError(t, err)
		require.Len(t, users, 1)
		assert.Equal(t, prefixForUser+"ivan", users[0].Login)
		assert.Equal(t, "с префиксом", users[0].Name)
	})

	t.Run("empty query returns all user_ keys", func(t *testing.T) {
		mr.FlushAll()

		require.NoError(t, storage.SaveUser(ctx, "a", "A"))
		require.NoError(t, storage.SaveUser(ctx, "b", "B"))
		require.NoError(t, storage.SaveUser(ctx, "c", "C"))

		users, err := storage.GetUsers(ctx, "")
		require.NoError(t, err)
		assert.Len(t, users, 3)
	})
}

func TestSaveUserAndGetUsers(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	login := "alex"
	name := "Алексей"

	require.NoError(t, storage.SaveUser(ctx, login, name))

	users, err := storage.GetUsers(ctx, login)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, prefixForUser+login, users[0].Login)
	assert.Equal(t, name, users[0].Name)
}

func TestAddTTL(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("extends ttl when minutes <= 40", func(t *testing.T) {
		key := "session_test1"
		// TTL = 30 минут → minutes % 60 = 30 <= 40 → должен продлить
		mr.Set(key, "value")
		mr.SetTTL(key, 30*time.Minute)

		AddTTL(ctx, storage.rdb, key)

		ttl := mr.TTL(key)
		// 30m + 1h = 1h30m
		assert.True(t, ttl > 80*time.Minute && ttl <= 90*time.Minute,
			"expected ~90m, got %v", ttl)
	})

	t.Run("does not extend when minutes > 40", func(t *testing.T) {
		key := "session_test2"
		mr.Set(key, "value")
		mr.SetTTL(key, 50*time.Minute)

		before := mr.TTL(key)
		AddTTL(ctx, storage.rdb, key)
		after := mr.TTL(key)

		// не должен менять TTL (с небольшой погрешностью)
		assert.InDelta(t, before.Seconds(), after.Seconds(), 2)
	})

	t.Run("does not extend when ttl >= maxTimeToSaveUser", func(t *testing.T) {
		key := "session_test3"
		mr.Set(key, "value")
		mr.SetTTL(key, maxTimeToSaveUser)

		before := mr.TTL(key)
		AddTTL(ctx, storage.rdb, key)
		after := mr.TTL(key)

		assert.InDelta(t, before.Seconds(), after.Seconds(), 2)
	})

	t.Run("key without ttl - no panic", func(t *testing.T) {
		key := "session_nottl"
		mr.Set(key, "value")
		// TTL = -1 (no expire)

		assert.NotPanics(t, func() {
			AddTTL(ctx, storage.rdb, key)
		})
	})
}
