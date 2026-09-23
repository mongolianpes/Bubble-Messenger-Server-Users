package service

import (
	"context"
	"errors"
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
			return nil, errors.New("User with this login registrating now, retry 10 seconds")
		}
	} else {
		if _, ok := s.activeUsers.auth[req.Id]; ok {
			return nil, errors.New("User with this login auth now, retry 10 seconds")
		}
	}

	sharedSecret, serverPublicKey, err := ecdh.GenerateKeys(req.ClientPublicKey)
	if err != nil {
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
		return nil, errors.New("This request is not expected for you")
	}

	delete(s.activeUsers.reg, req.Device)

	key := info.key

	var err error
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		return nil, err
	}

	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
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
		return nil, err
	}

	// if err := s.cacheStorage.SetSession(ctx, req.Device, key); err != nil {
	// 	return nil, err
	// }

	hashedPassword, err := crypto.HashString(req.Password, true)
	if err != nil {
		return resp, err
	}

	if err := s.SQLStorage.RegisterUser(ctx, req.Login, req.Name, hashedPassword); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *UsersService) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	s.activeUsers.Lock()
	defer s.activeUsers.Unlock()
	info, ok := s.activeUsers.auth[req.Device]
	if !ok {
		return nil, errors.New("This request is not expected for you")
	}

	var err error
	key := info.key
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		return nil, err
	}
	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
		return nil, err
	}

	delete(s.activeUsers.auth, req.Device)

	resp := &pb.AuthResponse{
		Key: key,
	}

	currentPassword, err := s.SQLStorage.GetPassword(ctx, req.Login)
	if err != nil {
		return resp, err
	}

	if !crypto.VerifyPassword(currentPassword, req.Password) {
		return resp, errors.New("Incorrect login or password")
	}

	_, id, err := s.SQLStorage.GetUserInfo(ctx, req.Login)
	if err != nil {
		return resp, err
	}

	if err := s.cacheStorage.SetSession(ctx, req.Device, key, id); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *UsersService) GetAuthInfo(ctx context.Context, req *pb.GetAuthInfoRequest) (*pb.GetAuthInfoResponse, error) {
	key, id, err := s.cacheStorage.GetAuthInfo(ctx, req.Device)
	if err != nil {
		return nil, err
	}

	return &pb.GetAuthInfoResponse{
		Key:    key,
		UserId: int64(id),
	}, nil
}
