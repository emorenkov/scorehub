package grpcserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/common/models"
	userpb "github.com/emorenkov/scorehub/pkg/user/models/proto"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the generated gRPC interface and forwards calls to the domain service.
type Server struct {
	userpb.UnimplementedUserServiceServer
	svc service.Service
	log *zap.Logger
}

func NewServer(svc service.Service, log *zap.Logger) *Server {
	return &Server{svc: svc, log: log}
}

func (s *Server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Create(ctx, req.GetName(), req.GetEmail())
	if err != nil {
		if s.log != nil {
			s.log.Error("grpc CreateUser failed", zap.Error(err), zap.String("email", req.GetEmail()))
		}
		return nil, status.Errorf(codes.InvalidArgument, "create user: %v", err)
	}
	if s.log != nil {
		s.log.Info("grpc CreateUser succeeded", zap.Int64("user_id", user.ID))
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Get(ctx, req.GetId())
	if err != nil {
		if s.log != nil {
			s.log.Error("grpc GetUser failed", zap.Error(err), zap.Int64("user_id", req.GetId()))
		}
		return nil, mapError(err)
	}
	if s.log != nil {
		s.log.Info("grpc GetUser succeeded", zap.Int64("user_id", user.ID))
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	user, err := s.svc.Update(ctx, req.GetId(), req.GetName(), req.GetEmail())
	if err != nil {
		if s.log != nil {
			s.log.Error("grpc UpdateUser failed", zap.Error(err), zap.Int64("user_id", req.GetId()))
		}
		return nil, mapError(err)
	}
	if s.log != nil {
		s.log.Info("grpc UpdateUser succeeded", zap.Int64("user_id", user.ID))
	}
	return &userpb.UserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.Empty, error) {
	if err := s.svc.Delete(ctx, req.GetId()); err != nil {
		if s.log != nil {
			s.log.Error("grpc DeleteUser failed", zap.Error(err), zap.Int64("user_id", req.GetId()))
		}
		return nil, mapError(err)
	}
	if s.log != nil {
		s.log.Info("grpc DeleteUser succeeded", zap.Int64("user_id", req.GetId()))
	}
	return &userpb.Empty{}, nil
}

func (s *Server) ListUsers(ctx context.Context, _ *userpb.Empty) (*userpb.ListUsersResponse, error) {
	users, err := s.svc.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error("grpc ListUsers failed", zap.Error(err))
		}
		return nil, status.Errorf(codes.Internal, "list users: %v", err)
	}
	resp := &userpb.ListUsersResponse{
		Users: make([]*userpb.User, 0, len(users)),
	}
	for i := range users {
		resp.Users = append(resp.Users, toProtoUser(&users[i]))
	}
	if s.log != nil {
		s.log.Info("grpc ListUsers succeeded", zap.Int("count", len(resp.Users)))
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
	var se *apperrors.StatusError
	if errors.As(err, &se) {
		switch se.Status {
		case http.StatusBadRequest:
			return status.Error(codes.InvalidArgument, se.Message)
		case http.StatusNotFound:
			return status.Error(codes.NotFound, se.Message)
		default:
			return status.Error(codes.Internal, se.Error())
		}
	}
	return status.Error(codes.Internal, err.Error())
}
