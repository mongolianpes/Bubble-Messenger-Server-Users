package service

import (
	"context"
	"strings"
	pb "users/proto"
)

const prefixForSearchExactMatch = "@"

func (s *UsersService) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if strings.HasPrefix(req.Query, prefixForSearchExactMatch) {
		name, err := s.SQLStorage.GetUserInfo(ctx, req.Query)
		if err != nil {
			return nil, err
		}

		resultUserInfo := []*pb.UserInfo{}
		resultUserInfo = append(resultUserInfo, &pb.UserInfo{
			Login: req.Query,
			Name:  name,
		})

		s.cacheStorage.SaveUser(ctx, req.Query, name)

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
