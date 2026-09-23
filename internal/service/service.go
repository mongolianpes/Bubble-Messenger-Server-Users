package service

import (
	"users/internal/db"
	"users/internal/rdb"
	pb "users/proto"
)

type UsersService struct {
	pb.UnimplementedUsersServiceServer
	cacheStorage rdb.CacheStorage
	SQLStorage   db.SQLStorage
	activeUsers  *ActiveUsers
}

func NewUsersService(sessionStorage rdb.CacheStorage, usersStorage db.SQLStorage, activeUsers *ActiveUsers) *UsersService {
	return &UsersService{
		cacheStorage: sessionStorage,
		SQLStorage:   usersStorage,
		activeUsers:  activeUsers,
	}
}
