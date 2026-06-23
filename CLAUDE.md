# 项目核心规范：企业级 IAM 统一认证平台 (Go + MySQL + Vue3)

## 1. 核心 Harness 思想（测试装备与自动化）
- **开发哲学**：采用 “Test-Driven 思维” 与 “Makefile 驱动”。所有核心逻辑必须附带单元测试，测试即文档。
- **必备工具链**：`make` 命令必须包含 `make test`（运行单测）、`make build`（编译）、`make lint`（代码检查）、`make migrate-up`（数据库迁移）。
- **Mock 策略**：Repository 层和第三方 Client 必须定义 Interface，使用 `mockery` 生成 Mock 对象，Service 层单元测试必须基于 Mock，不依赖真实 MySQL。
- **CI/CD 就绪**：代码必须支持通过环境变量覆盖 `config.yaml`（`viper.AutomaticEnv()` + `viper.SetEnvPrefix`），确保后续部署到 Linux 服务器无障碍。

## 2. 后端技术栈 (Go)
- **语言版本**：Go 1.22+。
- **Web 框架**：Gin。
- **数据库 (MySQL)**：GORM (使用 `gorm.io/driver/mysql`)。**禁止**使用 `gorm.AutoMigrate` 做生产迁移，统一使用 `golang-migrate` 维护 `db/migrations` 文件夹下的 SQL 脚本。
- **缓存**：Redis (使用 `go-redis/redis`)，用于存储 JWT 版本号、限流计数、验证码。
- **配置**：Viper (支持 YAML + 环境变量覆盖)。
- **日志**：使用 Zap 输出结构化 JSON 日志，包含 `trace_id`（`pkg/log/logger.go` + `middleware/zap_logger.go`）。Logger 未初始化时自动回退 `log.Printf`。
- **依赖注入**：使用 `google/wire` 生成依赖注入代码（`cmd/wire.go` → `cmd/wire_gen.go`），主函数通过 `InitComponents()` 获取所有组件。

## 3. 前端技术栈 (Web UI)
- **框架**：Vue 3 + TypeScript + Vite。
- **UI 组件库**：Arco Design Vue (字节出品，企业级)。
- **状态管理**：Pinia (持久化存储 Token、菜单权限树)。
- **HTTP 请求**：Axios (封装拦截器，自动注入 `Authorization`，处理 401 跳转)。

## 4. 数据库设计规范 (MySQL)
- 字符集统一 `utf8mb4`，排序规则 `utf8mb4_unicode_ci`。
- 表名采用单数形式（如 `user`, `role`, `permission`）。
- 必备审计字段：`created_at` (DATETIME), `updated_at` (DATETIME), `deleted_at` (DATETIME NULL, 软删除)。
- 禁止使用`select *`等慢sql语句。

## 5. 代码分层架构 (必须严格遵循)
- **Handler (Controller)**：只负责绑定参数（ShouldBindJSON），调用 Service，返回 HTTP 状态码。**不包含**业务逻辑。
- **Service (业务逻辑层)**：**必须定义 Interface**（如 `UserService`），实现结构体私有（小写）。包含权限判断、事务逻辑。
- **Repository (数据访问层)**：**必须定义 Interface**，只负责 GORM 的 CRUD 操作。
- **Model (实体)**：GORM 的 Tag 定义，与数据库表一一映射。
- **Middleware (中间件)**：JWT 解析、租户注入、限流、日志链路。

## 6. 关键业务难点实现方案
- **JWT 增强**：Claims 必须包含 `UserID`, `TenantID`, `DeviceID`, `TokenVersion`。
- **多端登录互踢**：利用 Redis 维护 `user:{user_id}:token_version`。每次请求中间件校验 JWT 中的 `TokenVersion` 是否与 Redis 一致，不一致则强制过期。
- **动态数据权限 (AOP 思想)**：利用 GORM 的 `Scopes` 功能，在 Repository 层根据当前用户的角色拼接 `tenant_id` 和 `dept_id` 过滤条件。

