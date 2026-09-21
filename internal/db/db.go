package db

import "context"

const (
	registerUser = "INSERT INTO users (login, name, password) VALUES ($1, $2, $3)"
	getPassword  = "SELECT password FROM users WHERE login = $1"
	getUserInfo  = "SELECT login, name FROM users WHRER login = $1"
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

func (s *PostgresStorage) GetUserInfo(ctx context.Context, login string) (string, error) {
	userName := ""
	if err := s.db.QueryRowContext(ctx, getUserInfo, login).Scan(&userName); err != nil {
		return "", err
	}
	return userName, nil
}
