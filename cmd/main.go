package main

import (
	"log/slog"
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
		slog.Error("Connect to DB", "error", err)
		return
	}

	sessionStorage, err := rdb.NewRedisStorage()
	if err != nil {
		slog.Error("Connect to DB")
		return
	}

	activeUsers := service.NewAuthAndRegUsers()

	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		slog.Error("Start listen port", "error", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUsersServiceServer(grpcServer, service.NewUsersService(sessionStorage, usersStorage, activeUsers))

	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("Serve gRPC service", "error", err)
	}
}
