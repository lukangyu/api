# API 网关（通用中转平台）

公司内部 API 中转网关，将各类第三方 API 统一收口，通过员工 API Key 分发访问权限。

支持转发：LLM API（OpenAI / Claude / Gemini）、第三方 API（Google / YouTube / Product Hunt）、任意 HTTP API。

提供：员工 API Key 分发、请求日志、用量统计、Web 管理面板。

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：React + Ant Design + Vite
- 部署：Docker Compose

---

## 快速启动

### 1. 准备环境变量

```bash
cp .env.example .env
```

**必须修改**：`JWT_SECRET`、`DEFAULT_ADMIN_PASS`

### 2. 启动后端

```bash
go mod tidy
go run ./cmd/server
```

服务地址：`http://localhost:8080`，健康检查：`GET /api/health`

### 3. 启动前端（开发模式）

```bash
cd web && npm install && npm run dev
```

前端开发地址：`http://localhost:5173`

### 4. Docker 部署

```bash
docker compose up -d --build
```

默认映射到宿主机 `8081` 端口，可在 `.env` 中修改 `HOST_PORT`。

---

## 核心概念

### 转发流程

```
客户端 → /proxy/{上游名称}/{路径...} → 网关鉴权 + 限流 → 转发至上游 API → 返回响应
```

客户端请求必须携带员工 API Key：

```
Authorization: Bearer sk-xxxx
```

网关根据 URL 中的**上游名称**查找对应的上游配置，自动完成：URL 改写、鉴权注入、额外请求头注入，然后转发请求。

---

## 上游注册

所有上游通过管理后台手动创建：**上游 API 管理 → 新建上游**。

### 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 上游唯一标识，用在 URL 路径中（`/proxy/{name}/...`）。只允许字母、数字、下划线、连字符，最长 64 位。例如：`openai`、`google`、`producthunt` |
| `display_name` | 是 | 显示名称，仅用于管理界面展示。例如：`OpenAI API` |
| `base_url` | 是 | 上游 API 的基础地址，必须以 `http://` 或 `https://` 开头。例如：`https://api.openai.com` |
| `auth_type` | 是 | 鉴权方式，见下方详细说明。可选值：`none`、`bearer`、`header`、`query` |
| `auth_key` | 条件必填 | 鉴权键名。`auth_type` 为 `header` 或 `query` 时必填。例如：`x-api-key`、`key` |
| `auth_value` | 否 | 鉴权值（上游密钥/Token）。例如：`sk-proj-xxx`、`AIza-xxx` |
| `timeout_seconds` | 否 | 请求超时时间（秒），默认 `120`。超时后网关返回 502 |
| `strip_prefix` | 否 | 是否去掉 `/proxy/{name}` 前缀，默认 `true`。绝大多数场景保持默认即可 |
| `extra_headers` | 否 | 额外请求头，JSON 格式。例如：`{"X-Custom": "value"}` |
| `is_active` | 否 | 是否启用，默认 `true`。禁用后该上游不可被访问 |
| `description` | 否 | 备注说明 |

### 四种鉴权方式

#### `none` — 不注入鉴权

网关不做任何鉴权处理，由调用方在请求中自行携带鉴权信息。

适用场景：上游本身不需要鉴权，或鉴权信息由调用方动态传入。

#### `bearer` — Bearer Token

网关自动注入请求头：

```
Authorization: Bearer {auth_value}
```

适用场景：OpenAI、Claude、Product Hunt 等使用 Bearer Token 鉴权的 API。

#### `header` — 自定义请求头

网关自动注入请求头：

```
{auth_key}: {auth_value}
```

适用场景：使用 `x-api-key`、`api-key` 等自定义头鉴权的 API。

#### `query` — URL 查询参数

网关自动在 URL 中追加查询参数：

```
?{auth_key}={auth_value}
```

适用场景：Google API 等使用 `?key=xxx` 鉴权的 API。

---

## 上游注册示例

### OpenAI

| 字段 | 值 |
|------|----|
| name | `openai` |
| display_name | `OpenAI API` |
| base_url | `https://api.openai.com` |
| auth_type | `bearer` |
| auth_value | `sk-proj-xxx` |

