# house-manager

AI 找房、发房和后台管理的 Go 后端服务。

本项目是从 Python `ai_house` 迁移到 Go 的新后端。迁移只保留当前确认的业务逻辑，不兼容旧前端、旧数据库结构或旧接口形态。接口、数据模型和架构约束以 [shared-docs](./shared-docs/README.md) 为准。

## 先读哪里

1. [shared-docs/README.md](./shared-docs/README.md)：项目共享文档入口。
2. [shared-docs/api/index.md](./shared-docs/api/index.md)：所有 HTTP/API 契约入口。
3. [shared-docs/backend/architecture.md](./shared-docs/backend/architecture.md)：Go 后端分层和代码边界。
4. [shared-docs/schema/db-design/v4/index.md](./shared-docs/schema/db-design/v4/index.md)：当前 Mongo schema。
5. [shared-docs/workstreams/backend-go/status.md](./shared-docs/workstreams/backend-go/status.md)：Go 后端当前状态。

端侧接口文档：

- [shared-docs/api/miniapp/index.md](./shared-docs/api/miniapp/index.md)：小程序接口。
- [shared-docs/api/publish/index.md](./shared-docs/api/publish/index.md)：发房端接口。
- [shared-docs/api/admin/index.md](./shared-docs/api/admin/index.md)：管理后台接口。
- [shared-docs/api/internal/index.md](./shared-docs/api/internal/index.md)：Go / Python 服务间接口。
- [shared-docs/api/system/index.md](./shared-docs/api/system/index.md)：系统接口。

## 当前能力

已落地：

- 小程序端：
  - 微信登录、注册、session 校验
  - 房源搜索、房源详情
  - 用户资料、偏好更新、个人页 dashboard
  - 收藏、取消收藏、收藏列表
  - 足迹写入、足迹列表
  - AI chat：`miniapp -> Go -> Python -> Go internal tools -> Go -> miniapp`
- 发房端：
  - 房东登录、session、退出
  - 集中式项目、楼栋、房型、房间
  - 分散式小区、房间
  - HMD 写入与 HPD read model 投影
- 管理后台：
  - 员工登录、session、退出
  - 员工、角色、发房方管理
  - 后台房源 root / building / room 查询
- 内部接口：
  - Go 调 Python：`POST /internal/chat/respond`
  - Python 调 Go tools：`POST /internal/tools/house/search`
  - Python 调 Go tools：`POST /internal/tools/house/public_detail`

## 架构边界

代码按“端侧入口 + 应用服务 + 领域能力 + 仓储”组织：

```text
handler/v1/{terminal}/{module}
  -> service/{terminal}/{module}
    -> domain/{capability}
      -> repository/{module-or-terminal_module}
        -> MongoDB / Redis
```

AI chat 的 Go/Python 调用边界：

```text
handler/v1/miniapp/chat
  -> service/chat.Service
    -> repository/chat.SessionRepository
    -> repository/chat.MessageRepository
    -> service/chat.RuntimeContextStore port
    -> service/chat.AIResponder port
         <- integration/pythonchat.Client

Python service
  -> handler/internaltools/house
    -> service/miniapp/house
    -> repository/hpd
```

核心约束：

- `service/chat` 只依赖 `AIResponder` port，不直接 import `integration/pythonchat`。
- Go 是 chat session、message 和 runtime context owner。
- Python 不校验小程序 Bearer token，不写 Go 的 Mongo chat collection。
- internal tools 只返回确定性业务事实，不返回 AI 推理内容。

## 目录结构

```text
cmd/server/                 服务入口
config/                     YAML 配置
internal/
  app/                      Gin 应用与 route 注册
  config/                   配置读取
  domain/                   跨端领域能力
  handler/
    internaltools/house/    Python 调 Go 的 internal tools
    v1/admin/               管理后台接口
    v1/miniapp/             小程序接口
    v1/publish/             发房端接口
  integration/
    pythonchat/             Go 调 Python HTTP adapter
  middleware/               Auth / Logger / Recovery / RateLimit
  model/                    Mongo model 与枚举
  repository/               Mongo / Redis repository
  service/                  端侧应用服务与 chat service
pkg/                        基础设施与通用包
scripts/                    种子数据与辅助脚本
shared-docs/                共享文档 submodule
wire/                       Wire provider
```

## 本地启动

本地默认读取：

```text
config/config.local.yaml
```

敏感信息通过 `.env` 或环境变量覆盖：

```text
MONGODB_USERNAME
MONGODB_PASSWORD
REDIS_PASSWORD
AI_CHAT_PYTHON_BASE_URL
AI_CHAT_INTERNAL_TOKEN
AI_CHAT_RESPOND_TIMEOUT
AI_CHAT_TOOL_TIMEOUT
AI_CHAT_SESSION_IDLE_TIMEOUT
AI_CHAT_RECENT_MESSAGE_LIMIT
AI_CHAT_RUNTIME_CONTEXT_TTL
```

启动依赖隧道和服务：

```bash
# 打通 Mongo / Redis 隧道
./dev.sh

# 启动 Go 服务
./run.sh
```

AI chat 联调时，需要另起 Python 服务，默认 Go 会调用：

```text
http://127.0.0.1:5000/internal/chat/respond
```

## 常用命令

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
go build ./...
```

普通 `go test ./...` 不应依赖真实 Mongo / Redis。

## 数据与联调

发房端测试账号数据：

```bash
./scripts/seed_publish_auth_collections.sh
```

HMD Mongo 集成测试默认跳过，显式开启：

```bash
PUBLISH_HMD_INTEGRATION=1 \
PUBLISH_HMD_TEST_DB=rent-house \
PUBLISH_HMD_TEST_AUTH_SOURCE=rent-house \
PUBLISH_HMD_TEST_ALLOW_RENT_HOUSE=1 \
go test ./internal/domain/hmd -run Integration -count=1 -v
```

Chat 数据落库：

- session：`hs_chat_session`
- message：`hs_chat_message`
- runtime context：Redis，key 由 chat runtime context store 管理

## 日志

日志分两层：

- access log：由 `internal/middleware/logger.go` 统一输出 HTTP 请求结果。
- business flow log：由 service / domain / integration 输出关键业务阶段。

排障时先用 access log 找 `request_id`，再按同一个 `request_id` 串联业务日志。AI chat 主链会覆盖：

- `chat.send.*`
- `pythonchat.respond.*`
- `miniapp.house.*`
- `http.request`

## CI

当前 CI 只跑基础后端检查：

```text
gofmt -l .
go test ./...
go build ./...
```

需要真实 Mongo / Redis 的验证在本地或专门集成环境手动执行。
