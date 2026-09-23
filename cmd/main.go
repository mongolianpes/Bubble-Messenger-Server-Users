package main

import (
	"net"
	"users/internal/db"
	"users/internal/rdb"
	"users/internal/service"

	pb "users/proto"

	"google.golang.org/grpc"
)

func main() {
	usersStorage, err := db.NewPostgresStorage()
	if err != nil {
		panic(err)
	}

	sessionStorage, err := rdb.NewRedisStorage()
	if err != nil {
		panic(err)
	}

	activeUsers := service.NewAuthAndRegUsers()

	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUsersServiceServer(grpcServer, service.NewUsersService(sessionStorage, usersStorage, activeUsers))

	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
