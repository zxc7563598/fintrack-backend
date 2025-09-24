#!/bin/bash

# FinBoard 构建脚本
# 支持构建桌面应用和服务端应用

set -e

echo "🚀 FinBoard 构建脚本"
echo "================================"

# 检查参数
if [ "$1" = "server" ]; then
    echo "📦 构建服务端应用..."
    go build -o finance-tracker-server .
    echo "✅ 服务端应用构建完成: finance-tracker-server"
    echo "💡 运行方式: ./finance-tracker-server -server"
elif [ "$1" = "desktop" ]; then
    echo "📦 构建桌面应用..."
    wails build
    echo "✅ 桌面应用构建完成"
    echo "💡 运行方式: 直接运行生成的应用程序"
elif [ "$1" = "all" ]; then
    echo "📦 构建所有版本..."
    echo "  🔨 构建服务端应用..."
    go build -o finance-tracker-server .
    echo "  🔨 构建桌面应用..."
    wails build
    echo "✅ 所有版本构建完成"
    echo "💡 服务端运行方式: ./finance-tracker-server -server"
    echo "💡 桌面应用运行方式: 直接运行生成的应用程序"
else
    echo "使用方法:"
    echo "  ./build.sh server   - 构建服务端应用"
    echo "  ./build.sh desktop  - 构建桌面应用"
    echo "  ./build.sh all      - 构建所有版本"
    echo ""
    echo "示例:"
    echo "  ./build.sh server   # 构建服务端，用户可部署到服务器"
    echo "  ./build.sh desktop  # 构建桌面应用，用户本地使用"
    echo "  ./build.sh all      # 构建两个版本"
fi

