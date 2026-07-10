# 邀请码管理模块集成说明

## 概述
邀请码管理功能的后端和组件已完整实现，但前端路由和注册页面尚未集成。

## 需要修改的文件

### 1. `app/src/api/auth.ts`

在第 32 行的 `code: string;` 后添加一行：

```typescript
export type RegisterForm = {
  username: string;
  password: string;
  repassword: string;
  email: string;
  code: string;
  invitation?: string;  // ← 添加这一行
};
```

### 2. `app/src/router.tsx`

#### 2.1 在第 40 行后添加导入：

```typescript
const AdminLogger = lazyFactor(() => import("@/routes/admin/Logger.tsx"));
const AdminInvitation = lazyFactor(() => import("@/routes/admin/Invitation.tsx"));  // ← 添加这一行
```

#### 2.2 在第 259 行后（admin-logger 路由的 `},` 后）添加路由配置：

```typescript
          {
            id: "admin-logger",
            path: "logger",
            element: (
              <Suspense>
                <AdminLogger />
              </Suspense>
            ),
          },
          {  // ← 从这里开始添加
            id: "admin-invitation",
            path: "invitation",
            element: (
              <Suspense>
                <AdminInvitation />
              </Suspense>
            ),
          },  // ← 到这里结束
        ],
```

### 3. `app/src/routes/Register.tsx`

在第 210 行的 `</div>` 和第 212 行的 `<Button` 之间添加邀请码输入字段：

```typescript
      </div>

      <Label>  // ← 从这里开始添加
        {t("auth.invitation-code") || "Invitation Code"}
      </Label>
      <Input
        placeholder={t("auth.invitation-code-placeholder") || "Enter invitation code (optional)"}
        value={form.invitation || ""}
        onChange={(e) =>
          dispatch({
            type: "update:invitation",
            payload: e.target.value,
          })
        }
      />  // ← 到这里结束

      <Button
```

## 验证

修改完成后，运行以下命令验证：

```bash
cd app
npm run build
```

如果编译成功，邀请码管理功能将完全可用：

- 管理后台访问：`http://localhost:5173/admin/invitation`
- 注册页面将显示邀请码输入框（可选填）

## 功能说明

- 邀请码为可选字段，用户可以选择是否填写
- 管理员可以在后台生成、查看、禁用、删除邀请码
- 使用邀请码注册的用户会获得额外的配额奖励
