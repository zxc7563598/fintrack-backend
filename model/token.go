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
	CreatedAt    time.Time      `json:"created_at"`              // 创建时间，自动填充
	UpdatedAt    time.Time      `json:"updated_at"`              // 更新时间，自动更新
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 逻辑删除
}
