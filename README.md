# 🧾 FinBoard Backend

基于 **Go + Wails** 开发的个人财务管理应用后端，支持 **桌面应用模式** 和 **服务端模式**，可配合前端进行账单导入、统计与分析。

---

## ✨ 功能特性

- 🖥️ **桌面应用模式**：Wails 打包的跨平台桌面客户端
- 🌐 **服务端模式**：Web API 服务，可部署到服务器使用
- 🔐 **安全加密**：支持 **RSA + AES** 混合加密，保障数据安全
- 🌍 **国际化支持**：内置中英文多语言切换
- 📊 **账单管理**：支持支付宝/微信账单导入与可视化分析

---

## 🚀 构建与运行

### 快速构建

```bash
# 构建所有版本
./build.sh all

# 或者分别构建
./build.sh desktop   # 构建桌面应用
./build.sh server    # 构建服务端应用
```

---

### 桌面应用模式

适合个人用户，提供完整 GUI：

```bash
./build.sh desktop
# 生成可执行文件：
# Windows: FinBoard.exe
# macOS:  FinBoard.app
# Linux:  FinBoard
```

---

### 服务端模式

适合部署到服务器，提供 Web API：

```bash
./build.sh server
./finance-tracker-server -server
# 默认启动地址: http://localhost:9090
```

---

## ⚙️ 开发模式

```bash
# 服务端开发模式
go run . -server

# 桌面应用开发模式
wails dev
```

---

## 📂 项目结构

```
├── main.go              # 程序入口，支持两种模式
├── app.go               # Wails 应用定义
├── wails.json           # Wails 配置
├── build.sh             # 构建脚本
├── config/              # 配置文件
├── controller/          # API 控制器
├── middleware/          # 中间件
├── model/               # 数据模型
├── router/              # 路由配置
├── i18n/                # 国际化文件
├── frontend/build/      # 前端打包资源
└── private.pem          # RSA 私钥（需手动生成）
```

---

## 📦 部署说明

### 桌面应用

1. 执行 `./build.sh desktop`​
2. 分发生成的可执行文件给用户
3. 用户直接运行即可使用

### 服务端应用

1. 执行 `./build.sh server`​
2. 将 `finance-tracker-server` 上传到服务器
3. 在服务器运行：

    ```bash
    ./finance-tracker-server -server
    ```
4. 配置 Nginx 等反向代理转发至 `9090` 端口

---

## 🔑 配置与密钥

- ​`config.yaml`：数据库、JWT 等配置
- ​`i18n/`：多语言翻译文件
- ​`private.pem`：RSA 私钥文件（⚠️ 请自行生成，不要直接使用示例）

生成方式示例：

```bash
openssl genrsa -out private.pem 2048
openssl rsa -in private.pem -pubout -out public.pem
```

⚠️ 请勿将真实密钥上传至 GitHub，仅保留 `.example` 文件作为占位符。

---

## 🛠 技术栈

- **后端框架**：Go + Gin + GORM + SQLite
- **前端集成**：嵌入式静态文件（Vue3 + Vuetify）
- **桌面应用**：Wails v2
- **安全加密**：RSA + AES 混合方案
- **国际化**：go-i18n

---

## 📌 注意事项

1. 需要在项目根目录下准备 `private.pem` 文件
2. 前端需先构建到 `frontend/build/` 再执行打包
3. 服务端模式默认占用 `9090` 端口，请在防火墙或代理中开放
4. 桌面应用模式会自动嵌入静态资源，无需额外配置