调用：

```bash
curl http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-你的员工key" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}'
```

### Google API（含 YouTube）

| 字段 | 值 |
|------|----|
| name | `google` |
| display_name | `Google API` |
| base_url | `https://www.googleapis.com` |
| auth_type | `query` |
| auth_key | `key` |
| auth_value | `AIza-xxx` |

调用：

```bash
curl "http://localhost:8080/proxy/google/youtube/v3/search?part=snippet&q=ai" \
  -H "Authorization: Bearer sk-你的员工key"
```

实际转发到：`https://www.googleapis.com/youtube/v3/search?part=snippet&q=ai&key=AIza-xxx`

### Product Hunt GraphQL

| 字段 | 值 |
|------|----|
| name | `producthunt` |
| display_name | `Product Hunt GraphQL` |
| base_url | `https://api.producthunt.com` |
| auth_type | `bearer` |
| auth_value | `ph_xxx` |

调用：

```bash
curl http://localhost:8080/proxy/producthunt/v2/api/graphql \
  -H "Authorization: Bearer sk-你的员工key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ posts(first:3) { edges { node { name tagline } } } }"}'
```

### 豆包 Embedding

| 字段 | 值 |
|------|----|
| name | `doubao_embedding` |
| display_name | `Doubao Embedding` |
| base_url | `https://ark.cn-beijing.volces.com` |
| auth_type | `none` |
| description | 调用方需自行在请求头中携带鉴权 |

调用：

```bash
curl http://localhost:8080/proxy/doubao_embedding/api/v3/embeddings/multimodal \
  -H "Authorization: Bearer sk-你的员工key" \
  -H "Content-Type: application/json" \
  -d '{"input":["hello"],"dimensions":2048}'
```

---

## API 路由一览

### 管理端（JWT 鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 管理员登录 |
| GET/POST/PUT/DELETE | `/api/admin/users` | 用户管理 |
| GET/POST/PUT/DELETE | `/api/admin/upstreams` | 上游管理 |
| GET/POST/PUT/DELETE | `/api/admin/api-keys` | API Key 管理 |
| GET | `/api/admin/logs` | 请求日志 |
| GET | `/api/admin/stats/overview` | 统计概览 |
| GET | `/api/admin/stats/daily` | 每日统计 |

### 代理转发（API Key 鉴权）

```
ANY /proxy/:api_name/*path
```

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 服务端口 |
| `HOST_PORT` | `8081` | Docker 宿主机映射端口 |
| `DB_PATH` | `./data/gateway.db` | SQLite 文件路径 |
| `JWT_SECRET` | `change_me_jwt_secret` | JWT 签名密钥，**必须修改** |
| `JWT_EXPIRE_HOURS` | `24` | JWT 过期时间（小时） |
| `DEFAULT_ADMIN_USER` | `admin` | 默认管理员用户名 |
| `DEFAULT_ADMIN_PASS` | `changeme123` | 默认管理员密码，**必须修改** |
| `CORS_ORIGINS` | `*` | CORS 允许的来源 |
| `UPSTREAM_CACHE_TTL_SECONDS` | `30` | 上游缓存刷新间隔（秒） |
| `LOGGER_BUFFER_SIZE` | `1000` | 日志队列缓冲区大小 |
| `LOGGER_FLUSH_SIZE` | `100` | 日志批量写入阈值 |
| `RATE_LIMIT_RATE` | `20` | 每个 API Key 每秒允许的请求数 |
| `RATE_LIMIT_BURST` | `40` | 每个 API Key 的突发请求上限 |

---

## 项目结构

```
cmd/server/main.go          # 应用入口
internal/
  config/                    # 环境变量配置
  database/                  # SQLite 初始化与种子数据
  model/                     # 数据模型（User, ApiKey, Upstream, RequestLog）
  service/                   # 业务逻辑（鉴权、缓存、异步日志）
  handler/                   # 管理端 HTTP 接口
  middleware/                # JWT / API Key / 限流 / CORS
  proxy/                     # 反向代理核心（Director / Engine / Handler）
  router/                    # 路由注册
web/                         # React 管理前端
```
