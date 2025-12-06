package event

// ScoreEvent represents the payload published to Kafka and exposed via APIs.
type ScoreEvent struct {
	UserID   int64 `json:"user_id"`
	NewScore int64 `json:"new_score"`
	Change   int32 `json:"change"`
}

// EventAck mirrors the gRPC/REST acknowledgement response.
type EventAck struct {
	Status string `json:"status"`
}
