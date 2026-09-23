package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDSN() string {
	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "5432")
	user := envOrDefault("TEST_DB_USER", "postgres")
	password := envOrDefault("TEST_DB_PASSWORD", "123")
	dbname := envOrDefault("TEST_DB_NAME", "project_farm")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupTestDB(t *testing.T) (*sql.DB, *PostgresStorage) {
	t.Helper()

	db, err := sql.Open("postgres", getTestDSN())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx))

	_, err = db.ExecContext(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	storage := &PostgresStorage{db: db}
	return db, storage
}

func generateTestHashedPass() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	hashedInputPassword := make([]byte, 192)
	for i := range hashedInputPassword {
		hashedInputPassword[i] = letters[rand.IntN(len(letters))]
	}
	return string(hashedInputPassword)
}

func TestPostgresStorage_RegisterUser(t *testing.T) {
	db, storage := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		hashedInputPassword := generateTestHashedPass()

		err := storage.RegisterUser(ctx, "ivan", "Иван", hashedInputPassword)
		require.NoError(t, err)

		var login, name, password string
		err = db.QueryRowContext(ctx,
			`SELECT login, name, password FROM users WHERE login = $1`, "ivan",
		).Scan(&login, &name, &password)
		require.NoError(t, err)

		assert.Equal(t, "ivan", login)
		assert.Equal(t, "Иван", name)
		assert.Equal(t, hashedInputPassword, password)
	})

	t.Run("duplicate login", func(t *testing.T) {
		err := storage.RegisterUser(ctx, "petr", "Пётр", "pass1")
		require.NoError(t, err)

		err = storage.RegisterUser(ctx, "petr", "Пётр2", "pass2")
		require.Error(t, err)
	})
}

func TestPostgresStorage_GetPassword(t *testing.T) {
	db, storage := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	hashedPass := generateTestHashedPass()

	t.Run("success", func(t *testing.T) {
		require.NoError(t, storage.RegisterUser(ctx, "anna", "Анна", hashedPass))

		password, err := storage.GetPassword(ctx, "anna")
		require.NoError(t, err)
		assert.Equal(t, hashedPass, password)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := storage.GetPassword(ctx, "nobody")
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestPostgresStorage_GetUserInfo(t *testing.T) {
	db, storage := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	hashedPass := generateTestHashedPass()

	t.Run("success", func(t *testing.T) {
		require.NoError(t, storage.RegisterUser(ctx, "olga", "Ольга", hashedPass))

		name, err := storage.GetUserInfo(ctx, "olga")
		require.NoError(t, err)
		assert.Equal(t, "Ольга", name)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := storage.GetUserInfo(ctx, "ghost")
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
