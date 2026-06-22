# IAM Platform 开发路线图

> 版本：v1.0 | 日期：2026-06-** | 状态：Phase 1-5 已完成

---

## 一、现状总览

Phase 1-5 构建了项目骨架与核心鉴权链路，具备以下能力：

| 模块 | 已有能力 | 核心代码 |
|---|---|---|
| 注册登录 | Bcrypt 加密 + JWT 签发 + 防重复注册 | `user_handler.go` / `user_service.go` |
| 鉴权中间件 | Bearer Token 解析 + Claims 注入上下文 | `middleware/auth.go` |
| 多端互踢 | Redis 维护 `user:{id}:token_version`，低版本 401 | `token_version_repository.go` + `auth.go` |
| 接口权限 | role → role_permission → permission 三级联查 | `middleware/permission.go` + `permission_repository.go` |
| 前端 | Vue3 + Arco Design 登录页 + Dashboard 壳 | `frontend/` |
| 测试 | 28 个单元测试（jwt / bcrypt / service / middleware） | `*_test.go` |

**数据库表：** `user` / `role` / `permission` / `role_permission`（含种子数据：admin、user 两个角色 + profile 权限）。

---

## 二、当前存在的问题

以下问题经全量审计确认，按影响程度分三级。

### P0 —— 阻塞性问题（应立即修复）

| # | 问题 | 文件 | 影响 |
|---|---------|------|------|
| P0-1 | `role` 表迁移缺 `INDEX idx_deleted_at`，但模型有 `gorm:"index"` | `000002_add_rbac.up.sql` | GORM 软删查询无索引，数据量增长后查询退化 |
| P0-2 | 页面刷新后 Dashboard 角色显示为"普通用户"（roleId=0） | `frontend/src/stores/user.ts` | `fetchProfile()` 仅在 `login()` 时调用，F5 刷新后不触发 |
| P0-3 | favicon 404 | `frontend/public/` | `index.html` 引用 `/vite.svg`，实际文件不存在 |

### P1 —— 架构性缺失（应在下一轮补齐）

| # | 问题 | 说明 |
|---|---------|------|
| P1-1 | 无优雅关闭 | 服务无 `signal.Notify` / `http.Server.Shutdown`，进程直接被杀 |
| P1-2 | 无 CORS 中间件 | 开发时 Vite 代理规避了问题，生产部署跨域会报错 |
| P1-3 | 环境变量覆盖未实现 | `viper.AutomaticEnv()` 未调用，违反 CLAUDE.md 规范 |
| P1-4 | `/profile` 内联在 main.go | 违反分层架构，应移入 `handler/user_handler.go` |
| P1-5 | 无角色分配 API | 用户注册后 `role_id=0`，没有接口可以分配角色 |
| P1-6 | 无权限 CRUD API | 权限只能通过 SQL INSERT，没有管理界面 |
| P1-7 | 无登录限流 | `/api/v1/users/login` 无频率限制，可被暴力破解 |
| P1-8 | Claims 中 TenantID/DeviceID 始终为 0/空 | 多租户和 DeviceID 字段已定义但从未赋值 |

### P2 —— 可延后优化

| # | 问题 | 说明 |
|---|---------|------|
| P2-1 | 无注册页面 | 注册只能通过 curl，前端只有登录页 |
| P2-2 | 前端打包 945KB，未做代码分割 | Arco Design 全量引入 |
| P2-3 | 未实现 Zap 结构化日志 | 仍用 `log.Fatalf` / `fmt.Printf` |
| P2-4 | 未实现 Google Wire 依赖注入 | 全部手动在 main.go 中 `NewXxx()` |
| P2-5 | go.mod 模块路径指向不存在的 GitHub 仓库 | `github.com/cyancyan2020/iam-platform` |
| P2-6 | Handler / Repository 层无单元测试 | 只有 Service 和 Middleware 有测试 |

---

## 三、后续路线图

```
Phase 6   ████████ 修复 P0-1/2/3 + P1-1/2/3/4（地基加固）
Phase 7   ████████ 角色分配 API + 权限 CRUD API
Phase 8   ████████ 前端管理界面：用户管理 + 角色管理 + 权限树
Phase 9   ████████ 数据权限（GORM Scopes + DataScopeMiddleware）
Phase 10  ████████ 操作日志（日志中间件 + 异步写入 + 前端展示）
Phase 11  ████████ 登录限流 + 优雅关闭 + CORS + Zap 日志 + Wire DI
```

---

## 四、各阶段详细方案

### Phase 6：地基加固

**目标：** 修复 P0 全部问题 + P1 中阻塞性问题，让项目进入"生产就绪"状态。

**交付物：**

