# API 中转网关

一个面向内部团队的通用 API 中转平台。

它解决的是两个问题：

1. 把多个第三方 API 统一收口到同一个网关地址
2. 用内部下游 API Key 控制谁可以访问哪些上游，并保留审计、限流和统计能力

支持的上游类型包括：

- OpenAI / DeepSeek / Claude / Gemini 这类 LLM API
- Google / YouTube / Product Hunt 这类第三方 HTTP API
- 任意自定义 HTTP API

提供的能力包括：

- 管理员登录后台
- 上游注册与测试
- 下游用户管理
- 下游 API Key 生成与撤销
- 上游白名单控制
- 请求日志与统计面板
- 原生鉴权兼容模式

## 架构概览

系统由两部分组成：

- 后端：Go + Gin + GORM + SQLite
- 前端：React + Ant Design + Vite

运行形态有两种：

- 开发模式：前端跑在 `http://localhost:5173`，代理 `/api` 和 `/proxy` 到后端 `http://localhost:8080`
- 生产模式：前端构建到 `static/`，由 Go 服务统一对外提供

核心链路如下：

```text
下游调用方
  -> /proxy/{上游名}/{路径...}
  -> 下游 API Key 鉴权
  -> 上游权限校验
  -> 限流
  -> URL 重写 / 鉴权注入 / 额外请求头注入
  -> 转发到真实上游
  -> 记录日志 / 统计
  -> 返回响应
```

## 角色说明

### 上游

上游是你真正要访问的第三方 API。

例如：

- `https://api.openai.com`
- `https://www.googleapis.com`
- `https://api.producthunt.com`

每个上游在系统里对应一条配置，最关键的字段是：

- `name`：路由名，用于 `/proxy/{name}/...`
- `base_url`：真实上游基础地址
- `auth_type`：网关如何向上游带鉴权
- `auth_key` / `auth_value`：上游官方密钥位置和值

### 下游

下游是使用这个网关的内部调用方。

在系统里，下游由两层组成：

- `用户`：一个人、一个服务、一个业务线都可以建成用户
- `API Key`：真正给调用方使用的明文 Key，格式类似 `sk-xxxx`

一个下游 API Key 可以限制：

- 是否启用
- 请求总额度
- 可访问哪些上游

## 目录结构

```text
cmd/server/main.go          应用入口
internal/config             环境变量配置
internal/database           SQLite 初始化与默认管理员种子数据
internal/model              数据模型
internal/service            用户、API Key、上游、统计、日志等业务逻辑
internal/handler            管理端 HTTP 接口
internal/middleware         JWT、下游 API Key、限流、CORS
internal/proxy              反向代理核心
internal/router             路由注册
web/                        React 管理后台
static/                     前端构建产物，生产模式由后端直接提供
data/                       SQLite 数据库与运行数据
```

## 快速开始

### 1. 准备环境变量

复制配置文件：

```bash
cp .env.example .env
```

至少修改以下内容：

- `JWT_SECRET`
- `DEFAULT_ADMIN_PASS`

### 2. 启动后端

```bash
go mod tidy
go run ./cmd/server
```

启动后默认监听：

- 服务地址：`http://localhost:8080`
- 健康检查：`GET /api/health`

### 3. 启动前端开发服务

```bash
cd web
npm install
npm run dev
```

开发地址：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080`

如果前端不是由网关同域提供，而是单独部署或单独开发，建议设置：

```bash
VITE_GATEWAY_ORIGIN=http://localhost:8080
```

这样后台复制出来的示例 URL 会明确指向真正的网关地址，而不是管理台自身地址。

### 4. 登录后台

首次启动时，如果数据库中还没有用户，系统会自动创建默认管理员：

- 用户名：`DEFAULT_ADMIN_USER`，默认 `admin`
- 密码：`DEFAULT_ADMIN_PASS`

登录地址：

- 开发模式：`http://localhost:5173`
- 生产模式：`http://localhost:8080`

## 部署方式

### 方式一：Docker Compose

最简单的部署方式：

```bash
docker compose up -d --build
```

默认映射：

- 容器内服务端口：`8080`
- 宿主机端口：`.env` 中的 `HOST_PORT`，默认 `8081`

部署完成后访问：

```text
http://你的服务器IP:8081
```

### 方式二：二进制 / 进程方式

后端：

```bash
go build -o bin/api-gateway ./cmd/server
./bin/api-gateway
```

前端：

```bash
cd web
npm install
npm run build
```

前端构建产物会输出到 `static/`，之后由 Go 服务直接托管。

### 方式三：开发分离部署

适合本地联调：

