# house-manager

AI 找房与发房项目的 Go 后端服务。

当前项目正在从 Python `ai_house` 迁移到 Go。迁移原则：

- 保留已确认业务行为。
- 不兼容旧数据库结构、旧前端临时接口或 Python 技术边界。
- 当前事实源以 [shared-docs](./shared-docs/README.md) 为准。

## 先读哪里

想了解项目，不要从目录开始翻，按这个顺序读：

1. [shared-docs/README.md](./shared-docs/README.md)：项目当前状态和文档入口。
2. [shared-docs/changes/go_backend/current-plan.md](./shared-docs/changes/go_backend/current-plan.md)：迁移进度、已完成事项、下一步计划。
3. [shared-docs/backend/architecture.md](./shared-docs/backend/architecture.md)：Go 后端分层和代码边界。
4. [shared-docs/api/miniapp-api.md](./shared-docs/api/miniapp-api.md)：小程序 API 契约。
5. [shared-docs/api/publish.md](./shared-docs/api/publish.md)：发房端 API 契约。

## 当前后端形态

代码按“端侧应用服务 + 内部领域能力”组织：

```text
handler/v1/{terminal}/{module}
  -> service/{terminal}/{module}
    -> domain/{capability}
      -> repository/{data-module}
        -> MongoDB / Redis
```

当前已落地的关键链路：

```text
handler/v1/miniapp/auth
  -> service/miniapp/auth
    -> domain/auth

handler/v1/miniapp/house
  -> service/miniapp/house
    -> repository/hpd

handler/v1/publish
  -> service/publish
    -> domain/hmd
    -> domain/hpd
```

## 当前接口状态

已接入：

- 小程序认证：
  - `POST /api/v1/auth/wechat_login`
  - `POST /api/v1/auth/wechat_register`
  - `POST /api/v1/auth/session`
- 小程序找房：
  - `POST /api/v1/house/search`
  - `POST /api/v1/house/public_detail`
- 发房端第一阶段 HMD 接口：
  - `POST /api/v1/publish/{action}`

待接入：

- 小程序用户资料、收藏、足迹。
- 管理端 API。

## 目录结构

```text
cmd/server/                 服务入口
config/                     YAML 配置
internal/
  app/                      Gin 应用与 RouteGroup 注册
  config/                   配置读取
  domain/                   内部领域能力
    auth/                   微信身份、用户初始化
    hmd/                    房源主数据领域能力
    hpd/                    HPD 展示层投影
  handler/                  HTTP handler
    v1/miniapp/auth/        小程序认证接口
    v1/miniapp/house/       小程序找房接口
    v1/publish/             发房端接口
  middleware/               Auth / Logger / Recovery / RateLimit
  model/                    结构、枚举、字段校验
  repository/               Mongo repository
  service/                  端侧应用服务
    miniapp/auth/
    miniapp/house/
    publish/
pkg/                        基础设施与通用包
wire/                       Wire provider
shared-docs/                共享文档 submodule
```

## 快速启动

```bash
# 打通服务器 Mongo / Redis 隧道
./dev.sh

# 启动服务
./run.sh
```

默认读取：

```text
config/config.local.yaml
```

敏感信息通过 `.env` 或环境变量覆盖：

```text
MONGODB_USERNAME
MONGODB_PASSWORD
REDIS_PASSWORD
```

## 常用验证命令

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
go build ./...
```

HMD Mongo 集成测试默认跳过，显式开启：

```bash
PUBLISH_HMD_INTEGRATION=1 \
PUBLISH_HMD_TEST_DB=rent-house \
PUBLISH_HMD_TEST_AUTH_SOURCE=rent-house \
PUBLISH_HMD_TEST_ALLOW_RENT_HOUSE=1 \
go test ./internal/domain/hmd -run Integration -count=1 -v
```

普通 `go test ./...` 不访问 Mongo。
