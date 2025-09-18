package model

import (
	"time"

	"gorm.io/gorm"
)

// Token 表
type Token struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null;comment:用户ID" json:"user_id"`
	AccessToken  string         `gorm:"uniqueIndex;not null;comment:访问Token" json:"access_token"`
	RefreshToken string         `gorm:"uniqueIndex;not null;comment:刷新Token" json:"refresh_token"`
	ExpiresAt    time.Time      `gorm:"not null;comment:过期时间" json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
