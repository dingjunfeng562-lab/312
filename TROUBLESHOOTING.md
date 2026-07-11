# Chat AI 故障排除指南

## 🔴 致命错误

### 1. Redis 连接失败

#### 症状
```
[WARNING] - failed to connect to redis host: localhost
dial tcp [::1]:6379: connectex: No connection could be made
```

#### 原因
- Redis 服务器未启动
- 端口 6379 被占用
- 防火墙阻止连接

#### 解决步骤
```batch
# 1. 检查 Redis 是否运行
netstat -ano | findstr "6379"

# 2. 如果没有运行，启动 Redis
cd redis-portable
redis-server.exe

# 3. 验证连接
redis-cli.exe ping
# 应返回: PONG

# 4. 如果端口被占用，查找并终止进程
netstat -ano | findstr "6379"
taskkill /F /PID [进程ID]
```

#### 预防措施
- 始终先启动 Redis，再启动其他服务
- 将 Redis 添加到系统服务或开机启动

---

### 2. 前端模块加载失败

#### 症状
```
Failed to fetch dynamically imported module: 
http://localhost:5173/src/routes/Model.tsx
```

#### 原因
- Vite HMR 缓存损坏
- 多个开发服务器实例冲突
- 浏览器缓存问题

#### 解决步骤
```batch
# 方法 1: 使用修复脚本（推荐）
fix-frontend.bat

# 方法 2: 手动修复
cd app
taskkill /F /IM node.exe
rmdir /s /q node_modules\.vite
rmdir /s /q .vite
npm cache clean --force
npm run dev
```

#### 浏览器端修复
1. 按 `Ctrl+Shift+R` 硬刷新
2. 清除浏览器缓存
3. 打开隐私模式测试

---

## 🟡 警告错误

### 3. 数据库迁移错误

#### 症状
```
migration error: SQL logic error: duplicate column name: task_id (1)
```

#### 原因
数据库迁移脚本尝试添加已存在的列

#### 影响
非致命错误，不影响正常使用

#### 解决方案（可选）
```batch
# 方案 1: 忽略（推荐）
# 该错误不会影响系统运行

# 方案 2: 重建数据库（会丢失数据！）
taskkill /F /IM coai-dev.exe
del db\chatnio.db
# 重启后端
```

---

### 4. 端口冲突

#### 症状
```
Error: listen EADDRINUSE: address already in use :::5173
Error: bind: address already in use
```

#### 检查端口占用
```batch
# 检查所有相关端口
netstat -ano | findstr "6379"  # Redis
netstat -ano | findstr "8094"  # 后端
netstat -ano | findstr "5173"  # 前端
```

#### 解决方案
```batch
# 方案 1: 终止占用进程
netstat -ano | findstr "[端口号]"
taskkill /F /PID [进程ID]

# 方案 2: 修改配置使用其他端口
# 编辑 config.yaml (后端)
# 编辑 app/vite.config.ts (前端)
```

---

## 🟢 性能问题

### 5. 前端加载缓慢

#### 症状
页面加载时间超过 10 秒

#### 可能原因
- 开发模式下首次加载
- 网络问题
- 缓存过多

#### 优化方案
```batch
# 1. 清理缓存
cd app
rmdir /s /q node_modules\.vite
npm cache clean --force

# 2. 优化依赖
npm prune
npm dedupe

# 3. 考虑使用生产构建
npm run build
# 然后通过后端的静态文件服务访问
```

---

### 6. 后端响应慢

#### 症状
API 请求响应时间 > 1秒

#### 检查步骤
```batch
# 1. 检查数据库连接
# 查看 backend.dev.log 是否有数据库警告

# 2. 检查 Redis 连接
cd redis-portable
redis-cli.exe ping

# 3. 查看系统资源
taskmgr
# 检查 CPU 和内存使用率
```

#### 解决方案
- 确保 Redis 正常运行
- 检查数据库查询是否有索引
- 重启后端服务

---

## 🛠️ 开发环境问题

### 7. Node.js 版本不兼容

#### 症状
```
Error: The engine "node" is incompatible with this module
```

#### 检查版本
```batch
node --version
# 应为 v16+ 或 v18+
```

#### 解决方案
1. 访问 https://nodejs.org/
2. 下载并安装推荐的 LTS 版本
3. 重新安装依赖：
```batch
cd app
rmdir /s /q node_modules
del package-lock.json
npm install
```

