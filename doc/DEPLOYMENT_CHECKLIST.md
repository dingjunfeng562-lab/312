# Ai idea Linux 部署清单

## 已完成的优化工作

### ✅ 清理冗余文件
- [x] 删除 8 个 Windows 可执行文件（约 344MB）
- [x] 删除 42 个临时 Python 调试脚本
- [x] 删除所有 .bat/.ps1 批处理脚本
- [x] 删除散落的日志文件（*.log, *.err, *.txt）
- [x] 删除 Go 缓存目录（1.1GB）

### ✅ 生成生产部署配置
- [x] docker-compose.prod.yml - 生产环境 Docker 编排
- [x] Dockerfile.prod - 多阶段构建配置
- [x] deploy.sh - 一键部署脚本
- [x] health-check.sh - 健康检查脚本
- [x] doc/DEPLOYMENT.md - 详细部署文档
- [x] PROJECT_STRUCTURE.md - 项目结构说明
- [x] 更新 .gitignore - 防止再次污染

---

## 部署文件说明

### 1. docker-compose.prod.yml
**生产环境 Docker Compose 配置**
- MySQL 8.0 + Redis 7 + Ai idea 后端
- 内置健康检查机制
- 自动重启策略
- 数据持久化（./db, ./redis, ./config, ./logs, ./storage）

### 2. Dockerfile.prod
**多阶段构建 Dockerfile**
- 阶段 1: Node.js 构建前端（pnpm build）
- 阶段 2: Golang 编译后端（静态链接）
- 阶段 3: Alpine 最小运行镜像
- 最终镜像大小约 50-80MB

### 3. deploy.sh
**一键部署脚本**（使用方法：`chmod +x deploy.sh && ./deploy.sh`）
- 自动检查 Docker/Docker Compose 依赖
- 自动检查端口占用
- 自动生成 .env 配置文件（含随机密码和 JWT Secret）
- 自动构建镜像并启动服务
- 自动等待服务就绪
- 显示访问地址和管理员账号

### 4. health-check.sh
**健康检查脚本**（使用方法：`./health-check.sh`）
- 检查容器状态（docker ps）
- 检查端口监听（netstat/ss）
- 检查 HTTP 健康端点（curl /health）
- 检查数据库连接（MySQL + Redis）
- 显示资源占用（docker stats）

### 5. doc/DEPLOYMENT.md
**完整部署文档**，包含：
- 快速部署指南
- 环境变量配置说明
- 健康检查方法
- 常用运维命令
- Nginx 反向代理配置
- 性能优化建议
- 监控与日志管理
- 故障排查指南
- 安全加固建议
- 备份与恢复步骤

---

## Linux 服务器部署步骤

### 方式一：直接使用（推荐）

```bash
# 1. 将整个项目上传到 Linux 服务器
scp -r coai/ user@server:/opt/

# 2. 登录服务器
ssh user@server

# 3. 进入项目目录
cd /opt/coai

# 4. 执行一键部署
chmod +x deploy.sh
./deploy.sh

# 5. 等待部署完成（3-5 分钟）
# 访问 http://服务器IP:8094
```

### 方式二：使用 Git 克隆

```bash
# 1. 在服务器上克隆
git clone https://github.com/coaidev/coai.git
cd coai

# 2. 执行部署
chmod +x deploy.sh
./deploy.sh
```

---

## 部署后验证

### 1. 检查容器状态
```bash
docker-compose -f docker-compose.prod.yml ps
```
应该看到 3 个容器：coai-backend, coai-mysql, coai-redis 全部为 `Up` 状态。

### 2. 运行健康检查
```bash
chmod +x health-check.sh
./health-check.sh
```
应该看到所有检查项都是 ✓ 绿色通过。

### 3. 访问测试
```bash
# 本地测试
curl http://localhost:8094/health

# 浏览器访问
http://服务器IP:8094
```

### 4. 登录后台
- 访问首页
- 点击登录
- 账号：`baishuwan`
- 密码：`baishuwan0825`
- **立即修改密码！**

