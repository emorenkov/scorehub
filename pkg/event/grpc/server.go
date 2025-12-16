package grpcserver

import (
	"context"
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	eventpb "github.com/emorenkov/scorehub/pkg/event/proto"
	"github.com/emorenkov/scorehub/pkg/event/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	eventpb.UnimplementedEventServiceServer
	svc service.Event
	log *zap.Logger
}

func NewServer(svc service.Event, log *zap.Logger) *Server {
	return &Server{svc: svc, log: log}
}

func (s *Server) SendScoreEvent(ctx context.Context, req *eventpb.ScoreEventRequest) (*eventpb.EventAck, error) {
	ev := &event.ScoreEvent{
		UserID:   req.GetUserId(),
		NewScore: req.GetNewScore(),
		Change:   req.GetChange(),
	}
	ack, err := s.svc.Send(ctx, ev)
	if err != nil {
		if s.log != nil {
			s.log.Error("grpc SendScoreEvent failed", zap.Error(err), zap.Int64("user_id", ev.UserID))
		}
		return nil, mapError(err)
	}
	if s.log != nil {
		s.log.Info("grpc SendScoreEvent succeeded", zap.String("status", ack.Status), zap.Int64("user_id", ev.UserID))
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
