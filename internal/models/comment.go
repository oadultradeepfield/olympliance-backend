package models

import (
	"encoding/json"
	"time"
)

type Comment struct {
	CommentID       uint            `gorm:"primaryKey;autoIncrement" json:"comment_id"`
	ThreadID        uint            `gorm:"not null" json:"thread_id"`
	UserID          uint            `gorm:"not null" json:"user_id"`
	ParentCommentID uint            `gorm:"" json:"parent_comment_id"`
	Content         string          `gorm:"not null" json:"content"`
	Stats           json.RawMessage `gorm:"type:jsonb;default:'{\"upvotes\": 0, \"downvotes\": 0}'::jsonb" json:"stats"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted       bool            `gorm:"default:false" json:"is_deleted"`
}
