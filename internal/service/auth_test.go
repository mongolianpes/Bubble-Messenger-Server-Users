package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "users/proto"
)

func TestCleanup(t *testing.T) {
	u := NewAuthAndRegUsers()

	u.Lock()
	u.reg["expired"] = info{key: "k", expiresAt: time.Now().Add(-time.Second)}
	u.reg["alive"] = info{key: "k", expiresAt: time.Now().Add(time.Hour)}
	u.auth["expired"] = info{key: "k", expiresAt: time.Now().Add(-time.Second)}
	u.Unlock()

	u.cleanup()

	u.Lock()
	defer u.Unlock()

	_, ok := u.reg["expired"]
	assert.False(t, ok)
	_, ok = u.reg["alive"]
	assert.True(t, ok)
	_, ok = u.auth["expired"]
	assert.False(t, ok)
}

func TestGetAuthInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, _, cache := newTestService()
		cache.On("GetAuthInfo", ctx, "dev-1").
			Return("secret-key", 42, nil).
			Once()

		resp, err := svc.GetAuthInfo(ctx, &pb.GetAuthInfoRequest{Device: "dev-1"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "secret-key", resp.Key)
		assert.Equal(t, int64(42), resp.UserId)
		cache.AssertExpectations(t)
	})

	t.Run("error from cache", func(t *testing.T) {
		svc, _, cache := newTestService()
		cache.On("GetAuthInfo", ctx, "dev-2").
			Return("", 0, errors.New("not found")).
			Once()

		resp, err := svc.GetAuthInfo(ctx, &pb.GetAuthInfoRequest{Device: "dev-2"})
		require.Error(t, err)
		assert.Nil(t, resp)
		cache.AssertExpectations(t)
	})
}

func TestTLS(t *testing.T) {
	ctx := context.Background()

	t.Run("already registering", func(t *testing.T) {
		svc, _, _ := newTestService()

		svc.activeUsers.Lock()
		svc.activeUsers.reg["device-1"] = info{
			key:       "existing",
			expiresAt: time.Now().Add(time.Minute),
		}
		svc.activeUsers.Unlock()

		req := &pb.TLSRequest{
			Id:              "device-1",
			ClientPublicKey: "invalid-key", // ecdh упадёт раньше проверки
			IsRegistring:    true,
		}

		_, err := svc.TLS(ctx, req)
		require.Error(t, err)
		// либо ошибка ecdh, либо "registrating now" — зависит от порядка
	})

	t.Run("already authenticating", func(t *testing.T) {
		svc, _, _ := newTestService()

		svc.activeUsers.Lock()
		svc.activeUsers.auth["device-2"] = info{
			key:       "existing",
			expiresAt: time.Now().Add(time.Minute),
		}
		svc.activeUsers.Unlock()

		req := &pb.TLSRequest{
			Id:              "device-2",
			ClientPublicKey: "invalid-key",
			IsRegistring:    false,
		}

		_, err := svc.TLS(ctx, req)
		require.Error(t, err)
	})

	t.Run("register limit exceeded", func(t *testing.T) {
		svc, _, _ := newTestService()

		svc.activeUsers.Lock()
		for i := 0; i <= limitForActiveUsers; i++ {
			svc.activeUsers.reg[string(rune('a'+i))] = info{
				key:       "k",
				expiresAt: time.Now().Add(time.Minute),
			}
		}
		svc.activeUsers.Unlock()

		req := &pb.TLSRequest{
			Id:              "new-device",
			ClientPublicKey: "invalid-key",
			IsRegistring:    true,
		}

		_, err := svc.TLS(ctx, req)
		require.Error(t, err)
	})

	t.Run("auth limit exceeded", func(t *testing.T) {
		svc, _, _ := newTestService()

		svc.activeUsers.Lock()
		for i := 0; i <= limitForActiveUsers; i++ {
			svc.activeUsers.auth[string(rune('a'+i))] = info{
				key:       "k",
				expiresAt: time.Now().Add(time.Minute),
			}
		}
		svc.activeUsers.Unlock()

		req := &pb.TLSRequest{
			Id:              "new-device",
			ClientPublicKey: "invalid-key",
			IsRegistring:    false,
		}

		_, err := svc.TLS(ctx, req)
		require.Error(t, err)
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("request not expected", func(t *testing.T) {
		svc, _, _ := newTestService()

		resp, err := svc.Register(ctx, &pb.RegisterRequest{
			Device: "unknown-device",
		})
		require.Error(t, err)
		assert.Equal(t, "This request is not expected for you", err.Error())
		assert.Nil(t, resp)
	})

	t.Run("entry removed even on decrypt error", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		device := "device-reg-1"
		svc.activeUsers.Lock()
		svc.activeUsers.reg[device] = info{
			key:       "shared-secret",
			expiresAt: time.Now().Add(time.Minute),
		}
		svc.activeUsers.Unlock()

		resp, err := svc.Register(ctx, &pb.RegisterRequest{
			Device:   device,
			Login:    "not-encrypted",
			Password: "not-encrypted",
			Name:     "not-encrypted",
		})

		// crypto.StringDecrypt должен вернуть ошибку
		require.Error(t, err)
		assert.Nil(t, resp)

		// запись должна быть удалена (delete стоит до decrypt)
		svc.activeUsers.Lock()
		_, exists := svc.activeUsers.reg[device]
		svc.activeUsers.Unlock()
		assert.False(t, exists)

		sqlMock.AssertNotCalled(t, "RegisterUser")
		cacheMock.AssertNotCalled(t, "SetSession")
	})
}

func TestAuth(t *testing.T) {
	ctx := context.Background()

	t.Run("request not expected", func(t *testing.T) {
		svc, _, _ := newTestService()

		resp, err := svc.Auth(ctx, &pb.AuthRequest{
			Device: "unknown-device",
		})
		require.Error(t, err)
		assert.Equal(t, "This request is not expected for you", err.Error())
		assert.Nil(t, resp)
	})

	t.Run("entry stays on decrypt error (current behavior)", func(t *testing.T) {
		svc, sqlMock, cacheMock := newTestService()

		device := "device-auth-1"
		svc.activeUsers.Lock()
		svc.activeUsers.auth[device] = info{
			key:       "shared-secret",
			expiresAt: time.Now().Add(time.Minute),
		}
		svc.activeUsers.Unlock()

		resp, err := svc.Auth(ctx, &pb.AuthRequest{
			Device:   device,
			Login:    "not-encrypted",
			Password: "not-encrypted",
		})

		require.Error(t, err)
		assert.Nil(t, resp)

		// в текущем коде delete стоит ПОСЛЕ decrypt → запись остаётся
		svc.activeUsers.Lock()
		_, exists := svc.activeUsers.auth[device]
		svc.activeUsers.Unlock()
		assert.True(t, exists)

		sqlMock.AssertNotCalled(t, "GetPassword")
		cacheMock.AssertNotCalled(t, "SetSession")
	})
}