## 7. 单元测试硬性要求 (Claude 必须执行)
当生成或修改以下模块时，**必须同步生成 `_test.go` 单元测试文件**：
- **Utils 工具类**（如 JWT 加解密、密码 Bcrypt 哈希）：必须覆盖边界条件（如空字符串、过期 Token）。
- **Service 业务逻辑**：使用 `testify/assert` 和 `mockery` 生成的 Mock 对象。至少包含“正常流程”和“异常流程”（如数据库报错、用户不存在）两个测试用例。
- **Middleware 中间件**：使用 `net/http/httptest` 模拟请求测试拦截逻辑。

## 8. 交付规则
**严禁一次性生成整个项目或整个 Phase。** 必须将复杂任务拆分为可独立交付的小步，每步完成后停下来审视代码完整性。当前项目阶段划分见 `docs/ROADMAP.md`。

每完成一个步骤，必须：
1. 运行 `go test ./...` 确认历史用例零回归。
2. 运行 `go build ./...` 确认编译无错误。
3. 确认变更范围内的 `_test.go` 已同步生成且覆盖正常/异常路径。
4. 不可盲目自信，主动提出发现的问题。

## 9. 交付前强制自检清单（必须逐项执行）

**每轮交付前，必须对照以下清单逐项检查，不可跳过。**
已知违规模式详见 `memory/bug_patterns.md`。

### 9.1 编译与测试
- [ ] `go build ./...` 零错误
- [ ] `go test ./...` 全部通过，零 panic，零 timeout
- [ ] 前端变更时 `vue-tsc --noEmit` 和 `vite build` 通过

### 9.2 接口变更级联
- [ ] Repository 接口新增/修改方法后，执行 `mockery --name=XXX` 重新生成 Mock
- [ ] 构造函数签名变更后，搜索所有 `NewXXX(` 调用点确认参数匹配
- [ ] 删除或修改方法签名后，搜索项目中所有引用点

### 9.3 日志完整性
- [ ] 每个 Handler 方法的所有 500 分支前是否有 `pkgl.Error("方法名", zap.Error(err))`？
- [ ] 是否有 Handler 完全无日志？——用 `grep -L "pkgl.Error" internal/handler/*.go` 检测
- [ ] DB 错误是否被正确返回 500（而非 400）？是否用 `errors.Is` 区分业务错误和系统错误？

### 9.4 中间件作用域
- [ ] 新中间件是全局应用（`r.Use`）还是路由级应用（`group.Use`）？
- [ ] 是否会误拦截 `/health`、`/metrics` 等非业务路由？
- [ ] 全局中间件是否对每个请求都做 DB 查询？是否可改为路由级？

### 9.5 前端-后端对齐
- [ ] 每个前端 API 调用是否在后端有对应的 Handler + 路由注册？
- [ ] 种子数据中的 menu `path` 是否与前端 router 的 `path` 完全一致（含单复数）？
- [ ] 新 API 是否在种子数据中有对应的 GET/POST/PUT/DELETE 权限记录？

### 9.6 零值/指针语义
- [ ] `uint64`/`int` 字段零值是否有业务含义（如 role_id=0 表示未分配）？是否应改用 `*uint64` 指针？
- [ ] `_` 丢弃的 error 是否真的可忽略？时间解析、JSON 解析等关键路径不可丢弃。

### 9.7 迁移脚本
- [ ] 是修改已有迁移还是新建迁移？存量数据库是否需要 UPDATE 来修正已有数据？
- [ ] 新建迁移能否被 `golang-migrate` 正确 up/down？

### 9.8 原子性与关闭顺序
- [ ] Redis 多命令操作是否用 Lua 脚本保证原子性？
- [ ] 优雅关闭顺序：Shutdown → close(channel) → WaitGroup.Wait() → Close(DB/Redis)
- [ ] 依赖初始化顺序：被调用方在使用前是否已初始化（如 Logger）？

### 9.9 交付后自我审计
完成以上检查后，必须主动报告以下内容（不可等用户来问）：
1. 本轮变更的文件清单（新增 + 修改）
2. 测试数量变化（本轮新增几个，累计几个）
3. 主动发现的已知限制或潜在风险（≥ 1 条，若无则写"本轮未发现明显风险"）