| 任务 | 涉及文件 | 测试要求 |
|------|----------|----------|
| 迁移增加 `role.deleted_at` 索引 | `000002_add_rbac.up.sql` | SQL 语法校验通过 |
| Dashboard 挂载时调 `fetchProfile()` | `frontend/src/stores/user.ts` + `Dashboard.vue` | F5 刷新后角色正确显示 |
| 生成 SVG favicon | `frontend/public/vite.svg` | 浏览器 tab 图标正常 |
| 实现优雅关闭 | `cmd/main.go` | `ctrl+c` 后日志输出 shutdown 信息 |
| 添加 CORS 中间件 | `internal/middleware/cors.go` | 跨域 OPTIONS 预检返回 204 |
| 实现 `viper.AutomaticEnv()` | `cmd/main.go` | `DB_DSN=xxx make run` 可覆盖 yaml |
| `/profile` 移入 handler | `handler/user_handler.go` + `cmd/main.go` | 接口行为不变 |

**验收标准：**
- `make test` 28 个用例全部通过（不低于当前）
- 前端 `npm run build` 成功
- `ctrl+c` 终止服务时控制台输出 "server shutting down..."
- `curl -X OPTIONS http://localhost:8084/api/v1/users/login -H "Origin: http://localhost:3000"` 返回 204

---

### Phase 7：角色分配 + 权限 CRUD API

**目标：** 补齐管理员对角色和权限的管理能力。

**后端接口：**

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/api/v1/users/:id/role` | 为用户分配角色 | 管理员 |
| GET | `/api/v1/roles` | 角色列表 | 管理员 |
| POST | `/api/v1/roles` | 创建角色 | 管理员 |
| PUT | `/api/v1/roles/:id` | 编辑角色 | 管理员 |
| DELETE | `/api/v1/roles/:id` | 删除角色 | 管理员 |
| GET | `/api/v1/permissions` | 权限树列表 | 管理员 |
| POST | `/api/v1/permissions` | 创建权限 | 管理员 |
| PUT | `/api/v1/permissions/:id` | 编辑权限 | 管理员 |
| DELETE | `/api/v1/permissions/:id` | 删除权限 | 管理员 |
| POST | `/api/v1/roles/:id/permissions` | 为角色批量分配权限 | 管理员 |

**交付物：**

| 任务 | 涉及文件 |
|------|----------|
| RoleRepository 接口 + 实现 | `internal/repository/role_repository.go` |
| PermissionRepository 扩展（CRUD） | `internal/repository/permission_repository.go` |
| RolePermissionRepository | `internal/repository/role_permission_repository.go` |
| RoleService 接口 + 实现 | `internal/service/role_service.go` |
| RoleHandler | `internal/handler/role_handler.go` |
| PermissionHandler | `internal/handler/permission_handler.go` |
| mockery 生成所有新 Repository 的 Mock | `internal/repository/mocks/` |
| Service 层单元测试 | `internal/service/role_service_test.go` |
| 种子数据补充 | `db/migrations/000003_seed_permissions.up.sql` |

**测试要求：**
- 每个 Service 至少包含"正常流程"和"异常流程"两个测试用例
- Service 层测试基于 mockery Mock，不依赖真实 MySQL
- 角色 CRUD 测试覆盖：创建/列表/更新/删除 正常 + 角色已存在/角色不存在

---

### Phase 8：前端管理界面

**目标：** 提供可视化的管理后台，包含用户管理、角色管理、权限树配置三块。

**页面清单：**

#### 8a. 布局组件 `Layout.vue`

```
┌─────────────────────────────────────┐
│  IAM Platform          [admin ▾]   │ ← 顶栏 56px
├──────────┬──────────────────────────┤
│ 系统管理 │                          │
│   用户管理│       <router-view />    │ ← 内容区
│   角色管理│                          │
│   权限管理│                          │
└──────────┴──────────────────────────┘
          ↑ 侧边栏 220px
