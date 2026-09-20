package service

import (
	"users/internal/db"
	"users/internal/rdb"
	pb "users/proto"
)

type UsersService struct {
	pb.UnimplementedUsersServer
	sessionStorage rdb.SessionStorage
	usersStorage   db.UsersStorage
}

func NewUsersService(sessionStorage rdb.SessionStorage, usersStorage db.UsersStorage) *UsersService {
	return &UsersService{
		sessionStorage: sessionStorage,
		usersStorage:   usersStorage,
	}
}
