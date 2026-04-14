# API 网关（通用中转平台）

这是一个公司内部 API 中转网关，支持：

- LLM API（OpenAI / Claude / Gemini）
- 第三方 API（Google / YouTube / Product Hunt 等）
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
/proxy/producthunt/v2/api/graphql
/proxy/doubao_embedding/api/v3/embeddings/multimodal
```

请求头必须带员工 API Key：

```text
Authorization: Bearer sk-xxxx
```

---

## Product Hunt / Google / 豆包接入

### 1) Product Hunt GraphQL

- API 端点：`https://api.producthunt.com/v2/api/graphql`
- 鉴权：`Authorization: Bearer <token>`

在 `.env` 填写：

```bash
PRODUCT_HUNT_TOKEN=ph_xxx
```

重启服务后会自动生成 `producthunt` 上游。

调用示例：

```bash
curl http://localhost:8081/proxy/producthunt/v2/api/graphql \
  -H "Authorization: Bearer sk-你的员工key" \
  -H "Content-Type: application/json" \
  -d '{"query":"query { __typename }"}'
```

### 2) Google API Key（含 YouTube）

在 `.env` 填写：

```bash
GOOGLE_API_KEY=AIza...
```

重启后自动生成 `google` 上游（query `key=`）。

调用示例（YouTube）：

```bash
curl "http://localhost:8081/proxy/google/youtube/v3/search?part=snippet&q=ai&type=video" \
  -H "Authorization: Bearer sk-你的员工key"
```

### 3) 豆包 Embedding

在 `.env` 配置：

```bash
DOUBAO_EMBEDDING_API_BASE=https://ark.cn-beijing.volces.com
DOUBAO_EMBEDDING_DIMENSIONS=2048
```

重启后自动生成 `doubao_embedding` 上游。

调用示例：

```bash
curl http://localhost:8081/proxy/doubao_embedding/api/v3/embeddings/multimodal \
  -H "Authorization: Bearer sk-你的员工key" \
  -H "Content-Type: application/json" \
  -d '{"input":["hello"],"dimensions":2048}'
```

> `DOUBAO_EMBEDDING_DIMENSIONS` 是推荐值，实际请求中你仍可显式传 `dimensions`。

---

## 前端上游创建“点 OK 没反应”修复

已修复：

- 创建/更新/删除增加错误提示（会显示后端返回 error）
- 增加 loading 与提交状态
- 新增快捷模板按钮：Product Hunt / Google / 豆包 Embedding

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
- Product Hunt / Google / 豆包接入预置

后续可增强：

- 上游密钥加密存储
- 更细粒度的配额与速率策略
- 图表和导出功能增强
- 审计与告警能力
