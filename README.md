# API 网关（通用中转平台）

这是一个公司内部 API 中转网关，支持：

- LLM API（OpenAI / Claude / Gemini）
- 第三方 API（Google / YouTube 等）
- 任意 HTTP API（通过上游配置新增）

并提供：

- 员工 API Key 分发
- 请求日志记录
- 用量统计（按用户、按上游）
- Web 管理面板（用户、上游、Key、日志）

---

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：React + Ant Design + Vite
- 部署：Docker Compose

---

## 快速启动（本地）

### 1) 准备环境变量

```bash
cp .env.example .env
```

建议至少修改：

- `JWT_SECRET`
- `DEFAULT_ADMIN_PASS`

### 2) 启动后端

```bash
go mod tidy
go run ./cmd/server
```

服务地址：`http://localhost:8080`

健康检查：`GET /api/health`

### 3) 启动前端（开发）

```bash
cd web
npm install
npm run dev
```

前端开发地址：`http://localhost:5173`

---

## Docker 部署

默认映射到宿主机 `8081` 端口（容器内部仍是 8080）。

```bash
chmod +x deploy.sh
./deploy.sh
```

或手动：

```bash
docker compose up -d --build
```

如果你要改映射端口，修改 `.env` 的 `HOST_PORT` 即可。

---

## 路由说明

### 管理端 API

- 登录：`POST /api/auth/login`
- 用户管理：`/api/admin/users`
- 上游管理：`/api/admin/upstreams`
- API Key 管理：`/api/admin/api-keys`
- 日志：`/api/admin/logs`
- 统计：`/api/admin/stats/overview`、`/api/admin/stats/daily`

### 代理转发

```text
/proxy/{api_name}/{path...}
```

示例：

```text
/proxy/openai/v1/chat/completions
/proxy/youtube/youtube/v3/search
```

请求头必须带员工 API Key：

```text
Authorization: Bearer sk-xxxx
```

---

## 项目结构（核心）

- `cmd/server/main.go`：应用入口
- `internal/proxy/*`：反向代理核心
- `internal/middleware/*`：JWT/API Key/限流/CORS
- `internal/handler/*`：管理端接口
- `internal/service/*`：业务逻辑
- `internal/database/*`：SQLite 初始化与 seed
- `web/*`：管理台前端

---

## 当前实现状态

已完成：

- 后端核心代理链路
- API Key 鉴权
- 管理端基础 CRUD（用户、上游、Key、日志、统计）
- Docker 部署骨架
- 前端管理台基础页面

后续可增强：

- 上游密钥加密存储
- 更细粒度的配额与速率策略
- 图表和导出功能增强
- 审计与告警能力
