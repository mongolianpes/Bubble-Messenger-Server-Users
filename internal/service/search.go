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
