package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "users/proto"
)

func TestSearch(t *testing.T) {
	ctx := context.Background()

	t.Run("exact match with @ prefix - success", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "@ivan"
		expectedName := "Иван Иванов"
		expectedID := 10

		sqlMock.On("GetUserInfo", ctx, query).
			Return(expectedName, expectedID, nil).
			Once()

		cacheMock.On("SaveUser", ctx, query, expectedName).
			Return(nil).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Users, 1)

		assert.Equal(t, query, resp.Users[0].Login)
		assert.Equal(t, expectedName, resp.Users[0].Name)
		assert.Equal(t, int64(expectedID), resp.Users[0].Id)

		sqlMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("exact match with @ prefix - SQL error", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "@unknown"
		sqlErr := errors.New("user not found")

		sqlMock.On("GetUserInfo", ctx, query).
			Return("", 0, sqlErr).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.Error(t, err)
		assert.Equal(t, sqlErr, err)
		assert.Nil(t, resp)

		sqlMock.AssertExpectations(t)
		cacheMock.AssertNotCalled(t, "SaveUser")
	})

	t.Run("exact match with @ prefix - SaveUser error is ignored", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "@petr"
		expectedName := "Пётр"
		expectedID := 5

		sqlMock.On("GetUserInfo", ctx, query).
			Return(expectedName, expectedID, nil).
			Once()

		cacheMock.On("SaveUser", ctx, query, expectedName).
			Return(errors.New("redis is down")).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.Len(t, resp.Users, 1)
		assert.Equal(t, expectedName, resp.Users[0].Name)
		assert.Equal(t, int64(expectedID), resp.Users[0].Id)

		sqlMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("partial search - success from cache", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "iva"
		expectedUsers := []*pb.UserInfo{
			{Login: "user_ivan", Name: "Иван", Id: 1},
			{Login: "user_ivanov", Name: "Иванов", Id: 2},
		}

		cacheMock.On("GetUsers", ctx, query).
			Return(expectedUsers, nil).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, expectedUsers, resp.Users)

		cacheMock.AssertExpectations(t)
		sqlMock.AssertNotCalled(t, "GetUserInfo")
	})

	t.Run("partial search - cache error", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "pet"
		cacheErr := errors.New("redis scan failed")

		cacheMock.On("GetUsers", ctx, query).
			Return(nil, cacheErr).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.Error(t, err)
		assert.Equal(t, cacheErr, err)
		assert.Nil(t, resp)

		cacheMock.AssertExpectations(t)
		sqlMock.AssertNotCalled(t, "GetUserInfo")
	})

	t.Run("partial search - empty result", func(t *testing.T) {
		svc, _, cacheMock := newTestService()

		query := "xyz"

		cacheMock.On("GetUsers", ctx, query).
			Return([]*pb.UserInfo{}, nil).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Users)

		cacheMock.AssertExpectations(t)
	})

	t.Run("query is just @", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "@"
		expectedName := "Root"
		expectedID := 1

		sqlMock.On("GetUserInfo", ctx, query).
			Return(expectedName, expectedID, nil).
			Once()
		cacheMock.On("SaveUser", ctx, query, expectedName).
			Return(nil).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.Len(t, resp.Users, 1)
		assert.Equal(t, "@", resp.Users[0].Login)
		assert.Equal(t, expectedName, resp.Users[0].Name)
		assert.Equal(t, int64(expectedID), resp.Users[0].Id)

		sqlMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("query without @ goes to cache even if looks like login", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		query := "ivan"

		cacheMock.On("GetUsers", ctx, query).
			Return([]*pb.UserInfo{{Login: "user_ivan", Name: "Иван", Id: 3}}, nil).
			Once()

		resp, err := svc.Search(ctx, &pb.SearchRequest{Query: query})
		require.NoError(t, err)
		require.Len(t, resp.Users, 1)

		sqlMock.AssertNotCalled(t, "GetUserInfo")
		cacheMock.AssertExpectations(t)
	})
}
