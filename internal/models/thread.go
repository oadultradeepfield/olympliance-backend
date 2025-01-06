package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

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