```

- 侧边栏菜单根据后端返回的权限树动态渲染
- 顶栏显示当前用户名，下拉菜单含"退出登录"

#### 8b. 用户管理 `views/system/Users.vue`

```
┌──────────────────────────────────────────────┐
│  用户管理                        [+ 新增用户] │
│  [用户名        ] [状态▼] [查询] [重置]      │
│  ┌──────┬────────┬────┬──────┬──────┬──────┐│
│  │ ID   │ 用户名  │昵称│ 角色  │ 状态 │ 操作 ││
│  │ 1    │ admin  │... │管理员 │ 启用 │编辑删││
│  └──────┴────────┴────┴──────┴──────┴──────┘│
│           [< 1 2 3 ... 15 >]                │
└──────────────────────────────────────────────┘
```

- 表格基于 Arco `a-table`，分页基于 `a-pagination`
- "新增/编辑"弹窗（`a-modal`）：用户名、密码、昵称、角色下拉
- 删除需二次确认（`a-popconfirm`）

#### 8c. 角色管理 `views/system/Roles.vue`

```
┌──────────────────────────────────────────────┐
│  角色管理                        [+ 新增角色] │
│  ┌──────────┬──────┬──────────┬──────┐       │
│  │ 名称     │ 编码  │ 数据范围  │ 操作 │       │
│  │ 管理员   │ admin│ 全部     │编辑删│       │
│  │ 华南经理 │ south│ 本部门   │编辑删│       │
│  └──────────┴──────┴──────────┴──────┘       │
└──────────────────────────────────────────────┘
```

- 点击"编辑"弹出抽屉（`a-drawer`，宽度 480px）
- 抽屉内含三个区域：基本信息表单 + 权限树（`a-tree` 多选）+ 数据范围选择器

#### 8d. 权限管理 `views/system/Permissions.vue`

```
┌──────────────────────────────────────────────┐
│  权限管理                        [+ 新增权限] │
│                                              │
│  📁 系统管理                                  │
│    📄 用户管理   GET /api/v1/users   编辑 删除 │
│      🔘 新增     POST .../users      编辑 删除 │
│      🔘 编辑     PUT .../users/:id   编辑 删除 │
│  📁 角色管理     GET /api/v1/roles   编辑 删除 │
│                                              │
└──────────────────────────────────────────────┘
```

- `a-tree` 展示完整权限树
- 节点类型：`menu`（📁 菜单组）/ `page`（📄 页面）/ `button`（🔘 按钮）
- 右键或悬停显示操作按钮
- 新增/编辑弹窗：名称、路径、方法、父级选择器

**交付物：**

| 任务 | 涉及文件 |
|------|----------|
| Layout 组件 | `src/views/Layout.vue` |
| 用户管理页 | `src/views/system/Users.vue` |
| 角色管理页 | `src/views/system/Roles.vue` |
| 权限管理页 | `src/views/system/Permissions.vue` |
| API 封装 | `src/api/system.ts`（用户/角色/权限接口） |
| 路由配置 | `src/router/index.ts`（嵌套路由 + 懒加载） |
| store 扩展 | `src/stores/user.ts`（新增角色/权限列表状态） |
| 动态菜单渲染 | Layout.vue `onMounted` 根据权限树生成侧边栏 |

**测试要求：**
- `npm run build` 无报错
- `vue-tsc --noEmit` 类型检查通过
- 登录 → 管理员看到全部菜单 → 普通用户看到受限菜单
- F5 刷新后菜单不丢失

---

### Phase 9：数据权限

**目标：** 同一接口不同角色返回不同范围的数据。

**数据流设计：**

```
AuthMiddleware → PermissionCheck → DataScopeMiddleware
                                       │
                          ┌────────────┴────────────┐
                          │ 查 role → 取 data_scope │
                          │ All / Dept(3,4,5) / Self│
                          └────────────┬────────────┘
                                       │ c.Set("dataScope", scope)
                                       ▼
                          Repository.Scopes(scope)
                            → GORM 自动拼接 WHERE
```

**需要改造的层：**

| 层 | 改动 |
|------|------|
| Model | `role` 表加 `data_scope` 字段（`all` / `dept` / `self`） |
| Middleware | 新增 `DataScopeMiddleware`，查角色表构造 Scope 对象 |
| Repository | 每个 Repository 方法接收 `dataScope`，用 GORM Scopes 拼 WHERE |
| Handler | 无改动（对上层透明） |
| 数据库迁移 | `000004_add_data_scope.up.sql` |

**DataScope 定义：**

```go
// internal/middleware/data_scope.go
type DataScope struct {
    All    bool     // 管理员 → 不限制
    DeptIDs []uint64 // 部门经理 → WHERE dept_id IN (...)
    Self   bool     // 普通用户 → WHERE user_id = currentUserID
}
```

**Repository 层示例（以订单为例）：**

```go
// 假设将来有订单模块
func (r *orderRepo) List(ctx context.Context) ([]Order, error) {
    scope := ctx.Value("dataScope").(DataScope)
    return r.db.WithContext(ctx).Scopes(applyScope(scope)).Find(&orders)
}

