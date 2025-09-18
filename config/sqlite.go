package config

import (
	"log"

	"github.com/zxc7563598/fintrack-backend/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 SQLite 数据库
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(Cfg.Database.SqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接 SQLite 数据库: %v", err)
	}

	log.Println("✅ SQLite 数据库连接成功")

	// 自动创建表
	err = DB.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.BillRecord{},
		&model.UserMailbox{},
	)
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	log.Println("✅ 数据表自动迁移完成")
}
