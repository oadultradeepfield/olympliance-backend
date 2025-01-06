package models

import "time"

type Interaction struct {
	InteractionID   uint      `gorm:"primaryKey;autoIncrement" json:"interaction_id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	ThreadID        uint      `gorm:"" json:"thread_id"`
	CommentID       uint      `gorm:"" json:"comment_id"`
	InteractionType string    `gorm:"not null" json:"interaction_type"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}
