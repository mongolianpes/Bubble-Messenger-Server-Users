package service

import (
	"users/internal/db"
	"users/internal/rdb"
	pb "users/proto"
)

type UsersService struct {
	pb.UnimplementedUsersServer
	cacheStorage rdb.CacheStorage
	SQLStorage   db.SQLStorage
}

func NewUsersService(sessionStorage rdb.CacheStorage, usersStorage db.SQLStorage) *UsersService {
	return &UsersService{
		cacheStorage: sessionStorage,
		SQLStorage:   usersStorage,
	}
}
