package notification

import "time"

// Notification represents a notification persisted in Postgres.
type Notification struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"index;not null"`
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// ScoreEvent is the incoming event payload from Kafka.
type ScoreEvent struct {
	UserID   int64 `json:"user_id"`
	NewScore int64 `json:"new_score"`
	Change   int32 `json:"change"`
}
