package service

import (
	"context"
	pb "users/proto"

	"github.com/stretchr/testify/mock"
)

type MockSQLStorage struct {
	mock.Mock
}

func (m *MockSQLStorage) GetUserInfo(ctx context.Context, login string) (string, error) {
	args := m.Called(ctx, login)
	return args.String(0), args.Error(1)
}

func (m *MockSQLStorage) GetPassword(ctx context.Context, login string) (string, error) {
	args := m.Called(ctx, login)
	return args.String(0), args.Error(1)
}

func (m *MockSQLStorage) RegisterUser(ctx context.Context, login, name, password string) error {
	args := m.Called(ctx, login, name, password)
	return args.Error(0)
}

type MockCacheStorage struct {
	mock.Mock
}

func (m *MockCacheStorage) SetSession(ctx context.Context, sessionKey, userID string) error {
	args := m.Called(ctx, sessionKey, userID)
	return args.Error(0)
}

func (m *MockCacheStorage) GetKey(ctx context.Context, device string) (string, error) {
	args := m.Called(ctx, device)
	return args.String(0), args.Error(1)
}

func (m *MockCacheStorage) GetUsers(ctx context.Context, login string) ([]*pb.UserInfo, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.UserInfo), args.Error(1)
}

func (m *MockCacheStorage) SaveUser(ctx context.Context, login, name string) error {
	args := m.Called(ctx, login, name)
	return args.Error(0)
}

func (m *MockCacheStorage) Close() error { return nil }

// ---------- Хелпер ----------

func newTestService() (*UsersService, *MockSQLStorage, *MockCacheStorage) {
	sql := new(MockSQLStorage)
	cache := new(MockCacheStorage)
	svc := &UsersService{
		SQLStorage:   sql,
		cacheStorage: cache,
		activeUsers:  NewAuthAndRegUsers(),
	}
	return svc, sql, cache
}
