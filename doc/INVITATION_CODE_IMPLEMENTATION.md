# 邀请码注册机制实现文档

## 功能概述

为 Ai idea 聊天网站增加了完整的邀请码注册机制，用户注册时必须填写有效的邀请码才能成功注册。管理员可以在后台生成、管理、禁用邀请码。

---

## 一、修改文件清单

### 后端文件修改

1. **数据库迁移文件**
   - `migration/invitation_enhancement.sql` (新增) - MySQL版本迁移脚本
   - `migration/invitation_enhancement_sqlite.sql` (新增) - SQLite版本迁移脚本

2. **邀请码管理逻辑**
   - `admin/invitation.go` - 新增禁用/启用邀请码功能
   - `admin/controller.go` - 新增禁用/启用接口、优化生成接口
   - `admin/router.go` - 新增 `/admin/invitation/disable` 和 `/admin/invitation/enable` 路由
   - `auth/invitation.go` - 增强邀请码校验，支持状态检查

### 前端文件修改

3. **前端API接口**
   - `app/src/admin/api/invitation.ts` (新增) - 邀请码管理API接口定义

4. **管理页面**
   - `app/src/routes/admin/Invitation.tsx` (新增) - 邀请码管理页面

5. **注册页面**
   - `app/src/routes/Register.tsx` - 增加邀请码必填校验和错误提示

6. **路由配置**
   - `app/src/router.tsx` - 添加邀请码管理页面路由

7. **菜单配置**
   - `app/src/components/admin/MenuBar.tsx` - 添加邀请码管理菜单项

8. **国际化翻译**
   - `app/src/resources/i18n/en.json` - 添加英文翻译
   - `app/src/resources/i18n/cn.json` - 更新中文翻译

---

## 二、数据库变更

### 新增字段（invitation表）

| 字段名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| status | VARCHAR(20) | 'unused' | 状态：unused（未使用）、used（已使用）、disabled（已禁用） |
| used_at | DATETIME | NULL | 使用时间 |
| used_ip | VARCHAR(45) | NULL | 使用者IP地址 |
| expires_at | DATETIME | NULL | 过期时间，NULL表示永久有效 |
| creator_id | INT | NULL | 创建者管理员ID |
| creator_name | VARCHAR(255) | 'system' | 创建者名称 |
| notes | TEXT | NULL | 备注信息 |

### 索引

- `idx_invitation_status` - 状态索引
- `idx_invitation_expires_at` - 过期时间索引
- `idx_invitation_creator_id` - 创建者索引

---

## 三、功能说明

### 1. 前台注册功能

#### 邀请码必填
- 用户注册时，邀请码输入框标记为必填项
- 未填写邀请码时，提示"注册需要邀请码"
- 邀请码校验失败时，给出明确的错误提示：
  - 邀请码不存在
  - 邀请码已被使用
  - 邀请码已被禁用
  - 邀请码已过期

#### 注册流程
1. 用户填写用户名、密码
2. 点击"下一步"
3. 填写邮箱、验证码、**邀请码**（必填）
4. 提交注册
5. 后端校验邀请码有效性
6. 注册成功后，邀请码自动绑定到该用户，记录使用时间和IP

### 2. 后台管理功能

#### 邀请码列表
- 显示所有邀请码及其状态
- 支持分页查看
- 显示字段：
  - 邀请码
  - 配额
  - 状态（未使用/已使用/已禁用/已过期）
  - 使用人
  - 创建者
  - 创建时间
  - 过期时间

#### 生成邀请码
- **普通模式**：
  - 类型（前缀）：如 AI、VIP 等
  - 配额：每个邀请码赠送的积分
  - 数量：1-100个

- **高级模式**（超级管理员）：
  - 支持设置过期时间（天数，0表示永久有效）
  - 支持添加备注信息
  - 自动记录创建者信息

#### 邀请码操作
- **复制**：一键复制邀请码到剪贴板
- **禁用**：禁用未使用的邀请码（已使用的不能禁用）
- **启用**：重新启用已禁用的邀请码
- **删除**：删除邀请码（需二次确认）

---

## 四、API接口清单

### 管理员接口（需要管理员权限）

1. **GET** `/admin/invitation/list?page=0`
   - 获取邀请码列表（分页）

