package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/zxc7563598/fintrack-backend/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 SQLite 数据库
func InitDB() {
	var err error
	// 获取数据库文件路径
	dbPath := getDBPath()
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("无法创建数据库目录: %v", err)
	}
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
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

// getDBPath 获取数据库文件路径
func getDBPath() string {
	// 如果配置中的路径是绝对路径，直接使用
	if filepath.IsAbs(Cfg.Database.SqlitePath) {
		return Cfg.Database.SqlitePath
	}
	// 检查是否在Wails应用环境中
	// 通过检查是否存在特定的环境变量或文件来判断
	if isWailsApp() {
		// 在Wails应用中，使用用户主目录下的应用数据目录
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("无法获取用户主目录: %v", err)
		}
		// 创建应用数据目录
		appDataDir := filepath.Join(homeDir, ".finance-tracker")
		if err := os.MkdirAll(appDataDir, 0755); err != nil {
			log.Fatalf("无法创建应用数据目录: %v", err)
		}
		// 使用应用数据目录中的数据库文件
		return filepath.Join(appDataDir, "finance.db")
	}
	// 开发环境或服务端模式，使用相对路径
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取工作目录: %v", err)
	}
	// 构建绝对路径
	return filepath.Join(workDir, Cfg.Database.SqlitePath)
}

// isWailsApp 检查是否在Wails应用环境中运行
func isWailsApp() bool {
	// 检查是否存在Wails相关的环境变量或文件
	// 这里使用一个简单的方法：检查当前可执行文件的名称
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	execName := filepath.Base(execPath)
	// 如果可执行文件名为"财务管理系统"，说明是Wails应用
	return execName == "财务管理系统"
}
