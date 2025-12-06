package grpcserver

import (
	"context"
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	eventpb "github.com/emorenkov/scorehub/pkg/event/proto"
	"github.com/emorenkov/scorehub/pkg/event/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	eventpb.UnimplementedEventServiceServer
	svc service.Service
}

func NewServer(svc service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) SendScoreEvent(ctx context.Context, req *eventpb.ScoreEventRequest) (*eventpb.EventAck, error) {
	ev := &event.ScoreEvent{
		UserID:   req.GetUserId(),
		NewScore: req.GetNewScore(),
		Change:   req.GetChange(),
	}
	ack, err := s.svc.Send(ctx, ev)
	if err != nil {
		return nil, mapError(err)
	}
	return &eventpb.EventAck{Status: ack.Status}, nil
}

func mapError(err error) error {
	if se, ok := apperrors.AsStatusError(err); ok {
		switch se.Status {
		case http.StatusBadRequest:
			return status.Error(codes.InvalidArgument, se.Message)
		default:
			return status.Error(codes.Internal, se.Error())
		}
	}
	return status.Error(codes.Internal, err.Error())
}
