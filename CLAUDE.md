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
- **日志**：规划使用 Zap 输出结构化 JSON 日志并包含 `trace_id`（当前为 `log.Printf`，待 Phase 11 实施）。
- **依赖注入**：规划使用 `google/wire` 生成依赖注入代码（当前手动在 `cmd/main.go` 中组装，待 Phase 11 实施）。

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
1. 运行 `make test` 确认历史用例零回归。
2. 运行 `go build` 确认编译无错误。
3. 确认变更范围内的 `_test.go` 已同步生成且覆盖正常/异常路径。
4. 不可盲目自信，主动提出发现的问题。