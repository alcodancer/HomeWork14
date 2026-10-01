package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "example/api/user/v1"
	"example/internal/domain"
	"example/internal/service"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	svc *service.UserService
}

func NewUserServer(svc *service.UserService) *UserServer {
	return &UserServer{svc: svc}
}

func (s *UserServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	u, err := s.svc.CreateUser(ctx, req.GetName(), req.GetEmail())
	if err != nil {
		return nil, toGRPCErr(err)
	}
	return &userv1.CreateUserResponse{User: toProto(u)}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	u, err := s.svc.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCErr(err)
	}
	return &userv1.GetUserResponse{User: toProto(u)}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	u, err := s.svc.UpdateUser(ctx, req.GetId(), req.GetName(), req.GetEmail())
	if err != nil {
		return nil, toGRPCErr(err)
	}
	return &userv1.UpdateUserResponse{User: toProto(u)}, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if err := s.svc.DeleteUser(ctx, req.GetId()); err != nil {
		return nil, toGRPCErr(err)
	}
	return &userv1.DeleteUserResponse{}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, _ *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, err := s.svc.ListUsers(ctx)
	if err != nil {
		return nil, toGRPCErr(err)
	}
	out := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProto(u))
	}
	return &userv1.ListUsersResponse{Users: out}, nil
}

func to