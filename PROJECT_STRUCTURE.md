# Ai idea 项目结构

本项目为 **Ai idea** - 下一代 AIGC 商业一体化解决方案。

## 目录结构

```
coai/
├── adapter/          # 模型适配器（OpenAI、Claude、Gemini 等）
├── addition/         # 扩展功能模块
├── admin/            # 后台管理模块
├── app/              # 前端源码（React + Vite + TypeScript）
│   ├── src/          # 前端源代码
│   ├── public/       # 静态资源
│   └── dist/         # 构建产物（部署时生成）
├── auth/             # 认证授权模块
├── channel/          # 渠道管理模块
├── cli/              # 命令行工具
├── config/           # 运行时配置目录（Docker 挂载）
├── connection/       # 连接管理
├── database/         # 数据库初始化脚本
├── db/               # MySQL 数据持久化目录（Docker 挂载）
├── doc/              # 项目文档
│   └── DEPLOYMENT.md # Linux 部署文档
├── globals/          # 全局配置与常量
├── logs/             # 应用日志目录
├── manager/          # 业务管理模块
├── middleware/       # 中间件
├── migration/        # 数据库迁移脚本
├── prototype/        # 产品原型设计
├── redis/            # Redis 数据持久化目录（Docker 挂载）
├── screenshot/       # 项目截图与宣传图
├── storage/          # 用户上传文件存储目录
├── tools/            # 开发工具脚本
├── utils/            # 公共工具函数
├── go.mod            # Go 依赖管理
├── go.sum            # Go 依赖锁定
├── main.go           # 后端主入口
├── docker-compose.yml        # 开发环境 Docker 配置
├── docker-compose.prod.yml   # 生产环境 Docker 配置
├── Dockerfile.prod           # 生产环境多阶段构建
├── deploy.sh                 # 一键部署脚本
├── health-check.sh           # 健康检查脚本
└── README.md                 # 项目说明
```

## 技术栈

### 前端
- **框架**: React 18
- **构建工具**: Vite 4
- **语言**: TypeScript
- **状态管理**: Redux Toolkit
- **UI 框架**: Radix UI + Tailwind CSS
- **图表**: Tremor Charts
- **Markdown 渲染**: react-markdown + rehype/remark
- **PWA**: Workbox

### 后端
- **语言**: Golang 1.21+
- **Web 框架**: Gin
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **ORM**: 自实现（基于 database/sql）
- **WebSocket**: Gorilla WebSocket
- **日志**: Logrus + Lumberjack
- **配置管理**: Viper

### 部署
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx（可选）
- **进程管理**: Docker 自动重启
- **健康检查**: 内置 HTTP /health 端点

## 端口分配

| 服务 | 端口 | 说明 |
|------|------|------|
| CoAI 主服务 | 8094 | 后端 API + 前端静态文件 |
| MySQL | 3306 | 数据库（仅容器内部） |
| Redis | 6379 | 缓存（仅容器内部） |

## 快速开始

### 开发环境
```bash
# 前端开发
cd app
pnpm install
pnpm dev

# 后端开发
go mod download
go run main.go
```

### 生产部署（Linux）
```bash
# 一键部署（推荐）
chmod +x deploy.sh
./deploy.sh

# 手动部署
docker-compose -f docker-compose.prod.yml up -d

# 健康检查
./health-check.sh
```

详细部署文档请查看：[doc/DEPLOYMENT.md](doc/DEPLOYMENT.md)

## 默认账号

- **管理员账号**: `baishuwan`
- **默认密码**: `baishuwan0825`

⚠️ **首次登录后请立即修改密码！**

## 主要特性

1. ✅ 多模型支持（OpenAI、Claude、Gemini、Midjourney 等）
2. ✅ 美观的 UI 设计（PC/平板/移动端适配）
3. ✅ 完整的 Markdown 支持（LaTeX、Mermaid、代码高亮）
4. ✅ 对话同步与分享
5. ✅ 文件解析（PDF、Docx、Excel、图片）
6. ✅ 订阅制 + 弹性计费
7. ✅ 强大的渠道管理（多渠道、负载均衡、自动重试）
8. ✅ OpenAI API 兼容代理
9. ✅ PWA 支持
10. ✅ 国际化（中文/英文/日文）

## 数据库模式

数据库初始化脚本位于 `database/` 目录，首次启动时会自动执行。

持久化数据存储在：
- MySQL: `./db/`
- Redis: `./redis/`
- 上传文件: `./storage/`
- 应用日志: `./logs/`

## 环境变量

主要环境变量（在 `.env` 中配置）：

```bash
# 数据库
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DB=chatnio
MYSQL_USER=root
MYSQL_PASSWORD=your_password

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# JWT 密钥
SECRET=your_random_secret

# 服务配置
PORT=8094
SERVE_STATIC=true
```

## 常见问题

### 1. 端口被占用
修改 `.env` 中的 `PORT` 配置，然后重启服务。

### 2. 数据库连接失败
检查 MySQL 容器是否正常运行：
```bash
docker-compose -f docker-compose.prod.yml logs mysql
```

### 3. 前端资源 404
确保环境变量 `SERVE_STATIC=true`，并检查前端是否已构建。

## 技术支持

- 📖 官方文档: https://coai.dev/docs
- 💬 Discord 社区: https://discord.gg/rpzNSmqaF2
- 🐛 问题反馈: https://github.com/coaidev/coai/issues

## 开源协议

Apache-2.0 License - 商业友好

---

**最后更新**: 2026-07-10  
**维护者**: Crow5 Agent
