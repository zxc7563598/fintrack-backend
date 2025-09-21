package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 SQLite 数据库
func InitDB() {
	var err error
	// 获取数据库文件路径
	dbPath := helpers.GetDataPath(Cfg.Database.SqlitePath)
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("无法创建数据库目录: %v", err)
	}
	DB, err = gorm.Open(sqlite.Open(filepath.Join(dbPath, "finance.db")), &gorm.Config{})
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