---

### 8. npm 依赖安装失败

#### 症状
```
npm ERR! code ERESOLVE
npm ERR! ERESOLVE unable to resolve dependency tree
```

#### 解决方案
```batch
cd app

# 方案 1: 使用 --legacy-peer-deps
npm install --legacy-peer-deps

# 方案 2: 清理后重装
rmdir /s /q node_modules
del package-lock.json
npm cache clean --force
npm install

# 方案 3: 使用 yarn（备选）
npm install -g yarn
yarn install
```

---

## 📱 浏览器问题

### 9. 浏览器控制台错误

#### 常见错误及解决方案

| 错误信息 | 原因 | 解决方案 |
|---------|------|----------|
| `Failed to fetch` | 后端未运行 | 启动后端服务 |
| `WebSocket connection failed` | 代理配置错误 | 检查 vite.config.ts |
| `CORS error` | 跨域问题 | 检查后端 CORS 配置 |
| `404 Not Found` | 路由问题 | 检查路由配置 |

#### 通用解决步骤
1. 打开浏览器开发者工具 (F12)
2. 查看 Console 标签的错误信息
3. 查看 Network 标签的请求状态
4. 硬刷新页面 (Ctrl+Shift+R)

---

## 🔧 系统级问题

### 10. 防火墙阻止

#### 症状
- 外部设备无法访问
- 某些端口无法绑定

#### 解决方案
```batch
# 添加防火墙规则（管理员权限）
netsh advfirewall firewall add rule name="Redis" dir=in action=allow protocol=TCP localport=6379
netsh advfirewall firewall add rule name="Backend" dir=in action=allow protocol=TCP localport=8094
netsh advfirewall firewall add rule name="Frontend" dir=in action=allow protocol=TCP localport=5173
```

---

### 11. 权限问题

#### 症状
```
Error: EACCES: permission denied
Error: EPERM: operation not permitted
```

#### 解决方案
1. 以管理员身份运行命令提示符
2. 检查文件/文件夹权限
3. 关闭杀毒软件或添加例外

---

## 📊 诊断工具

### 健康检查脚本

创建 `health-check.bat`:
```batch
@echo off
echo ==========================================
echo 系统健康检查
echo ==========================================
echo.

echo [1] 检查 Redis...
cd redis-portable
redis-cli.exe ping && echo ✓ Redis 正常 || echo ✗ Redis 异常

echo.
echo [2] 检查后端...
curl -s http://localhost:8094/api/v1/state >nul && echo ✓ 后端正常 || echo ✗ 后端异常

echo.
echo [3] 检查前端...
curl -s http://localhost:5173 >nul && echo ✓ 前端正常 || echo ✗ 前端异常

echo.
echo [4] 检查端口占用...
netstat -ano | findstr "6379 8094 5173"

echo.
echo [5] 检查进程...
tasklist | findstr "redis-server.exe coai-dev.exe node.exe"

pause
```

---

## 🆘 紧急恢复

### 完全重置

如果所有方法都失败，执行完全重置：

```batch
@echo off
echo 警告：此操作将重置所有服务和缓存！
pause

echo [1/6] 停止所有服务...
taskkill /F /IM redis-server.exe
taskkill /F /IM coai-dev.exe
taskkill /F /IM node.exe

echo [2/6] 清理前端...
cd app
rmdir /s /q node_modules\.vite
rmdir /s /q dist
rmdir /s /q .vite

echo [3/6] 清理 Redis 缓存...
cd ..\redis-portable
redis-cli.exe FLUSHALL

echo [4/6] 清理日志...
cd ..
del *.log

echo [5/6] 重装前端依赖...
cd app
call npm install

echo [6/6] 重启所有服务...
cd ..
call start-all-services.bat

echo.
echo ✓ 重置完成！
pause
```

---

## 📞 获取更多帮助

1. **查看日志**: 使用 `start-all-services.bat` 的选项 8
2. **检查状态**: 使用 `start-all-services.bat` 的选项 7
3. **完整文档**: 查看 `README-STARTUP.md`
4. **GitHub Issues**: 提交问题到项目仓库

---

**最后更新**: 2026/7/9  
**适用版本**: v4.25.0
