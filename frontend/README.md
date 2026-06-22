# IAM Platform Frontend

基于 Vue 3 + TypeScript + Vite + Arco Design 构建的统一身份认证平台前端。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建**: Vite 5
- **UI 组件**: Arco Design Vue
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP**: Axios

## 项目结构

```
src/
  api/          Axios 实例与接口封装
  router/       路由配置（含导航守卫）
  stores/       Pinia 状态（用户 Token 持久化）
  styles/       全局样式
  views/        页面组件
```

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:3000）
npm run dev

# 构建生产版本
npm run build
```

## 后端代理

开发环境下 `/api` 请求自动代理到 `http://localhost:8084`（见 `vite.config.ts`）。生产环境请通过 Nginx 反向代理。

## 功能

- 登录 / 注册页面
- JWT Token 自动注入 Axios 请求头
- 401 自动跳转登录页
- Pinia 持久化 Token 至 localStorage
- 路由导航守卫（未登录拦截）
