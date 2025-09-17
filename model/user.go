package model

import (
	"encoding/base64"
	"time"

	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`         // 用户ID
	Name      string         `gorm:"size:100;not null" json:"name"`              // 名称
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"` // 邮箱
	Password  string         `gorm:"size:255;not null" json:"password"`          // 密码
	Salt      string         `gorm:"size:64;not null;comment:随机盐" json:"salt"`
	CreatedAt time.Time      `json:"created_at"`              // 创建时间，自动填充
	UpdatedAt time.Time      `json:"updated_at"`              // 更新时间，自动更新
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 逻辑删除
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
