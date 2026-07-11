# Ai idea Linux 生产环境部署文档

## 项目信息

- **项目名称**: Ai idea
- **技术栈**: 
  - 前端: React + Vite + TypeScript
  - 后端: Golang + Gin + MySQL + Redis
- **默认端口**: 8094
- **默认管理员**: baishuwan / baishuwan0825

---

## 快速部署（推荐）

### 前置要求

- Linux 服务器（Ubuntu 20.04+ / CentOS 7+ / Debian 10+）
- Docker 20.10+
- Docker Compose 1.29+ 或 Docker Compose V2
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

### 一键部署

```bash
# 1. 克隆代码（或上传到服务器）
git clone https://github.com/coaidev/coai.git
cd coai

# 2. 执行部署脚本
chmod +x deploy.sh
./deploy.sh

# 3. 等待部署完成（首次需要 3-5 分钟）
# 部署成功后访问 http://你的服务器IP:8094
```

---

## 环境变量配置

部署脚本会自动生成 `.env` 文件，你可以根据需要修改：

```bash
# 数据库配置
MYSQL_DB=chatnio
MYSQL_PASSWORD=your_strong_password_here

# Redis 配置（如需密码可设置）
REDIS_PASSWORD=

# JWT 密钥（必须修改为随机字符串）
SECRET=your_random_secret_key_here

# 服务端口
PORT=8094
```

**安全提示**：
- 务必修改 `MYSQL_PASSWORD` 为强密码
- 务必修改 `SECRET` 为随机字符串（至少 32 字符）
- 生产环境建议使用外部数据库

---

## 健康检查

```bash
# 检查所有服务状态
./health-check.sh

# 查看容器日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f backend
```

---

## 常用操作

### 启动服务
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 停止服务
```bash
docker-compose -f docker-compose.prod.yml down
```

### 重启服务
```bash
docker-compose -f docker-compose.prod.yml restart
```

### 更新服务
```bash
# 拉取最新代码
git pull

# 重新部署
./deploy.sh
```

### 备份数据
```bash
# 备份数据库
docker exec coai-mysql mysqldump -uroot -p$MYSQL_PASSWORD chatnio > backup_$(date +%Y%m%d).sql

# 备份 Redis 数据
docker exec coai-redis redis-cli SAVE
cp redis/dump.rdb backup_redis_$(date +%Y%m%d).rdb

# 备份配置和存储
tar -czf backup_storage_$(date +%Y%m%d).tar.gz config/ storage/
```

### 恢复数据
```bash
# 恢复数据库
docker exec -i coai-mysql mysql -uroot -p$MYSQL_PASSWORD chatnio < backup_20260710.sql

# 恢复 Redis
docker cp backup_redis_20260710.rdb coai-redis:/data/dump.rdb
docker-compose -f docker-compose.prod.yml restart redis
```

---

## 使用 Nginx 反向代理（可选）

如果需要使用域名或 HTTPS，可以配置 Nginx：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8094;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

配置 HTTPS（使用 Let's Encrypt）：
```bash
# 安装 certbot
apt install certbot python3-certbot-nginx -y

# 获取证书
certbot --nginx -d your-domain.com

# 自动续期
certbot renew --dry-run
```

---

## 性能优化建议

### 1. 数据库优化
编辑 `docker-compose.prod.yml`，在 MySQL 服务中添加：
```yaml
mysql:
  command: 
    - --max_connections=1000
    - --innodb_buffer_pool_size=1G
    - --query_cache_type=1
    - --query_cache_size=128M
```

### 2. Redis 优化
```yaml
redis:
  command: 
    - redis-server 
    - --appendonly yes
    - --maxmemory 512mb
    - --maxmemory-policy allkeys-lru
```

### 3. 后端并发优化
在 `.env` 中添加：
```bash
GOMAXPROCS=4  # 根据 CPU 核心数调整
```

---

## 监控与日志

### 日志位置
- 应用日志: `./logs/`
- 数据库日志: `docker-compose logs mysql`
- Redis 日志: `docker-compose logs redis`

### 日志轮转配置
创建 `/etc/logrotate.d/coai`：
```
/path/to/coai/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
```

---

## 故障排查

### 1. 服务无法启动
```bash
# 检查端口占用
ss -tuln | grep 8094

# 查看详细日志
docker-compose -f docker-compose.prod.yml logs --tail=100

# 检查磁盘空间
df -h
```

### 2. 数据库连接失败
```bash
# 检查 MySQL 容器状态
docker exec coai-mysql mysqladmin ping -h localhost

# 进入数据库容器
docker exec -it coai-mysql mysql -uroot -p$MYSQL_PASSWORD
```

### 3. 内存不足
```bash
# 查看容器资源占用
docker stats

# 限制容器内存（编辑 docker-compose.prod.yml）
services:
  backend:
    mem_limit: 1g
  mysql:
    mem_limit: 1g
```

---

## 安全加固

1. **防火墙配置**
```bash
# 只开放必要端口
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 22/tcp
ufw enable
```

2. **更改默认密码**
- 首次登录后立即修改管理员密码
- 定期更换数据库密码

3. **定期更新**
```bash
# 更新系统
apt update && apt upgrade -y

# 更新 Docker 镜像
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

---

## 端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| CoAI 后端 | 8094 | 主服务端口（包含前端静态文件） |
| MySQL | 3306 | 数据库（仅容器内部访问） |
| Redis | 6379 | 缓存（仅容器内部访问） |

---

## 技术支持

- 官方文档: https://coai.dev/docs
- GitHub 仓库: https://github.com/coaidev/coai
- Discord 社区: https://discord.gg/rpzNSmqaF2

---

**部署日期**: 2026-07-10  
**文档版本**: 1.0  
**维护者**: Crow5 Agent
