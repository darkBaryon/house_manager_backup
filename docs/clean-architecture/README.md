# 《Clean Architecture》读书笔记（结合 house-manager 项目）

这是《架构整洁之道》（Robert C. Martin, *Clean Architecture: A Craftsman's Guide to
Software Structure and Design*）的中文导读笔记。每章不复述原文，而是用自己的话讲清
"这章解决什么问题"，再从本仓库（house-manager Go 后端）里找真实代码对照，最后留几个
思考题，方便对着代码消化。

从第七章开始读（第 7~11 章是全书 Part III：SOLID 设计原则）。

## 目录

### Part III 设计原则（SOLID）

| 章 | 文件 | 主题 | 状态 |
|---|---|---|---|
| 7 | [ch07-srp.md](./ch07-srp.md) | SRP 单一职责原则 | 已讨论，含沉淀 |
| 8 | [ch08-ocp.md](./ch08-ocp.md) | OCP 开闭原则 | 已讨论，含沉淀 |
| 9 | [ch09-lsp.md](./ch09-lsp.md) | LSP 里氏替换原则 | 已讨论，含沉淀 |
| 10 | [ch10-isp.md](./ch10-isp.md) | ISP 接口隔离原则 | 已讨论，含沉淀 |
| 11 | [ch11-dip.md](./ch11-dip.md) | DIP 依赖反转原则 | 已讨论，含沉淀（有更正） |

每章末尾的"讨论沉淀"一节是对话讨论的产出，内容深度超过正文，重读时优先看。

### 后续章节（待写，按需继续）

- Part IV 组件原则：第 12~14 章（组件、组件聚合、组件耦合）
- Part V 软件架构：第 15~29 章（架构是什么、划分边界、业务逻辑、整洁架构本体、
  Presenter/Humble Object、部分边界、Main 组件、服务、测试边界、嵌入式）
- Part VI 实现细节：第 30~34 章（数据库、Web、框架都是细节；案例分析；包组织方式）

想继续往下读时，直接说"写第 12~14 章"即可。

## 本项目与书的对应关系速查

书里的核心图景是"依赖只能从外圈指向内圈"。本项目的分层正好是一组同心圆：

```text
外圈  handler/v1/*        —— 接口适配器（Interface Adapters）
      integration/*       —— 外部服务适配器
      repository/*        —— 数据库网关实现
中圈  service/*           —— 用例（Use Cases / Interactors）
内圈  domain/*            —— 业务实体与领域规则（Entities）
组装  wire/*              —— Main 组件（第 26 章），最脏、最外圈的地方
```

## Part III 讨论后的项目体检结论（2026-07）

五条原则里四条做得在水准之上（ISP/DIP 尤其好：消费侧未导出窄接口、wire 组装点、
按端切分 service、HMD/HPD 读写分离）。真正的欠账：

1. **变化轴 B（新增实体类型）未封闭**：加"车位"级别的新实体要横跨两包七八个文件
   （详见 ch08 沉淀；缓解方向 = scope→refresh 注册表化，新实体来了再做）；
2. **同步投影挤占写链路**：发房 = HMD 写入 + 三读模型全刷成功；已规划 outbox 升级，
   `listingProjectionApplier` 接口是预留的插座（详见 ch08 沉淀）。

风险提示——三处"AI 重构生成、作者未验收的承重代码"，值得亲自通读：

- `internal/service/publish/create_compensation.go`（失败补偿回滚——出事时决定数据一致性）
- `internal/domain/listingprojection/refresh_dispatch.go`（投影路由，含 nil 字段隐患——
  详见 ch08 沉淀）
- `internal/domain/publishaccess/`（房东数据隔离鉴权——炸了是安全事故，排第一）

另有两颗有意识保留的雷：`pkg/errcode` 被 domain/service 直接依赖（引信 = 前端提出
错误码规范化，详见 ch07 沉淀）；`AIResponder` 第二个实现出现时需补契约测试
（详见 ch09/ch11 沉淀）。

几个最值得反复看的"活教材"：

- `internal/service/chat/ports.go` —— 一个只有 6 行的接口，却是全书 DIP 思想最纯的体现。
- `wire/providers_chat.go` —— Main 组件如何把具体实现注入抽象端口。
- `internal/domain/hmd/` —— 领域层不 import 任何 handler/repository 的实体与规则。
- README「架构边界」一节 —— 用一页纸写清了依赖方向约束，这正是第 17 章"划分边界"。
