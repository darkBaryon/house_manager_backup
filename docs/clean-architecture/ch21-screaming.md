# 第 21 章 尖叫的架构（Screaming Architecture）

## 本章讲什么

看建筑图纸一眼能分辨教堂还是住宅——**软件顶层结构也该"尖叫"出它的业务，
而不是它用的框架**。满眼 controllers/models/views 的项目在喊"我是 Rails 工程"；
好项目该喊"我是医疗系统/租房系统"。推论：框架是工具不是主人，顶层目录不该向
框架宣誓效忠；且好架构应该允许不启动框架、不连数据库就能测试业务规则。

## 对照本项目

诚实版判词：**部分及格**。目录第一层（internal/handler、service、domain、repository）
喊的是"我是个分层后端"（技术词），业务名字（favorite、publish、chat、publishaccess）
第二层才出现。顶层没有失守给框架（handler/middleware 是仅有的 Gin 腔，被压在第二层），
但也没有喊业务。

## 讨论沉淀（2026-07）

### package by layer vs package by feature

本章立场对接的著名争论。两种组织法与各自的税：

```text
layer-first（本项目）             feature-first（尖叫版）
internal/                        internal/
  handler/  ← 技术词                favorite/  ← 业务词，内含自己的 handler+service+repo
  service/                          chat/
  domain/                           publish/
  repository/
```

- layer-first 买到：**分层规则好执行**——"handler 不许 import repository"一条
  depguard 规则管全项目。牺牲：业务一眼可见 + 功能内聚；
- feature-first 买到：第一眼即业务清单；一个功能的变更落在一个目录（CCP）；
  将来拆组件/拆服务整目录端走。牺牲：每个功能目录内部还要再分层，纪律执行成本
  逐目录付，新人易在目录内把层写串。

本项目的选择（第一层技术分层、第二层按端和业务切）是 Go 后端服务的主流折中，
README 第一段即写明业务，实际损失不大——**不值得返工**。

### 下个项目第一天清单（追加项）

新项目若功能模块边界清晰、且预期某些模块将来独立成服务：**feature-first 起步**，
让第 13 章（CCP/组件化路径）和第 21 章（尖叫）同时满意。

一句话：尖叫检验的是目录第一眼透露的信息结构——本项目第一眼喊架构风格、
第二眼才喊业务，**口音略重但能听懂**。