- 后端独立跑在 `:8080`
- 前端 Vite 跑在 `:5173`

这时：

- 浏览器访问管理台用 `:5173`
- 对外给调用方使用的网关地址仍然应该是 `:8080`

## 使用顺序

推荐按这个顺序配置：

1. 登录后台
2. 创建上游
3. 创建下游用户
4. 给下游用户生成 API Key
5. 把下游 API Key 发给调用方
6. 调用 `/proxy/{上游名}/...`

## 如何添加上游

后台路径：

- `上游 API 管理`

点击：

- `新建上游`

### 上游字段说明

| 字段 | 是否必填 | 说明 |
|------|----------|------|
| `name` | 是 | 路由标识。调用时使用 `/proxy/{name}/...` |
| `display_name` | 是 | 后台显示名称 |
| `base_url` | 是 | 真实上游基础地址，必须以 `http://` 或 `https://` 开头 |
| `auth_type` | 是 | 上游鉴权方式，可选 `none`、`bearer`、`header`、`query` |
| `auth_key` | 条件必填 | 当 `auth_type` 是 `header` 或 `query` 时必填 |
| `auth_value` | 否 | 上游官方密钥或 Token |
| `allow_native_client_auth` | 否 | 是否允许下游调用方用上游原生位置传内部员工 Key |
| `timeout_seconds` | 否 | 上游超时时间，默认 `120` 秒 |
| `proxy_url` | 否 | 访问上游时要走的代理，例如 `http://127.0.0.1:7890` |
| `strip_prefix` | 否 | 是否去掉 `/proxy/{name}` 前缀，默认开启 |
| `extra_headers` | 否 | 额外请求头，JSON 格式 |
| `is_active` | 否 | 是否启用 |
| `description` | 否 | 说明、备注、模型名、推荐路径等 |

### 四种上游鉴权方式

#### `none`

网关不额外向上游注入鉴权。

#### `bearer`

网关会把 `auth_value` 作为：

```text
Authorization: Bearer {auth_value}
```

#### `header`

网关会把 `auth_value` 写到：

```text
{auth_key}: {auth_value}
```

#### `query`

网关会把 `auth_value` 追加到：

```text
?{auth_key}={auth_value}
```

### 原生鉴权兼容模式

默认情况下，下游调用方访问网关都应该使用：

```text
Authorization: Bearer sk-你的下游APIKey
```

如果你开启 `allow_native_client_auth`，则网关还支持下游调用方把内部 API Key 放到上游原生位置：

- `query` 上游：下游可以把内部 API Key 放到对应 query 参数中
- `header` 上游：下游可以把内部 API Key 放到对应请求头中
- `bearer` 上游：仍然建议继续用 `Authorization: Bearer`

注意：

- 下游调用方传入的是内部 API Key，不是上游官方 API Key
- 网关转发时会把它替换成上游配置里的 `auth_value`

### 上游示例

#### OpenAI

建议配置：

| 字段 | 值 |
|------|----|
| `name` | `openai` |
| `display_name` | `OpenAI API` |
| `base_url` | `https://api.openai.com` |
| `auth_type` | `bearer` |
| `auth_value` | `sk-proj-xxx` |

调用示例：

```bash
curl http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-你的下游APIKey" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"hello"}]}'
```

#### Google / YouTube

建议配置：

| 字段 | 值 |
|------|----|
| `name` | `google` |
| `display_name` | `Google API` |
| `base_url` | `https://www.googleapis.com` |
| `auth_type` | `query` |
| `auth_key` | `key` |
| `auth_value` | `AIza-xxx` |

调用示例：

```bash
curl "http://localhost:8080/proxy/google/youtube/v3/search?part=snippet&q=golang" \
  -H "Authorization: Bearer sk-你的下游APIKey"
```

实际转发到：

```text
https://www.googleapis.com/youtube/v3/search?part=snippet&q=golang&key=AIza-xxx
```

#### Product Hunt GraphQL

建议配置：

| 字段 | 值 |
|------|----|
| `name` | `producthunt` |
| `display_name` | `Product Hunt GraphQL` |
| `base_url` | `https://api.producthunt.com` |
| `auth_type` | `bearer` |
| `auth_value` | `ph_xxx` |

调用示例：

```bash
curl http://localhost:8080/proxy/producthunt/v2/api/graphql \
  -H "Authorization: Bearer sk-你的下游APIKey" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ posts(first:3) { edges { node { name tagline } } } }"}'
```

## 如何添加下游

后台里的下游配置分两步。

### 第一步：创建用户

后台路径：

- `用户管理`

