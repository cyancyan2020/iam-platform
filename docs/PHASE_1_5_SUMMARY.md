# Phase 1-5 交付记录

> 完成日期：2026-06-22 | 累计 28 个单元测试通过 | 编译零错误

---

## Phase 1：项目骨架

**文件：**

| 文件 | 说明 |
|------|------|
| `Makefile` | run / build / migrate-up / migrate-down / test / lint |
| `config/config.yaml` | MySQL DSN + Redis + JWT 配置 |
| `cmd/main.go` | Gin 引擎（Logger + Recovery）+ `/health` 端点 + Viper 加载配置 |
| `db/migrations/000001_init.up.sql` | 创建 `user` 表 |
| `db/migrations/000001_init.down.sql` | 删除 `user` 表 |

---

## Phase 2：用户注册

**文件：**

| 文件 | 说明 |
|------|------|
| `pkg/utils/bcrypt.go` | `HashPassword` + `CheckPassword` |
| `pkg/utils/bcrypt_test.go` | 8 个测试：正常/空/长/超长/正确校验/错误/空哈希/往返 |
| `internal/model/user.go` | GORM User 实体，映射 `user` 表 |
| `internal/repository/user_repository.go` | 接口 + 实现：`Create` / `FindByUsername` |
| `internal/service/user_service.go` | 接口 + 实现：`Register`（防重复注册 + bcrypt 哈希） |
| `internal/handler/user_handler.go` | `POST /api/v1/users/register` |
| `internal/service/user_service_test.go` | 4 个测试：成功/重复/查库异常/写库异常 |

---

## Phase 3：JWT 登录与中间件

**文件：**

| 文件 | 说明 |
|------|------|
| `pkg/jwt/jwt.go` | `GenerateToken` / `ParseToken`，Claims 含 UserID/TenantID/Username/DeviceID/TokenVersion |
| `pkg/jwt/jwt_test.go` | 7 个测试：生成/解析/无效签名/过期/垃圾/空/往返 |
| `internal/middleware/auth.go` | Bearer Token 提取 → 解析 → `c.Set("user", claims)` |
| `internal/middleware/auth_test.go` | 6 个测试：有效/无Token/无效Token/畸形Header/版本不匹配/Claims注入 |
| `internal/service/user_service.go` | 新增 `Login`：查用户 → 验密码 → 生成 Token |
| `internal/handler/user_handler.go` | 新增 `POST /api/v1/users/login` |

---

## Phase 4：多端登录互踢

**新增文件：**

| 文件 | 说明 |
|------|------|
| `internal/repository/token_version_repository.go` | 接口 + Redis 实现：`Incr` / `Get` |
| `internal/repository/mocks/TokenVersionRepository.go` | mockery 自动生成 |

**修改文件：**

| 文件 | 变更 |
|------|------|
| `internal/service/user_service.go` | Login 时 `Incr` 版本号，写入 JWT TokenVersion |
| `internal/middleware/auth.go` | 解析 JWT 后从 Redis 取当前版本，JWT 版本 < Redis 版本 → 401 |
| `internal/middleware/auth_test.go` | 新增 `TestAuthMiddleware_TokenVersionMismatch` |
| `internal/service/user_service_test.go` | 新增 `TestLogin_TokenVersionIncrements` / `TestLogin_TokenVersionIncrError` |
| `cmd/main.go` | 初始化 Redis 客户端，串联 TokenVersionRepository |

**多端互踢流程：**

```
设备A 登录 → Redis INCR user:1:token_version → version=1 → JWT{TokenVersion:1}
设备B 登录 → Redis INCR user:1:token_version → version=2 → JWT{TokenVersion:2}
设备A 请求 → 中间件 GET Redis version=2 > JWT version=1 → 401
```

---

## Phase 5：RBAC 权限控制 + 前端

**后端新增：**

| 文件 | 说明 |
|------|------|
| `db/migrations/000002_add_rbac.up.sql` | 创建 `role` / `permission` / `role_permission` 表，user 加 `role_id`，种子数据 |
| `internal/model/role.go` / `permission.go` / `role_permission.go` | 模型 |
| `internal/repository/permission_repository.go` | 接口 + `HasPermission(userID, path, method)` |
| `internal/repository/mocks/PermissionRepository.go` | mockery 自动生成 |
| `internal/middleware/permission.go` | `PermissionCheck` 中间件：查用户角色权限 → 403 |
| `internal/middleware/permission_test.go` | 3 个测试：有权限200/无权限403/未认证401 |

**前端新增：**

| 文件 | 说明 |
|------|------|
| `frontend/package.json` | Vue3 / Arco Design / Pinia / Axios / Vue Router / Vite |
| `frontend/vite.config.ts` | `@` 别名 + `/api` 代理到 `localhost:8084` |
| `frontend/src/main.ts` | 应用入口 |
| `frontend/src/router/index.ts` | `/login` + `/dashboard` + 导航守卫 |
| `frontend/src/api/index.ts` | Axios 实例 + 拦截器（Token 注入 / 401 跳转） |
| `frontend/src/stores/user.ts` | Pinia：login / logout / Token 持久化 |
| `frontend/src/views/Login.vue` | 登录页 |
| `frontend/src/views/Dashboard.vue` | 工作台 |

**中间件链：**

```
请求 → AuthMiddleware（JWT + 版本号）→ PermissionCheck（路径权限）→ Handler
```

---

## 最终项目结构

```
iam-platform/
├── Makefile
├── config/config.yaml
├── cmd/main.go
├── db/migrations/
│   ├── 000001_init.up.sql / down.sql
│   └── 000002_add_rbac.up.sql / down.sql
├── internal/
│   ├── handler/      user_handler.go
│   ├── middleware/    auth.go (+test) / permission.go (+test)
│   ├── model/        user / role / permission / role_permission
│   ├── repository/    user / token_version / permission + mocks
│   └── service/      user_service.go (+test)
├── pkg/
│   ├── jwt/          jwt.go (+test)
│   └── utils/        bcrypt.go (+test)
├── frontend/         Vue3 + TS + Arco Design
├── docs/
│   ├── ROADMAP.md          后续路线图
│   └── PHASE_1_5_SUMMARY.md  本文件
└── CLAUDE.md          项目规范（AI 行为准则）
```

---

## 测试覆盖

| 包 | 用例数 | 覆盖内容 |
|----|--------|----------|
| pkg/utils | 8 | bcrypt 哈希/校验/边界 |
| pkg/jwt | 7 | JWT 生成/解析/过期/签名/往返 |
| internal/service | 9 | 注册4 + 登录5（含版本自增） |
| internal/middleware/auth | 6 | Token校验5 + 版本互踢1 |
| internal/middleware/perm | 3 | 有权限200/无权限403/未认证401 |
| **合计** | **28** | |
