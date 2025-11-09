package event

import (
	"context"
	"strconv"

	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	eventcfg "github.com/emorenkov/scorehub/pkg/event/config"
)

type Handler struct {
	producer *ckafka.Producer
	topic    string
}

type ScoreEventRequest struct {
	UserID   int64 `json:"user_id"`
	NewScore int64 `json:"new_score"`
	Change   int32 `json:"change"`
}

type EventAck struct {
	Status string `json:"status"`
}

func NewHandler(cfg *eventcfg.Config) *Handler {
	return &Handler{
		producer: ckafka.NewProducerWithBrokers(cfg.KafkaBrokers, cfg.ScoreEventsTopic),
		topic:    cfg.ScoreEventsTopic,
	}
}

func (h *Handler) Close() error { return h.producer.Close() }

// SendScoreEvent publishes the given event to Kafka.
func (h *Handler) SendScoreEvent(ctx context.Context, req *ScoreEventRequest) (*EventAck, error) {
	if err := h.producer.SendMessage(ctx, keyForUser(req.UserID), req); err != nil {
		return nil, err
	}
	return &EventAck{Status: "ok"}, nil
}

func keyForUser(userID int64) string { return "user:" + strconv.FormatInt(userID, 10) }
