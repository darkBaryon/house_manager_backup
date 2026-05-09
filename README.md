# house-manager

AI 找房与发房项目的 Go 后端服务。

当前项目处于从 Python `ai_house` 迁移到 Go 的阶段。迁移原则是：保留已确认业务行为，但不兼容旧数据库结构、旧前端临时接口或 Python 技术边界。

## 当前状态

已完成并验证：

- 配置加载：YAML / `.env` / 环境变量覆盖
- MongoDB / Redis 基础设施接入
- Auth 主链路：微信登录/注册服务、Redis Session、Bearer token 中间件
- HMD repository：房源主数据 6 个 collection 的基础 CRUD 与字段约束
- Publish 第一阶段：发房端 HMD 录入、详情、列表、更新、房态更新
- Publish 分层测试：HMD Mongo 集成测试、Publish facade 单测、Handler HTTP binding 测试
- Publish 真实服务联调：route / wire / middleware / Redis session / handler / service / Mongo 落库已跑通

仍未完成：

- HPD 发布与展示层数据
- 小程序展示层依赖的 HPD projector / outbox
- 后台管理系统 API 与 handler / service

## 技术栈

| 组件 | 技术 |
| --- | --- |
| 语言 | Go |
| HTTP | Gin |
| 数据库 | MongoDB driver v2 |
| 缓存 | Redis go-redis v9 |
| 认证 | Redis Session |
| 配置 | Viper |
| 日志 | slog |
| DI | Wire |

## 项目结构

```text
cmd/server/                 服务入口
config/                     YAML 配置
internal/
  app/                      Gin 应用与 RouteGroup 注册
  config/                   配置读取与转换
  handler/                  HTTP handler
    v1/auth.go              auth 接口
    v1/publish/             发房系统第一期接口
  middleware/               Auth / Logger / Recovery / RateLimit
  model/                    领域模型、枚举、字段校验
  repository/               Mongo repository
    auth/                   用户与认证数据
    common/                 泛型 repository 基础能力
    hmd/                    房源主数据 repository
  service/                  业务 service
    auth/                   auth 子服务
    publish/                发房域 facade
      hmd/                  HMD 子 service
      hpd/                  HPD 预留入口，当前 no-op
pkg/
  database/mongo            Mongo 基础设施
  database/redis            Redis 基础设施
  errcode                   统一业务错误码
  response                  统一响应
  session                   Redis session store
wire/                       Wire provider
shared-docs/                共享文档 submodule
```

## 分层约定

```text
Router
  -> Middleware
    -> Handler
      -> Service
        -> Repository
          -> MongoDB / Redis
```

- Handler：只处理 HTTP 输入输出、ObjectID 解析、调用 service。
- Service：承载业务动作、跨 repository 编排、错误语义。
- Repository：只负责数据库读写、字段白名单、软删除过滤。
- Model：定义结构、枚举和字段取值校验。
- `pkg/database/*`：基础设施包，不依赖 `internal/config`。

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

## API 约定

接口命名遵守：

```text
POST /api/v{version}/{模块}/{动作}
```

不使用 RESTful 风格，也不使用 `/模块/对象/动作` 这类多级业务路径。

认证：

```text
Authorization: Bearer <session-token>
```

当前主要接口：

```text
POST /api/v1/health/check
POST /api/v1/auth/wechat_login
POST /api/v1/auth/wechat/register
POST /api/v1/auth/session
POST /api/v1/publish/{action}
```

publish 第一阶段已验证的业务对象：

- 集中式项目
- 楼栋
- 集中式房型
- 集中式房间
- 分散式小区
- 分散式房间

响应格式：

```json
{
  "code": 0,
  "error": "",
  "data": {}
}
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
go test ./internal/service/publish/hmd -run Integration -count=1 -v
```

说明：

- 普通 `go test ./...` 不访问 Mongo。
- 集成测试只清理 `itest_` 前缀测试数据。
- 有独立测试库时优先使用独立测试库。

## 文档入口

共享文档在 [shared-docs](./shared-docs/README.md)。

常用入口：

- [项目通用代码规范](./shared-docs/overview/project-spec.md)
- [项目背景](./shared-docs/overview/project-background.md)
- [Go 后端架构](./shared-docs/backend/index.md)
- [发房系统 API](./shared-docs/api/publish.md)
- [发房域后端设计](./shared-docs/backend/publish.md)
- [小程序 HPD 规划](./shared-docs/backend/miniapp-hpd.md)
- [当前迁移进度](./shared-docs/changes/migration/current-status.md)
- [下一步计划](./shared-docs/changes/migration/next-steps.md)
