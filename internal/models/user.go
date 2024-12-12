package models

import "time"

type User struct {
	UserID       uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	RoleID       int       `gorm:"default:0;not null" json:"role_id"`
	Reputation   int       `gorm:"default:0" json:"reputation"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsBanned     bool      `gorm:"default:false" json:"is_banned"`
}
