package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

type PostgresStorage struct {
	db *sql.DB
}

type UsersStorage interface {
	RegisterUser(ctx context.Context, login, name, password string) error
	GetPassword(ctx context.Context, login string) (string, error)
}

func NewPostgresStorage() (*PostgresStorage, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func connectToDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
