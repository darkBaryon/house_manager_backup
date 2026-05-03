# house-manager

AI租房运管端 — Go 后端服务

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.25 |
| HTTP | Gin | v1.10 |
| 数据库 | MongoDB | driver v2 |
| 缓存 | Redis | go-redis v9 |
| 认证 | Redis Session | crypto/rand token |
| DI | Wire | v0.6 |
| 配置 | Viper | v1.19 |
| 日志 | Slog | stdlib |
| 限流 | token bucket | golang.org/x/time |

## 项目结构

```
cmd/server/main.go       程序入口（信号管理、优雅关闭）
internal/
  app/                    应用引导（RouteGroup 路由注册、优雅关闭）
  config/                 Viper YAML 配置加载
  handler/                RouteRegistrar 接口定义
    v1/                   v1 handlers（POST 风格）
    v2/                   v2 handlers（RESTful GET）
  middleware/             Gin 中间件（Recovery/Logger/Auth/RateLimit）
  model/                  领域模型（BaseModel、LiteBaseModel、PageReq）
  repository/             数据访问层（泛型 Repository[T] + TxManager）
  service/                业务逻辑层（Cache-Aside、errcode 映射）
pkg/
  cache/                  缓存封装（泛型 GetSet[T]、singleflight、TTL jitter）
  database/               MongoDB / Redis 客户端封装
  errcode/                统一错误码（code/message/cause 三元组）
  logger/                 Slog 初始化
  response/               统一 JSON 响应（Success/Error/SuccessPage）
  session/                Redis Session 存储（Create/Get/Delete）
  configpath/             配置文件路径解析
wire/                     Wire DI 汇编（providers 按模块拆分）
```

## 架构

```
Request → Gin Middleware Chain → Handler → Service → Repository → MongoDB
                                      ↕                  ↕
                                   errcode            Cache (Redis)
                                   response          singleflight
```

**分层职责：**
- **Handler**：参数绑定、响应输出，定义 `RouteRegistrar` 接口，v1/v2 通过目录隔离
- **Service**：业务编排、缓存策略，定义 `RoomServicer` 等消费端接口
- **Repository**：泛型 `Repository[T]` 提供通用 CRUD，`TxManager` 管理事务
- **App**：通用 `RouteGroup` 注册，不感知版本概念
- **Wire**：编译时依赖注入，路由配置集中到 `providers_router.go`

**路由配置**（声明式，集中在 `wire/providers_router.go`）：
```go
return []app.RouteGroup{
    {Prefix: "/api/v1", Registrars: []handler.RouteRegistrar{healthH}},
    {Prefix: "/api/v1", Middleware: []gin.HandlerFunc{auth}, Registrars: []handler.RouteRegistrar{roomH}},
    {Prefix: "/api/v2", Middleware: []gin.HandlerFunc{auth}, Registrars: []handler.RouteRegistrar{v2RoomH}},
}
```

## 快速开始

```bash
# 启动依赖（MongoDB + Redis）
./dev.sh

# 启动服务（本地配置）
go run ./cmd/server -c ./config/config.local.yaml

# 指定配置文件启动
go run cmd/server/main.go -c ./config/config.test.yaml
```

## 配置架构

- 默认读取 `./config/config.local.yaml`（若未传 `-c`）
- 可用 `-c` 显式覆盖配置文件路径
- 敏感信息通过环境变量覆盖（前缀 `HM_`）
  - `HM_MONGODB_USERNAME`
  - `HM_MONGODB_PASSWORD`
  - `HM_REDIS_PASSWORD`
- 示例见 `.env.example`

## 开发命令

```bash
go build ./...
go test ./...
go test ./internal/handler/v1/  # 单包测试
go vet ./...
go mod tidy
```

## API 约定

**认证**：`Authorization: Bearer <session-token>`（Redis Session）

**v1 请求风格**（POST + JSON body）：
```
POST /api/v1/room/detail   {"id": "..."}
POST /api/v1/room/list     {"offset": 0, "limit": 20, "sort": ["createdAt:desc"]}
POST /api/v1/health/check
```

**v2 请求风格**（RESTful GET + query params）：
```
GET /api/v2/rooms/:id
GET /api/v2/rooms?offset=0&limit=20
```

**响应 — 成功：**
```json
{"code": 0, "error": "", "data": {...}}
```

**响应 — 分页：**
```json
{"code": 0, "error": "", "maxSize": 100, "size": 20, "data": [...]}
```

**响应 — 失败：**
```json
{"code": 20001, "error": "房间不存在", "data": null}
```

## 命名规范

- ID 命名：`Id` 不用 `ID`（`roomId` 非 `roomID`）
- BSON 标签：全小写无分隔符（`bson:"createdat"`）
- JSON 标签：camelCase（`json:"createdAt"`）
