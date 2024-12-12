package models

import "time"

type Category struct {
	CategoryID uint      `gorm:"primaryKey;autoIncrement" json:"category_id"`
	Name       string    `gorm:"unique;not null" json:"name"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
