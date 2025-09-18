package model

import (
	"time"

	"gorm.io/gorm"
)

type UserMailbox struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint           `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Email     string         `gorm:"size:100;not null;comment:邮箱账号" json:"email"`
	AuthCode  string         `gorm:"size:255;not null;comment:邮箱授权码" json:"auth_code"`
	IMAP      string         `gorm:"size:255;not null;comment:IMAP服务器" json:"imap"`
	Remark    string         `gorm:"size:255;comment:备注" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
