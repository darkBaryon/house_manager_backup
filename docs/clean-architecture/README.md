# 《Clean Architecture》读书笔记（结合 house-manager 项目）

这是《架构整洁之道》（Robert C. Martin, *Clean Architecture: A Craftsman's Guide to
Software Structure and Design*）的中文导读笔记。每章不复述原文，而是用自己的话讲清
"这章解决什么问题"，再从本仓库（house-manager Go 后端）里找真实代码对照，最后留几个
思考题，方便对着代码消化。

从第七章开始读（第 7~11 章是全书 Part III：SOLID 设计原则）。

## 目录

### Part III 设计原则（SOLID）

| 章 | 文件 | 主题 |
|---|---|---|
| 7 | [ch07-srp.md](./ch07-srp.md) | SRP 单一职责原则 |
| 8 | [ch08-ocp.md](./ch08-ocp.md) | OCP 开闭原则 |
| 9 | [ch09-lsp.md](./ch09-lsp.md) | LSP 里氏替换原则 |
| 10 | [ch10-isp.md](./ch10-isp.md) | ISP 接口隔离原则 |
| 11 | [ch11-dip.md](./ch11-dip.md) | DIP 依赖反转原则 |

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

几个最值得反复看的"活教材"：

- `internal/service/chat/ports.go` —— 一个只有 6 行的接口，却是全书 DIP 思想最纯的体现。
- `wire/providers_chat.go` —— Main 组件如何把具体实现注入抽象端口。
- `internal/domain/hmd/` —— 领域层不 import 任何 handler/repository 的实体与规则。
- README「架构边界」一节 —— 用一页纸写清了依赖方向约束，这正是第 17 章"划分边界"。
