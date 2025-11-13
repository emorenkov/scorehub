package grpcserver

import (
	"context"
	"errors"
	"time"

	"github.com/emorenkov/scorehub/pkg/common/models"
	userpb "github.com/emorenkov/scorehub/pkg/user/models/proto"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Server implements the generated gRPC interface and forwards calls to the domain service.
type Server struct {
	userpb.UnimplementedUserServiceServer
	svc service.Service
}

func NewServer(svc service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Create(ctx, req.GetName(), req.GetEmail())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create user: %v", err)
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, mapError(err)
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Update(ctx, req.GetId(), req.GetName(), req.GetEmail())
	if err != nil {
		return nil, mapError(err)
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.Empty, error) {
	if err := s.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, mapError(err)
	}
	return &userpb.Empty{}, nil
}

func (s *Server) ListUsers(ctx context.Context, _ *userpb.Empty) (*userpb.ListUsersResponse, error) {
	users, err := s.svc.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %v", err)
	}
	resp := &userpb.ListUsersResponse{
		Users: make([]*userpb.User, 0, len(users)),
	}
	for i := range users {
		resp.Users = append(resp.Users, toProtoUser(&users[i]))
	}
	return resp, nil
}

func toProtoUser(u *models.User) *userpb.User {
	return &userpb.User{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Score:     u.Score,
		CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}