func applyScope(s DataScope) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if s.All { return db }
        if len(s.DeptIDs) > 0 { return db.Where("dept_id IN ?", s.DeptIDs) }
        if s.Self { return db.Where("user_id = ?", s.UserID) }
        return db.Where("1 = 0") // 默认不可见
    }
}
```

**交付物：**

| 任务 | 涉及文件 |
|------|----------|
| role 表加 data_scope 字段 | `db/migrations/000004_add_data_scope.up.sql` |
| DataScope 中间件 | `internal/middleware/data_scope.go` |
| 中间件单元测试 | `internal/middleware/data_scope_test.go` |
| main.go 注册 | `cmd/main.go` |
| Role 模型更新 | `internal/model/role.go` |

**验收标准：**
- 管理员查 `/api/v1/users` 返回所有用户
- 部门经理查 `/api/v1/users` 返回本部门
- 普通用户查 `/api/v1/users` 只能看到自己
- 3 个中间件测试过（All → 200 / Dept → 200 过滤 / Self → 200 过滤）
- 回归：已有 28+ 测试全部通过

---

### Phase 10：操作日志

**目标：** 记录所有敏感操作（登录、CRUD、权限变更），可查询可导出。

**后端：**

```sql
CREATE TABLE operation_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    username VARCHAR(64) NOT NULL,
    method VARCHAR(10) NOT NULL,       -- GET/POST/PUT/DELETE
    path VARCHAR(256) NOT NULL,        -- /api/v1/users
    ip VARCHAR(45) NOT NULL,
    user_agent VARCHAR(512),
    status_code INT NOT NULL,          -- HTTP 状态码
    duration_ms INT NOT NULL,          -- 请求耗时
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);
```

核心是**异步写入**——在中间件中用 goroutine + channel 写入，不阻塞请求响应：

```go
// internal/middleware/operation_log.go
func OperationLog(logChan chan<- model.OperationLog) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        // 异步写入，不阻塞响应
        go func() {
            logChan <- model.OperationLog{
                UserID:     claims.UserID,
                Path:       c.Request.URL.Path,
                Method:     c.Request.Method,
                StatusCode: c.Writer.Status(),
                DurationMs: int(time.Since(start).Milliseconds()),
            }
        }()
    }
}
```

然后在 main.go 启动一个 consumer goroutine 批量写入 MySQL。

**前端：**

只读表格 + 时间范围筛选（`a-range-picker`）+ 导出 CSV 按钮。

**交付物：**

| 任务 | 涉及文件 |
|------|----------|
| 迁移脚本 | `db/migrations/000005_operation_log.up.sql` |
| Model | `internal/model/operation_log.go` |
| 日志中间件 + consumer | `internal/middleware/operation_log.go` |
| 日志查询接口 | `internal/handler/log_handler.go` |
| 前端日志页面 | `src/views/system/Logs.vue` |

**验收标准：**
- 任意登录/注册操作后，`operation_log` 表有新记录（<1s 延迟）
- `GET /api/v1/logs?start=...&end=...` 返回分页结果
- 前端日志页时间筛选 + CSV 导出正常

---

### Phase 11：生产加固

**目标：** 补齐登录限流、Zap 日志、Wire 依赖注入、CORS、优雅关闭（Phase 6 未完成则在此完成）。

**交付物：**

| 任务 | 说明 |
|------|------|
| 登录限流 | Redis 滑动窗口，`/api/v1/users/login` 每分钟最多 5 次，超限 429 |
| Zap 日志 | 替换 `log.Printf` / `fmt.Printf`，输出 JSON + trace_id |
| Wire DI | 编写 `wire.go`，`wire gen` 生成 `wire_gen.go`，替代 main.go 中的手动依赖组装 |
| CORS | `internal/middleware/cors.go`，允许配置的前端域名 |
| 优雅关闭 | signal 监听 + `http.Server.Shutdown(ctx)` |

---

## 五、交付规则（适用于所有后续 Phase）

1. **未修复 P0 问题前不进新功能。**
2. **每个 Phase 必须有单元测试增量。** 新增代码的 `*_test.go` 不超过 Phase 内文件总数的 80%。
3. **Service 层测试始终基于 mockery Mock，不依赖真实 MySQL / Redis。**
4. **Middleware 测试全部基于 `net/http/httptest`。**
5. **每个 Phase 完成后运行 `make test` 确认全部历史用例通过，零回归。**
6. **前端每个 Phase 完成后运行 `vue-tsc --noEmit && vite build` 确认编译零错误。**
7. **禁止一次性生成整个 Phase。** 复杂 Phase 应再拆分为 2-3 轮，每轮交付后停下来审视。
8. **数据库变更统一通过 `golang-migrate` 迁移脚本，禁止手动改表。**

---

## 六、测试目标汇总

| Phase | 最低用例增量 | 累计用例数（约） |
|-------|-------------|-----------------|
| Phase 6 | 无强制增量（加固为主） | 28 |
| Phase 7 | +12（Service CRUD × 3 接口 × 2 场景 × 2 模块） | 40 |
| Phase 8 | 前端编译通过即可 | 40 |
| Phase 9 | +4（DataScope 中间件：All/Dept/Self/无登录） | 44 |
| Phase 10 | +6（日志中间件 2 + 查询 service 4） | 50 |
| Phase 11 | +4（限流中间件 2 + CORS 2） | 54 |