创建一个用户，用来标识这个 API Key 属于谁。

适合的建法包括：

- 按个人建：例如 `alice`
- 按服务建：例如 `search-service`
- 按业务线建：例如 `data-platform`

用户字段包括：

- `username`
- `password`
- `display_name`
- `role`

### 第二步：生成 API Key

后台路径：

- `API Key 管理`

点击：

- `生成 Key`

需要填写：

- `所属用户`
- `名称`
- `request_limit`
- `allowed_upstream_ids`

说明：

- `request_limit = 0` 表示不限额
- `allowed_upstream_ids` 留空表示可访问全部上游
- 明文 API Key 只会显示一次，必须在生成时保存

### 下游 API Key 示例

生成后你会拿到类似：

```text
sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

调用方默认使用方式：

```bash
curl http://localhost:8080/proxy/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-你的下游APIKey"
```

## 常见调用方式

### 默认方式

这是最推荐的调用方式：

```bash
curl http://localhost:8080/proxy/{上游名}/{实际路径} \
  -H "Authorization: Bearer sk-你的下游APIKey"
```

### query 原生兼容方式

当上游开启了原生兼容，且 `auth_type=query` 时，可以这样调用：

```bash
curl "http://localhost:8080/proxy/google/youtube/v3/search?part=snippet&q=test&key=sk-你的下游APIKey"
```

### header 原生兼容方式

当上游开启了原生兼容，且 `auth_type=header` 时，可以这样调用：

```bash
curl http://localhost:8080/proxy/custom/resource \
  -H "X-API-Key: sk-你的下游APIKey"
```

## 管理端与代理端路由

### 健康检查

- `GET /api/health`

### 管理端

- `POST /api/auth/login`
- `GET /api/admin/users`
- `POST /api/admin/users`
- `PUT /api/admin/users/:id`
- `DELETE /api/admin/users/:id`
- `GET /api/admin/api-keys`
- `POST /api/admin/api-keys`
- `PUT /api/admin/api-keys/:id`
- `DELETE /api/admin/api-keys/:id`
- `GET /api/admin/upstreams`
- `POST /api/admin/upstreams`
- `POST /api/admin/upstreams/test`
- `PUT /api/admin/upstreams/:id`
- `DELETE /api/admin/upstreams/:id`
- `GET /api/admin/logs`
- `GET /api/admin/stats/overview`
- `GET /api/admin/stats/daily`

### 代理端

- `ANY /proxy/:api_name/*path`

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 后端监听端口 |
| `HOST_PORT` | `8081` | Docker 对外映射端口 |
| `DB_PATH` | `./data/gateway.db` | SQLite 文件路径 |
| `JWT_SECRET` | `change_me_jwt_secret` | 管理端 JWT 密钥 |
| `JWT_EXPIRE_HOURS` | `24` | JWT 过期时间 |
| `DEFAULT_ADMIN_USER` | `admin` | 默认管理员用户名 |
| `DEFAULT_ADMIN_PASS` | `changeme123` | 默认管理员密码 |
| `CORS_ORIGINS` | `*` | 允许访问管理端和代理端的来源 |
| `UPSTREAM_CACHE_TTL_SECONDS` | `30` | 上游缓存刷新间隔 |
| `LOGGER_BUFFER_SIZE` | `1000` | 日志队列缓冲区大小 |
| `LOGGER_FLUSH_SIZE` | `100` | 日志批量写入阈值 |
| `RATE_LIMIT_RATE` | `20` | 每个下游 API Key 每秒允许请求数 |
| `RATE_LIMIT_BURST` | `40` | 每个下游 API Key 的突发上限 |

前端可选环境变量：

| 变量 | 说明 |
|------|------|
| `VITE_GATEWAY_ORIGIN` | 管理台复制示例 URL 时使用的网关地址。适合前后端分离部署 |

## 管理建议

- 生产环境必须修改 `JWT_SECRET` 和默认管理员密码
- 建议为不同服务单独创建下游用户和 API Key，不要多人共用
- 对敏感上游建议限制 `allowed_upstream_ids`
- 对公共服务建议设置合理的 `request_limit`
- 上游官方密钥只保存在网关，不要发给下游调用方
- 如果前端单独部署，记得为示例 URL 配置 `VITE_GATEWAY_ORIGIN`

## 一句话理解

这个系统的核心不是“替你调用某个 API”，而是：

- 统一管理多个上游
- 统一给下游发 Key
- 统一做鉴权、限流、日志和统计

这样你可以把第三方 API 的接入方式稳定下来，而不必把真实上游密钥散落到每个下游服务里。
