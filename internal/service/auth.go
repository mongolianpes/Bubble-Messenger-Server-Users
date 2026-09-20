package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"users/internal/crypto"
	"users/internal/crypto/ecdh"

	pb "users/proto"
)

var registringUsers = map[string]activeUserInfo{}
var authUsers = map[string]activeUserInfo{}

type activeUserInfo struct {
	Key  string
	Time time.Time
}

var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func CheckStartRegAuthUsersTime() {
	durationDelete := time.Second * 12
	sleepTime := time.Second * 20
	for {
		for user := range registringUsers {
			durationStartToNow := time.Since(registringUsers[user].Time)
			if durationStartToNow > durationDelete {
				delete(registringUsers, user)
			}
		}
		for user := range authUsers {
			durationStartToNow := time.Since(authUsers[user].Time)
			if durationStartToNow > durationDelete {
				delete(authUsers, user)
			}
		}
		time.Sleep(sleepTime)
	}
}

func (s *UsersService) TLS(ctx context.Context, req *pb.TLSRequest) (*pb.TLSResponse, error) {
	sharedSecret, serverPublicKey, err := ecdh.GenerateKeys(req.ClientPublicKey)
	if err != nil {
		return nil, err
	}

	if req.IsRegistring {
		if _, ok := registringUsers[req.Id]; ok {
			return nil, errors.New("User with this login registrating now, retry 10 minutes")
		}
	} else {
		if _, ok := authUsers[req.Id]; ok {
			return nil, errors.New("User with this login auth now, retry 10 minutes")
		}
	}

	if req.IsRegistring {
		if len(registringUsers) > 15 {
			return nil, errors.New("Try registering in 3 minutes")
		}
		registringUsers[req.Id] = activeUserInfo{
			Key:  sharedSecret,
			Time: time.Now(),
		}
	} else {
		if len(authUsers) > 15 {
			return nil, errors.New("Please try logging in in 3 minutes")
		}
		authUsers[req.Id] = activeUserInfo{
			Key:  sharedSecret,
			Time: time.Now(),
		}
	}

	resp := &pb.TLSResponse{
		ServerPublicKey: serverPublicKey,
	}
	return resp, nil
}

func (s *UsersService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	_, ok := registringUsers[req.Device]
	if !ok {
		return nil, errors.New("This request is not expected for you")
	}

	key := registringUsers[req.Device].Key
	delete(registringUsers, req.Device)

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

	if err := s.sessionStorage.SetSession(ctx, req.Device, key); err != nil {
		return nil, err
	}

	hashedPassword, err := crypto.HashString(req.Password, true)
	if err != nil {
		return resp, err
	}

	if err := s.usersStorage.RegisterUser(ctx, req.Login, req.Name, hashedPassword); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *UsersService) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	_, ok := authUsers[req.Device]
	if !ok {
		return nil, errors.New("This request is not expected for you")
	}

	var err error
	key := authUsers[req.Device].Key
	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		return nil, err
	}
	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
		return nil, err
	}

	delete(authUsers, req.Device)

	resp := &pb.AuthResponse{
		Key: key,
	}

	currentPassword, err := s.usersStorage.GetPassword(ctx, req.Login)
	if err != nil {
		return resp, err
	}

	if !crypto.VerifyPassword(currentPassword, req.Password) {
		return resp, errors.New("Incorrect login or password")
	}

	if err := s.sessionStorage.SetSession(ctx, req.Device, key); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *UsersService) GetKey(ctx context.Context, req *pb.GetKeyRequest) (*pb.GetKeyResponse, error) {
	key, err := s.sessionStorage.GetKey(ctx, req.Device)
	if err != nil {
		return nil, err
	}

	return &pb.GetKeyResponse{
		Key: key,
	}, nil
}
