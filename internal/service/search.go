package service

import (
	"context"
	"strings"
	pb "users/proto"
)

const prefixForSearchExactMatch = "@"

func (s *UsersService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if strings.HasPrefix(req.Query, prefixForSearchExactMatch) {
		resultUserInfo := []*pb.UserInfo{}

		login := req.Query[1:]
		name, err := s.cacheStorage.GetUser(ctx, login)
		if err == nil {
			resultUserInfo = append(resultUserInfo, &pb.UserInfo{
				Login: login,
				Name:  name,
			})

			return &pb.SearchResponse{
				Users: resultUserInfo,
			}, nil
		}

		name, id, err := s.SQLStorage.GetUserInfo(ctx, login)
		if err != nil {
			return nil, err
		}

		resultUserInfo = append(resultUserInfo, &pb.UserInfo{
			Login: login,
			Name:  name,
			Id:    int64(id),
		})

		s.cacheStorage.SaveUser(ctx, login, name)

		return &pb.SearchResponse{
			Users: resultUserInfo,
		}, nil
	}

	usersInfo, err := s.cacheStorage.GetUsers(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	return &pb.SearchResponse{
		Users: usersInfo,
	}, nil
}

func (s *UsersService) GetInfoByID(ctx context.Context, req *pb.GetInfoByIDRequest) (*pb.GetInfoByIDResponse, error) {
	login, name, err := s.SQLStorage.GetInfoByID(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &pb.GetInfoByIDResponse{
		Login: login,
		Name:  name,
	}, nil
}

func (s *UsersService) AddUserAvatarPath(ctx context.Context, req *pb.AddUserAvatarPathRequest) (*pb.AddUserAvatarPathResponse, error) {
	if err := s.SQLStorage.AddUserAvatar(ctx, req.Login, req.StorageAvatarPath); err != nil {
		return nil, err
	}

	return &pb.AddUserAvatarPathResponse{}, nil
}

func (s *UsersService) GetUserAvatarPath(ctx context.Context, req *pb.GetUserAvatarPathRequest) (*pb.GetUserAvatarPathResponse, error) {
	avatarPath, err := s.SQLStorage.GetUserAvatarStoragePath(ctx, req.Login)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserAvatarPathResponse{
		StoragePath: avatarPath,
	}, nil
}
