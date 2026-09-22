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

	storage := &RedisStorage{rdb: client}
	return storage, mr
}

func TestSaveUser(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("successful save", func(t *testing.T) {
		login := "ivan"
		name := "Иван Иванов"

		err := storage.SaveUser(ctx, login, name)
		require.NoError(t, err)

		val, err := mr.Get(prefixForUser + login)
		require.NoError(t, err)
		assert.Equal(t, name, val)

		ttl := mr.TTL(prefixForUser + login)
		assert.True(t, ttl > 3*time.Hour && ttl <= 4*time.Hour, "expected TTL around 4h, got %v", ttl)
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		login := "petr"
		oldName := "Пётр"
		newName := "Пётр Петров"

		require.NoError(t, storage.SaveUser(ctx, login, oldName))
		require.NoError(t, storage.SaveUser(ctx, login, newName))

		val, err := mr.Get(prefixForUser + login)
		require.NoError(t, err)
		assert.Equal(t, newName, val)
	})

	t.Run("empty login and name", func(t *testing.T) {
		err := storage.SaveUser(ctx, "", "")
		require.NoError(t, err)

		val, err := mr.Get(prefixForUser)
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})
}

func TestGetUsers(t *testing.T) {
	storage, mr := setupTestRedis(t)
	defer mr.Close()
	defer storage.Close()

	ctx := context.Background()

	t.Run("empty result when no matching keys", func(t *testing.T) {
		users, err := storage.GetUsers(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("find users by partial login", func(t *testing.T) {
		mr.Set("user_ivan", "Иван")
		mr.Set("user_ivanov", "Иванов")
		mr.Set("user_petr", "Пётр")
		mr.Set("session_123", "some-session")
		mr.Set("other_key", "value")

		users, err := storage.GetUsers(ctx, "ivan")
		require.NoError(t, err)

		require.Len(t, users, 2)

		got := make(map[string]string)
		for _, u := range users {
			got[u.Login] = u.Name
		}

		assert.Equal(t, "Иван", got["user_ivan"])
		assert.Equal(t, "Иванов", got["user_ivanov"])
		assert.NotContains(t, got, "user_petr")
		assert.NotContains(t, got, "session_123")
	})

	t.Run("exact match", func(t *testing.T) {
		mr.FlushAll()

		mr.Set("user_alex", "Алексей")
		mr.Set("user_alexander", "Александр")

		users, err := storage.GetUsers(ctx, "user_alex")
		require.NoError(t, err)

		require.Len(t, users, 2)

		got := make(map[string]string)
		for _, u := range users {
			got[u.Login] = u.Name
		}
		assert.Equal(t, "Алексей", got["user_alex"])
		assert.Equal(t, "Александр", got["user_alexander"])
	})

	t.Run("no prefix match - key without user_ is ignored", func(t *testing.T) {
		mr.FlushAll()

		mr.Set("ivan", "Иван без префикса")
		mr.Set("user_ivan", "Иван с префиксом")

		users, err := storage.GetUsers(ctx, "ivan")
		require.NoError(t, err)

		require.Len(t, users, 1)
		assert.Equal(t, "user_ivan", users[0].Login)
		assert.Equal(t, "Иван с префиксом", users[0].Name)
	})

	t.Run("multiple scans and empty pattern edge", func(t *testing.T) {
		mr.FlushAll()

		mr.Set("user_a", "A")
		mr.Set("user_b", "B")
		mr.Set("user_c", "C")

		users, err := storage.GetUsers(ctx, "")
		require.NoError(t, err)
		assert.Len(t, users, 3)
	})
}
