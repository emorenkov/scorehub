package models

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	Score     int64     `gorm:"not null;default:0"`
	Deleted   bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