2. **POST** `/admin/invitation/generate`
   - 生成邀请码（普通模式）
   - 请求体：
     ```json
     {
       "type": "AI",
       "quota": 10,
       "number": 5
     }
     ```

3. **POST** `/admin/invitation/generate-advanced`（超级管理员）
   - 生成邀请码（高级模式）
   - 请求体：
     ```json
     {
       "type": "VIP",
       "quota": 100,
       "number": 10,
       "expires_days": 30,
       "notes": "新用户活动"
     }
     ```

4. **POST** `/admin/invitation/disable`
   - 禁用邀请码
   - 请求体：
     ```json
     {
       "code": "AI-xxxxxxxxxxxxxxxxxxxx"
     }
     ```

5. **POST** `/admin/invitation/enable`
   - 启用邀请码
   - 请求体：
     ```json
     {
       "code": "AI-xxxxxxxxxxxxxxxxxxxx"
     }
     ```

6. **POST** `/admin/invitation/delete`
   - 删除邀请码
   - 请求体：
     ```json
     {
       "code": "AI-xxxxxxxxxxxxxxxxxxxx"
     }
     ```

7. **GET** `/admin/invitation/usage/:code`
   - 查看邀请码使用详情

8. **GET** `/admin/invitation/check-expired`
   - 检查过期的邀请码

---

## 五、部署步骤

### 1. 执行数据库迁移

#### MySQL数据库
```bash
mysql -u用户名 -p密码 数据库名 < migration/invitation_enhancement.sql
```

#### SQLite数据库
```bash
sqlite3 db/chatnio.db < migration/invitation_enhancement_sqlite.sql
```

### 2. 编译后端代码
```bash
go build -o chatnio.exe main.go
```

### 3. 编译前端代码
```bash
cd app
pnpm install
pnpm build
```

### 4. 配置邀请码模式

在 `config.yaml` 或系统环境变量中设置：
```yaml
invitation_only: true  # 开启邀请码强制注册
```

或使用环境变量：
```bash
export INVITATION_ONLY=true
```

### 5. 重启服务
```bash
./chatnio.exe
```

---

## 六、测试步骤

### 测试1：数据库迁移验证

1. 查看 invitation 表结构：
   ```sql
   -- MySQL
   DESCRIBE invitation;
   
   -- SQLite
   PRAGMA table_info(invitation);
   ```

2. 确认新增字段：status, used_at, used_ip, expires_at, creator_id, creator_name, notes

### 测试2：后台生成邀请码

1. 登录管理员账号（默认用户名：`baishuwan`，密码：`baishuwan0825`）
2. 进入后台管理 → 邀请码管理（Invitation Code Management）
3. 点击"Generate Codes"按钮
4. 填写表单：
   - Type: `TEST`
   - Quota: `10`
   - Number: `3`
5. 点击"Generate"
6. 验证：
   - ✅ 成功生成3个邀请码
   - ✅ 邀请码自动复制到剪贴板
   - ✅ 列表中显示新生成的邀请码
   - ✅ 状态显示为"Unused"

### 测试3：高级模式生成邀请码

1. 勾选"Advanced Mode"
2. 填写表单：
   - Type: `VIP`
   - Quota: `100`
   - Number: `2`
   - Expires in: `7`（7天后过期）
   - Notes: `测试邀请码`
3. 点击"Generate"
4. 验证：
   - ✅ 成功生成2个邀请码
   - ✅ 过期时间显示为7天后
   - ✅ 创建者显示为当前管理员用户名

### 测试4：邀请码操作

1. **复制邀请码**
   - 点击邀请码行的复制按钮
   - 验证：✅ 邀请码已复制到剪贴板

2. **禁用邀请码**
   - 点击未使用邀请码的"禁用"按钮
   - 验证：✅ 状态变更为"Disabled"

3. **启用邀请码**
   - 点击已禁用邀请码的"启用"按钮
   - 验证：✅ 状态恢复为"Unused"

4. **删除邀请码**
   - 点击邀请码的"删除"按钮
   - 确认删除提示
   - 验证：✅ 邀请码从列表中移除

### 测试5：注册流程（邀请码必填）

