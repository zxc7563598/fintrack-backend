package main

import (
	"embed"
	"flag"
	"log"
	"os"

	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/i18n"
	"github.com/zxc7563598/fintrack-backend/middleware"
	"github.com/zxc7563598/fintrack-backend/router"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

//go:embed i18n
var i18nFiles embed.FS

//go:embed config.yaml
var configFile embed.FS

//go:embed private.pem
var privateKeyFile embed.FS

func main() {
	os.Setenv("GODEBUG", "netdns=cgo")
	// 解析命令行参数
	serverMode := flag.Bool("server", false, "在服务器模式下运行")
	flag.Parse()
	if *serverMode {
		// 服务端模式
		runServerMode()
	} else {
		// Wails桌面应用模式
		runWailsMode()
	}
}
func runServerMode() {
	log.Println("🚀 将FinBoard作为服务器web端启动...")
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
	// 引入路由
	r := router.SetupRouter()
	log.Println("🌐 服务器正在运行于 http://localhost:9090")
	r.Run(":9090")
}

func runWailsMode() {
	log.Println("🖥️  将FinBoard作为桌面应用程序启动...")
	// 创建应用实例
	app := NewApp()
	// 创建应用选项
	err := wails.Run(&options.App{
		Title:  "FinBoard",
		Width:  1440,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnDomReady:       app.OnDomReady,
		OnBeforeClose:    app.OnBeforeClose,
		OnShutdown:       app.OnShutdown,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		log.Fatal("无法启动Wails应用程序:", err)
	}
}
