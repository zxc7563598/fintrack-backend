package main

import (
	"context"
	"log"

	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/i18n"
	"github.com/zxc7563598/fintrack-backend/middleware"
	"github.com/zxc7563598/fintrack-backend/router"
)

// App 结构体
type App struct {
	ctx context.Context
}

// 创建一个新的 App 应用程序结构体
func NewApp() *App {
	return &App{}
}

// 当应用启动时，会调用 OnStartup。上下文会被保存
// 这样我们就可以调用运行时方法了
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// 在前端 DOM 加载完成后被调用
func (a *App) OnDomReady(ctx context.Context) {
	// 初始化配置
	config.SetConfigFS(configFile)
	config.InitConfig()
	// 初始化语言
	i18n.SetI18nFS(i18nFiles)
	i18n.InitI18n()
	// 初始化 SQLite
	config.InitDB()
	// 设置私钥文件系统
	middleware.SetPrivateKeyFS(privateKeyFile)
	// 启动后端服务器
	go a.startBackendServer()
	log.Println("✅ Wails 应用程序已成功初始化")
}

// 启动后端服务器
func (a *App) startBackendServer() {
	r := router.SetupRouter()
	log.Println("🌐 后端服务器正在启动中 http://localhost:9090")
	if err := r.Run(":9090"); err != nil {
		log.Printf("后端服务器错误: %v", err)
	}
}

// 在应用即将退出前调用
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	return false
}

// 在应用程序关闭时调用
func (a *App) OnShutdown(ctx context.Context) {
	log.Println("Wails 应用正在关闭")
}

// 获取应用信息
func (a *App) GetAppInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "财务管理系统",
		"version": "1.0.0",
		"mode":    "desktop",
	}
}
