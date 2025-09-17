package main

import (
	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/i18n"
	"github.com/zxc7563598/fintrack-backend/router"
)

func main() {
	// 初始化配置
	config.InitConfig()
	// 初始化语言
	i18n.InitI18n()
	// 初始化 SQLite
	config.InitDB()
	// 引入路由
	r := router.SetupRouter()
	r.Run(":9090")
}
