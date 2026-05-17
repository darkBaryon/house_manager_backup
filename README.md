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
      -> repository/{module-or-terminal_module}
        -> MongoDB / Redis
```

当前已落地的关键链路：

```text
handler/v1/miniapp/auth
  -> service/miniapp/auth
    -> repository/miniapp_auth.UserAuthRepository / UserRepository / UserProfileExtRepository

handler/v1/miniapp/house
  -> service/miniapp/house
    -> repository/hpd.MiniappListingRepository

handler/v1/publish
  -> service/publish
    -> domain/hmd
    -> domain/listingprojection.Service
    -> domain/publishaccess.Service
```

三端边界：

- `miniapp`：小程序端，只走 `handler/v1/miniapp/* -> service/miniapp/*`，接口契约见 `shared-docs/api/miniapp-api.md`。
- `publish`：发房 Web，只走 `handler/v1/publish/* -> service/publish` 和 `service/publish/auth`，接口契约见 `shared-docs/api/publish.md`。
- `admin`：后台管理端后续独立走 `handler/v1/admin/* -> service/admin/*`，可以复用 `domain/*` 和 `repository/*`，但不复用 miniapp 或 publish 的端侧 service。

四层命名规则：

- `model`：表达当前 Mongo 数据结构和枚举，不放 HTTP DTO。按数据库模块分包；每个模块目录固定 `model.go` 放集合常量和结构，`enum.go` 放枚举和值域方法，`validation.go` 放落库结构不变量。
- `repository`：按上层接口模块命名，不按纯数据库模块命名；模块名冲突时必须加 terminal 前缀，例如 `miniapp_auth`、`publish_auth`。
- `service`：按端侧入口命名，是三个前端的业务边界；跨端共享逻辑下沉到 `domain`。
- `handler`：按端侧和 API 模块命名，只做 HTTP binding / DTO mapping / response，不直接编排 repository。

Validation 边界：

- `handler`：校验 HTTP DTO 形状、必填字段、ObjectID 字符串等传输层问题。
- `service/domain`：校验业务规则、权限、数据作用域、状态流转和跨集合一致性。
- `model/{module}/validation.go`：校验单个 Mongo model 的落库结构不变量，例如枚举取值、必填字段、非负数。
- `repository`：只负责落库前调用对应 model validation，不承载业务判断。

## 当前接口状态

已接入：

- 小程序认证：
  - `POST /api/v1/auth/wechat_login`
  - `POST /api/v1/auth/wechat_register`
  - `POST /api/v1/auth/session`
- 小程序找房：
  - `POST /api/v1/house/search`
  - `POST /api/v1/house/public_detail`
- 小程序用户资料、收藏、足迹：
  - `POST /api/v1/user/profile`
  - `POST /api/v1/user/update_profile`
  - `POST /api/v1/user/dashboard`
  - `POST /api/v1/favorite/add`
  - `POST /api/v1/favorite/remove`
  - `POST /api/v1/favorite/list`
  - `POST /api/v1/history/add`
  - `POST /api/v1/history/list`
- 发房端认证：
  - `POST /api/v1/publish_auth/login`
  - `POST /api/v1/publish_auth/session`
  - `POST /api/v1/publish_auth/logout`
- 发房端第一阶段 HMD 接口：
  - `POST /api/v1/{业务module}/{action}`

待接入：

- 管理端 API。

## 目录结构

```text
cmd/server/                 服务入口
config/                     YAML 配置
internal/
  app/                      Gin 应用与 RouteGroup 注册
  config/                   配置读取
  domain/                   内部领域能力
    hmd/                    房源主数据领域能力
    listingprojection/      HMD 变更到 HPD read model 的投影
    publishaccess/          发房端 root owner scope 与数据作用域
  handler/                  HTTP handler
    v1/miniapp/auth/        小程序认证接口
    v1/miniapp/house/       小程序找房接口
    v1/publish/             发房端接口
  middleware/               Auth / Logger / Recovery / RateLimit
  model/                    数据库模型分包
    common/                 通用落库字段与状态
      model.go
    auth/                   账号与身份：hs_usr_* / hs_lld_* / hs_adm_*
      model.go
      enum.go
      validation.go
    hmd/                    HMD 主数据模型
      model.go
      enum.go
      validation.go
    hpd/                    HPD 展示与归属模型
      model.go
      enum.go
      validation.go
    useractivity/           用户行为：收藏、足迹等
      model.go
      enum.go
      validation.go
  repository/               Mongo repository
    miniapp_auth/           小程序认证：用户、微信绑定、资料扩展
    landlord/               房东主体与密码认证仓储
    favorite/               小程序收藏
    history/                小程序足迹
    hmd/                    HMD 主数据读写
    hpd/                    HPD listing / miniapp & publisher read model / root owner scope relation
  service/                  端侧应用服务
    miniapp/auth/
    miniapp/house/
    miniapp/favorite/
    miniapp/history/
    miniapp/user/
    publish/
    publish/auth/
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

## CI

当前 GitHub Actions 只跑基础后端检查：

```text
gofmt -l .
go test ./...
go build ./...
```

默认 CI 不连接 Mongo / Redis；需要真实库的集成测试仍然按下面的环境变量手动开启。

HMD Mongo 集成测试默认跳过，显式开启：

```bash
PUBLISH_HMD_INTEGRATION=1 \
PUBLISH_HMD_TEST_DB=rent-house \
PUBLISH_HMD_TEST_AUTH_SOURCE=rent-house \
PUBLISH_HMD_TEST_ALLOW_RENT_HOUSE=1 \
go test ./internal/domain/hmd -run Integration -count=1 -v
```

普通 `go test ./...` 不访问 Mongo。

本地联调 publish 房东账号可用：

```bash
./scripts/seed_publish_auth_collections.sh
```

默认写入 `hs_lld_landlord.phone=18002584637` 和 `hs_lld_auth` 密码认证，默认密码为 `123456`。如需自定义密码，请传入 `PUBLISH_AUTH_SEED_PASSWORD_HASH`（bcrypt hash）。

## 日志约定

当前后端日志分两层：

- `access log`：统一由 [internal/middleware/logger.go](/Users/xinyue/VSCode/ws_2026/house-manager/internal/middleware/logger.go) 输出，请求成功时保持短格式，请求失败时展开 `app_code / error_detail / handler / req_*`。
- `business flow log`：由具体 service / domain 文件输出，成功和失败都打，用来表达业务过程，而不只是报错。

`request_id` 会通过 [pkg/requestlog/context.go](/Users/xinyue/VSCode/ws_2026/house-manager/pkg/requestlog/context.go) 贯穿 middleware 和业务日志，排障时先看 access log，再按同一个 `request_id` 串业务流。

当前已经补齐的主链：

- `publish_auth` 与 `service/publish/*`
- `domain/publishaccess`
- `domain/listingprojection`
- `service/miniapp/auth`
- `service/miniapp/house`
- `service/miniapp/favorite`
- `service/miniapp/history`
- `service/miniapp/user`
