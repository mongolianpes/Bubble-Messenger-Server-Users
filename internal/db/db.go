package db

import (
	"context"
	"database/sql"
)

const (
	registerUser    = "INSERT INTO users (login, name, password) VALUES ($1, $2, $3)"
	getPassword     = "SELECT password FROM users WHERE login = $1"
	getUserInfo     = "SELECT name, id FROM users WHERE login = $1"
	getUserInfoByID = "SELECT login, name FROM users WHERE id = $1"
	addUserAvatar   = "UPDATE users SET avatar_path = $2 WHERE login = $1"
	GetUserAvatar   = "SELECT avatar_path FROM users WHERE login = $1"
)

func (s *PostgresStorage) RegisterUser(ctx context.Context, login, name, password string) error {
	_, err := s.db.ExecContext(ctx, registerUser, login, name, password)
	return err
}

func (s *PostgresStorage) GetPassword(ctx context.Context, login string) (string, error) {
	password := ""
	if err := s.db.QueryRowContext(ctx, getPassword, login).Scan(&password); err != nil {
		return "", err
	}

	return password, nil
}

func (s *PostgresStorage) GetUserInfo(ctx context.Context, login string) (userName string, userID int, err error) {
	if err = s.db.QueryRowContext(ctx, getUserInfo, login).Scan(&userName, &userID); err != nil {
		return
	}
	return
}

func (s *PostgresStorage) GetInfoByID(ctx context.Context, id int) (login, name string, err error) {
	if err = s.db.QueryRowContext(ctx, getUserInfoByID, id).Scan(&login, &name); err != nil {
		return
	}

	return
}

func (s *PostgresStorage) AddUserAvatar(ctx context.Context, login, storageAvatarPath string) error {
	_, err := s.db.ExecContext(ctx, addUserAvatar, login, storageAvatarPath)
	return err
}

func (s *PostgresStorage) GetUserAvatarStoragePath(ctx context.Context, login string) (string, error) {
	var avatarPath sql.NullString
	if err := s.db.QueryRowContext(ctx, GetUserAvatar, login).Scan(&avatarPath); err != nil {
		return "", err
	}
	if !avatarPath.Valid {
		return "", nil
	}
	return avatarPath.String, nil
}