1. 退出登录
2. 访问注册页面 `/register`
3. 填写用户名和密码，点击"下一步"
4. 填写邮箱和验证码
5. **不填写邀请码**，直接点击"Register"
6. 验证：
   - ✅ 显示错误提示："注册需要邀请码"
   - ✅ 注册失败

### 测试6：使用有效邀请码注册

1. 复制一个未使用的邀请码（从后台获取）
2. 返回注册页面
3. 填写完整信息，包括邀请码
4. 点击"Register"
5. 验证：
   - ✅ 注册成功
   - ✅ 自动登录
   - ✅ 用户积分增加（根据邀请码配额）
6. 返回后台查看邀请码列表
7. 验证：
   - ✅ 该邀请码状态变为"Used"
   - ✅ 使用人显示为新注册用户
   - ✅ 使用时间已记录

### 测试7：使用无效邀请码注册

1. **测试不存在的邀请码**
   - 输入一个随机字符串
   - 验证：✅ 提示"邀请码无效，请检查后重新输入"

2. **测试已使用的邀请码**
   - 输入已被使用的邀请码
   - 验证：✅ 提示"该邀请码已被使用"

3. **测试已禁用的邀请码**
   - 输入已禁用的邀请码
   - 验证：✅ 提示"该邀请码已失效，请联系管理员"

4. **测试已过期的邀请码**
   - 输入已过期的邀请码（或手动修改数据库设置过期时间）
   - 验证：✅ 提示邀请码已过期

### 测试8：分页功能

1. 生成超过10个邀请码（默认每页10个）
2. 验证：
   - ✅ 显示分页控件
   - ✅ 点击"Next"查看下一页
   - ✅ 点击"Previous"返回上一页
   - ✅ 页码显示正确

### 测试9：权限验证

1. 使用普通用户登录
2. 尝试访问 `/admin/invitation`
3. 验证：
   - ✅ 自动重定向到首页
   - ✅ 无法访问管理页面

### 测试10：邀请码搜索和筛选

1. 生成多个不同类型的邀请码（AI、VIP等）
2. 查看列表中的筛选功能
3. 验证邀请码是否按创建时间倒序排列

---

## 七、常见问题

### Q1：如何临时关闭邀请码强制注册？

**答：** 修改配置文件 `config.yaml`：
```yaml
invitation_only: false
```
或设置环境变量：
```bash
export INVITATION_ONLY=false
```
然后重启服务。

### Q2：如何批量生成大量邀请码？

**答：** 使用高级模式，Number字段最大支持100个。如需更多，可以多次生成。

### Q3：邀请码可以重复使用吗？

**答：** 不可以。每个邀请码只能使用一次，使用后状态会变为"used"。

### Q4：如何查看某个邀请码的使用情况？

**答：** 可以通过以下方式查看：
1. 在列表中找到该邀请码，查看"Used By"列
2. 调用API：`GET /admin/invitation/usage/:code`

### Q5：禁用的邀请码还能启用吗？

**答：** 可以。只要邀请码未被使用，管理员可以随时禁用或启用。

### Q6：邀请码过期后还能使用吗？

**答：** 不能。过期的邀请码会在注册时被拒绝。

---

## 八、安全注意事项

1. **邀请码保密**：生成的邀请码应妥善保管，不要公开泄露
2. **定期清理**：定期检查并删除过期的邀请码
3. **IP记录**：系统会记录邀请码使用者的IP地址，便于追踪
4. **事务保证**：注册过程使用数据库事务，防止同一邀请码被多人同时使用
5. **权限控制**：只有管理员可以访问邀请码管理功能

---

## 九、后续优化建议

1. **邀请码统计**：添加邀请码使用统计图表
2. **导出功能**：支持导出邀请码列表为CSV/Excel
3. **批量操作**：支持批量禁用/删除邀请码
4. **邀请链接**：生成带邀请码的注册链接
5. **邀请奖励**：创建者获得积分奖励
6. **使用限制**：支持设置每个邀请码的使用次数

---

## 十、技术栈

- **后端**: Go + Gin + MySQL/SQLite
- **前端**: React + TypeScript + Vite + TailwindCSS
- **UI组件**: Shadcn UI
- **图标**: Lucide Icons
- **国际化**: i18next

---

## 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues: [项目地址]
- 邮箱: [管理员邮箱]

---

**实施完成日期**: 2026-07-08  
**文档版本**: v1.0
