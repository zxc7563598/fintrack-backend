package model

import (
	"encoding/base64"
	"time"

	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Email          string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password       string         `gorm:"size:255;not null" json:"password"`
	Salt           string         `gorm:"size:64;not null;comment:随机盐" json:"salt"`
	DeepseekApiKey string         `gorm:"size:100;default:'';comment:deepseek密钥" json:"deepseek_api_key"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// 验证密码是否正确
func (u *User) VerifyPassword(password string) bool {
	if u.Salt == "" || u.Password == "" {
		return false
	}
	saltBytes, err := base64.RawStdEncoding.DecodeString(u.Salt)
	if err != nil {
		return false
	}
	hash := helpers.HashPassword(password, saltBytes)
	return hash == u.Password
}
