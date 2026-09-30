package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"users/internal/crypto"
	"users/internal/crypto/ecdh"

	pb "users/proto"
)

const (
	waitingTimeRegAndAuth = time.Second * 10
	limitForActiveUsers   = 15
)

type info struct {
	key       string
	expiresAt time.Time
}

type ActiveUsers struct {
	reg  map[string]info
	auth map[string]info
	sync.Mutex
}

func NewAuthAndRegUsers() *ActiveUsers {
	users := ActiveUsers{
		reg:  map[string]info{},
		auth: map[string]info{},
	}

	go func() {
		ticker := time.NewTicker(waitingTimeRegAndAuth)
		defer ticker.Stop()
		for range ticker.C {
			users.cleanup()
		}
	}()

	return &users
}

func (u *ActiveUsers) cleanup() {
	u.Lock()
	defer u.Unlock()
	now := time.Now()
	for k, v := range u.reg {
		if now.After(v.expiresAt) {
			delete(u.reg, k)
		}
	}

	for k, v := range u.auth {
		if now.After(v.expiresAt) {
			delete(u.auth, k)
		}
	}
}

func (s *UsersService) TLS(ctx context.Context, req *pb.TLSRequest) (*pb.TLSResponse, error) {
	s.activeUsers.Lock()
	defer s.activeUsers.Unlock()

	if req.IsRegistring {
		if _, ok := s.activeUsers.reg[req.Id]; ok {
			slog.WarnContext(ctx, "TLS: device has already completed the TLS handshake", "deviceID", req.Id)
			return nil, errors.New("User has already completed the TLS handshake, retry 10 seconds")
		}
	} else {
		if _, ok := s.activeUsers.auth[req.Id]; ok {
			slog.WarnContext(ctx, "TLS: device has already completed the TLS handshake", "deviceID", req.Id)
			return nil, errors.New("User has already completed the TLS handshake, retry 10 seconds")
		}
	}

	sharedSecret, serverPublicKey, err := ecdh.GenerateKeys(req.ClientPublicKey)
	if err != nil {
		slog.WarnContext(ctx, "TLS: Generate key", "clientPublicKey", req.ClientPublicKey, "error", err)
		return nil, err
	}

	if req.IsRegistring {
		if len(s.activeUsers.reg) > limitForActiveUsers {
			return nil, errors.New("Try registering in 3 minutes")
		}
		s.activeUsers.reg[req.Id] = info{
			key:       sharedSecret,
			expiresAt: time.Now(),
		}
	} else {
		if len(s.activeUsers.auth) > limitForActiveUsers {
			return nil, errors.New("Please try logging in 3 minutes")
		}
		s.activeUsers.auth[req.Id] = info{
			key:       sharedSecret,
			expiresAt: time.Now(),
		}
	}

	resp := &pb.TLSResponse{
		ServerPublicKey: serverPublicKey,
	}
	return resp, nil
}

func (s *UsersService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.activeUsers.Lock()
	defer s.activeUsers.Unlock()
	info, ok := s.activeUsers.reg[req.Device]
	if !ok {
		slog.WarnContext(ctx, "Register: device did not pass TLS", "device", req.Device)
		return nil, errors.New("This request is not expected for you")
	}

	delete(s.activeUsers.reg, req.Device)

	key := info.key

	var err error
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		slog.WarnContext(ctx, "Decrypt error", "device", req.Device, "error", err)
		return nil, err
	}

	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
		slog.WarnContext(ctx, "Decrypt error", "device", req.Device, "error", err)
		return nil, err
	}

	resp := &pb.RegisterResponse{
		Key: key,
	}
	if len(req.Password) <= 9 {
		return resp, err
	}
	req.Name, err = crypto.StringDecrypt(req.Name, key)
	if err != nil {
		slog.WarnContext(ctx, "Decrypt error", "device", req.Device, "error", err)
		return nil, err
	}

	// if err := s.cacheStorage.SetSession(ctx, req.Device, key); err != nil {
	// 	return nil, err
	// }

	hashedPassword, err := crypto.HashString(req.Password, true)
	if err != nil {
		slog.WarnContext(ctx, "Hash error", "device", req.Device, "error", err)
		return resp, err
	}

	if err := s.SQLStorage.RegisterUser(ctx, req.Login, req.Name, hashedPassword); err != nil {
		return resp, err
	}

	slog.InfoContext(ctx, "Success register user", "login", req.Login, "name", req.Name, "deviceID", req.Device)

	return resp, nil
}

func (s *UsersService) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	s.activeUsers.Lock()
	defer s.activeUsers.Unlock()
	info, ok := s.activeUsers.auth[req.Device]
	if !ok {
		slog.WarnContext(ctx, "Auth: device did not pass TLS", "device", req.Device)
		return nil, errors.New("This request is not expected for you")
	}

	var err error
	key := info.key
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		slog.WarnContext(ctx, "Decrypt error", "device", req.Device, "error", err)
		return nil, err
	}
	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
		slog.WarnContext(ctx, "Decrypt error", "device", req.Device, "error", err)
		return nil, err
	}

	delete(s.activeUsers.auth, req.Device)

	resp := &pb.AuthResponse{
		Key: key,
	}

	currentPassword, err := s.SQLStorage.GetPassword(ctx, req.Login)
	if err != nil {
		slog.ErrorContext(ctx, "Get user password from DB", "login", req.Login, "error", err)
		return resp, err
	}

	if !crypto.VerifyPassword(currentPassword, req.Password) {
		slog.WarnContext(ctx, "User input incorrect password", "login", req.Login, "deviceID", req.Device)
		return resp, errors.New("Incorrect login or password")
	}

	userName, id, err := s.SQLStorage.GetUserInfo(ctx, req.Login)
	if err != nil {
		slog.ErrorContext(ctx, "Get user info from DB", "login", req.Login, "error", err)
		return resp, err
	}

	resp.UserName = userName
	resp.UserId = int64(id)

	if err := s.cacheStorage.SetSession(ctx, req.Device, key, id); err != nil {
		slog.ErrorContext(ctx, "Set session", "deviceID", req.Device, "userID", id)
		return resp, err
	}

	slog.InfoContext(ctx, "Success auth user", "deviceID", req.Device, "error", err)

	return resp, nil
}

func (s *UsersService) GetAuthInfo(ctx context.Context, req *pb.GetAuthInfoRequest) (*pb.GetAuthInfoResponse, error) {
	key, id, err := s.cacheStorage.GetAuthInfo(ctx, req.Device)
	if err != nil {
		slog.ErrorContext(ctx, "Get session", "error", err)
		return nil, err
	}

	resp := &pb.GetAuthInfoResponse{
		Key: key,
	}

	currentPassword, err := s.SQLStorage.GetPassword(ctx, req.Login)
	if err != nil {
		slog.ErrorContext(ctx, "Get use password from DB", "login", req.Login)
		return resp, err
	}

	if !crypto.VerifyPassword(currentPassword, req.Password) {
		slog.WarnContext(ctx, "User input incorrect password", "login", req.Login)
		return resp, errors.New("Incorrect login or password")
	}

	resp.UserId = int64(id)

	return resp, nil
}