---

## 环境变量配置

部署脚本会自动生成 `.env` 文件，包含随机密码和 JWT Secret。

**重要配置项：**
```bash
# 数据库密码（自动生成，可手动修改）
MYSQL_PASSWORD=auto_generated_strong_password

# JWT 密钥（自动生成，可手动修改）
SECRET=auto_generated_random_secret

# 服务端口（默认 8094）
PORT=8094
```

**修改配置后需要重启：**
```bash
docker-compose -f docker-compose.prod.yml restart
```

---

## 数据持久化

以下目录会自动创建并持久化数据：

| 目录 | 用途 | 是否需要备份 |
|------|------|--------------|
| `./db/` | MySQL 数据库文件 | ✅ 必须备份 |
| `./redis/` | Redis 持久化数据 | ✅ 建议备份 |
| `./config/` | 运行时配置 | ✅ 建议备份 |
| `./storage/` | 用户上传文件 | ✅ 必须备份 |
| `./logs/` | 应用日志 | ❌ 可不备份 |

---

## 性能优化建议

### 1. 如果服务器内存 >= 4GB
编辑 `docker-compose.prod.yml`，增加 MySQL 缓冲池：
```yaml
mysql:
  command: 
    - --innodb_buffer_pool_size=1G
```

### 2. 如果需要高并发
在 `.env` 中添加：
```bash
GOMAXPROCS=4  # 设置为 CPU 核心数
```

### 3. 如果需要 HTTPS
参考 `doc/DEPLOYMENT.md` 中的 Nginx + Let's Encrypt 配置。

---

## 常见问题

### Q1: 端口 8094 被占用怎么办？
修改 `.env` 中的 `PORT` 为其他端口（如 8095），然后重启。

### Q2: 内存不够怎么办？
至少需要 2GB 内存。如果不够，可以减少 MySQL 缓冲：
```yaml
mysql:
  command: --innodb_buffer_pool_size=256M
```

### Q3: 如何查看日志？
```bash
# 查看所有日志
docker-compose -f docker-compose.prod.yml logs -f

# 只看后端日志
docker-compose -f docker-compose.prod.yml logs -f backend
```

### Q4: 如何备份数据？
```bash
# 备份数据库
docker exec coai-mysql mysqldump -uroot -p密码 chatnio > backup.sql

# 备份所有数据目录
tar -czf coai_backup.tar.gz db/ redis/ config/ storage/
```

### Q5: 如何更新到最新版本？
```bash
git pull
./deploy.sh  # 会自动重新构建
```

---

## 与 Windows 开发环境的区别

| 项目 | Windows 开发 | Linux 生产 |
|------|--------------|------------|
| 启动方式 | .exe 直接运行 | Docker 容器 |
| 数据库 | SQLite / 外部 MySQL | Docker MySQL |
| Redis | redis-portable | Docker Redis |
| 前端 | pnpm dev 热更新 | 构建后静态文件 |
| 端口 | 可能多个端口 | 统一 8094 |
| 日志 | 控制台输出 | 持久化到文件 |

---

## 下一步建议

1. ✅ **立即修改管理员密码**
2. ✅ **配置 Nginx 反向代理**（如需域名访问）
3. ✅ **配置 SSL 证书**（生产环境必须）
4. ✅ **设置定时备份**（crontab）
5. ✅ **配置防火墙规则**
6. ✅ **启用日志轮转**（logrotate）
7. ✅ **配置监控告警**（可选）

---

## 技术支持

- 📖 详细部署文档: [doc/DEPLOYMENT.md](DEPLOYMENT.md)
- 📖 项目结构说明: [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md)
- 💬 官方 Discord: https://discord.gg/rpzNSmqaF2
- 🐛 问题反馈: https://github.com/coaidev/coai/issues

---

**清单生成时间**: 2026-07-10  
**Crow5 Agent**: 优化完成，已就绪部署 🚀
