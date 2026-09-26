package db

import "context"

const (
	registerUser    = "INSERT INTO users (login, name, password) VALUES ($1, $2, $3)"
	getPassword     = "SELECT password FROM users WHERE login = $1"
	getUserInfo     = "SELECT name, id FROM users WHERE login = $1"
	getUserInfoByID = "SELECT login, name FROM users WHERE id = $1"
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
