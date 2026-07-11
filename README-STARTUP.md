# Chat AI 启动指南

## 🚀 快速启动

### 方法 1：一键启动（推荐）

双击运行：
```
start-all-services.bat
```

然后选择 `[1] 启动所有服务`

### 方法 2：分步启动

#### 1. 启动 Redis（必须第一步）
```batch
cd redis-portable
redis-server.exe
```

#### 2. 启动后端
```batch
start-backend-dev.bat
```

#### 3. 启动前端
```batch
start-frontend-dev.bat
```

## 🌐 访问地址

- **前端界面**: http://localhost:5173
- **后端API**: http://localhost:8094
- **API文档**: http://localhost:8094/api/docs

## 🔧 常见问题解决

### 问题 1: Redis 连接失败
**症状**: 日志显示 "failed to connect to redis"

**解决方案**:
1. 确保 Redis 已启动
2. 检查端口 6379 是否被占用：
   ```batch
   netstat -ano | findstr "6379"
   ```
3. 重启 Redis：
   ```batch
   taskkill /F /IM redis-server.exe
   cd redis-portable
   redis-server.exe
   ```

### 问题 2: 前端模块加载失败
**症状**: "Failed to fetch dynamically imported module"

**解决方案**:
运行修复脚本：
```batch
fix-frontend.bat
```

或使用服务管理器选择 `[5] 修复并重启前端`

### 问题 3: 端口被占用
**症状**: "Address already in use"

**解决方案**:
1. 查看端口占用：
   ```batch
   netstat -ano | findstr "5173"  # 前端
   netstat -ano | findstr "8094"  # 后端
   netstat -ano | findstr "6379"  # Redis
   ```

2. 停止占用进程：
   ```batch
   taskkill /F /PID [进程ID]
   ```

### 问题 4: 数据库迁移错误
**症状**: "duplicate column name: task_id"

**解决方案**:
这是非致命错误，可以忽略。如需完全修复：
```batch
# 删除数据库（会丢失数据！）
del db\chatnio.db
# 重启后端服务
```

## 📊 服务状态检查

使用服务管理器的 `[7] 查看服务状态` 选项，或手动检查：

### 检查 Redis
```batch
cd redis-portable
redis-cli.exe ping
# 应返回: PONG
```

### 检查后端
```batch
curl http://localhost:8094/api/v1/state
```

### 检查前端
```
浏览器访问: http://localhost:5173
```

## 🛑 停止服务

### 使用服务管理器
选择 `[6] 停止所有服务`

### 手动停止
```batch
taskkill /F /IM redis-server.exe
taskkill /F /IM coai-dev.exe
taskkill /F /IM node.exe
```

## 📝 日志文件位置

| 服务 | 日志文件 |
|------|----------|
| 后端标准输出 | `backend.dev.log` |
| 后端错误输出 | `backend.dev.err.log` |
| 前端标准输出 | `frontend.dev.log` |
| 前端错误输出 | `frontend.dev.err.log` |
| Redis | `redis.out.log` |

使用服务管理器的 `[8] 查看日志` 选项快速查看。

## 🔄 更新和维护

### 更新前端依赖
```batch
cd app
npm update
```

### 更新后端
```batch
# 根据你的构建流程重新编译
go build -o coai-dev.exe
```

### 清理缓存
```batch
# 前端缓存
cd app
rmdir /s /q node_modules\.vite
npm cache clean --force

# Redis 缓存
redis-cli.exe FLUSHALL
```

## 🎯 开发建议

1. **始终先启动 Redis** - 这是最常见的错误来源
2. **保持终端窗口打开** - 方便查看实时日志
3. **使用硬刷新** - 前端修改后按 `Ctrl+Shift+R`
4. **定期清理缓存** - 每周运行一次 `fix-frontend.bat`
5. **检查端口冲突** - 确保 6379/8094/5173 端口可用

## 📞 获取帮助

如果遇到问题：
1. 查看日志文件（服务管理器选项 8）
2. 检查服务状态（服务管理器选项 7）
3. 尝试修复前端（服务管理器选项 5）
4. 完全重启所有服务（选项 6 然后选项 1）

## 🏗️ 项目结构

```
coai/
├── app/                    # 前端代码
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── redis-portable/         # Redis 便携版
├── connection/             # 数据库连接
├── *.go                    # 后端代码
├── coai-dev.exe           # 后端可执行文件
├── start-all-services.bat  # 服务管理器
├── fix-frontend.bat        # 前端修复脚本
└── config.yaml            # 配置文件
```

---

**版本**: 1.0  
**最后更新**: 2026/7/9
