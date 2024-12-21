package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type User struct {
	UserID       uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	RefreshToken string    `gorm:"default:null" json:"refresh_token"`
	RoleID       int       `gorm:"default:0;not null" json:"role_id"`
	Reputation   int       `gorm:"default:0" json:"reputation"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsBanned     bool      `gorm:"default:false" json:"is_banned"`
}

type Thread struct {
	ThreadID   uint            `gorm:"primaryKey;autoIncrement" json:"thread_id"`
	UserID     uint            `gorm:"not null" json:"user_id"`
	Title      string          `gorm:"not null" json:"title"`
	Content    string          `gorm:"not null" json:"content"`
	CategoryID uint            `gorm:"not null" json:"category_id"`
	Stats      json.RawMessage `gorm:"type:jsonb;default:'{\"followers\": 0, \"upvotes\": 0, \"downvotes\": 0, \"comments\": 0}'::jsonb" json:"stats"`
	Tags       pq.StringArray  `gorm:"type:text[]" json:"tags"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted  bool            `gorm:"default:false" json:"is_deleted"`
}

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

type Interaction struct {
	InteractionID   uint      `gorm:"primaryKey;autoIncrement" json:"interaction_id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	ThreadID        uint      `gorm:"" json:"thread_id"`
	CommentID       uint      `gorm:"" json:"comment_id"`
	InteractionType string    `gorm:"not null" json:"interaction_type"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Category struct {
	CategoryID uint      `gorm:"primaryKey;autoIncrement" json:"category_id"`
	Name       string    `gorm:"unique;not null" json:"name"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
