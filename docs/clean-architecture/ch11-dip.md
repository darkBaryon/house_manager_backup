# 第 11 章 DIP：依赖反转原则

## 这章讲什么

依赖反转原则是 SOLID 的压轴，也是整本书后半部分（边界、整洁架构）的理论根基：

> 高层策略不应该依赖低层细节；两者都应该依赖**抽象**。
> 源码依赖应当只指向抽象，不指向具体实现。

为什么叫"反转"？自然写代码时，依赖跟控制流同向：业务逻辑调用数据库代码，于是
业务逻辑 import 数据库包——高层依赖了低层。DIP 的手法是在中间插一个属于**高层**的
接口：业务逻辑调用自己定义的接口，数据库代码来实现它。控制流没变（还是业务调
数据库），但**源码依赖的箭头反过来了**——数据库实现指向业务的接口。这就是
"反转"的字面意思，也是跨越架构边界的唯一手法。

这章的配套概念：

- **稳定抽象**：接口比实现稳定，所以依赖应尽量落在接口上；改接口牵连所有实现，
  改实现谁也不牵连。
- **不要依赖易变的具体类**：不 import、不继承、不在业务代码里 `new` 它们。
- **工厂/组装点**：总得有地方创建具体对象。把这种"脏活"集中到一个专门的组件
  （工厂、Main），让系统其余部分保持只依赖抽象。
- 100% 遵守做不到（比如谁都依赖标准库的具体类型），但标准库**稳定**，依赖稳定的
  具体类无害。DIP 真正防的是**易变的**具体实现。

## 对照本项目

本项目里 DIP 不是理论，是每天在跑的结构。

**1. 接口放在高层，实现放在低层。**
`AIResponder` 定义在 `internal/service/chat/ports.go`（用例层，高层），实现
`pythonchat.Client` 在 `internal/integration/`（细节层，低层）。运行时控制流是
`service -> pythonchat -> HTTP`，源码依赖却是 `pythonchat`（通过 wire）指向
`service/chat` 的接口。箭头和控制流反向——这就是教科书上那张 DIP 图的实物。

**2. `wire/` 就是书里说的"组装点"。**
看 `wire/providers_chat.go`：

```go
func newChatAIResponder(cfg *config.Config) (chatsvc.AIResponder, error) {
    ...
    return pythonchat.NewClient(pythonchat.Config{...}), nil
}
```

注意返回类型是**接口** `chatsvc.AIResponder`，函数体里 `new` 的是**具体类**
`pythonchat.Client`。全项目对易变具体类的构造，都被圈禁在 `wire/` 这十几个
provider 文件里；`service/`、`domain/` 的代码里见不到一句 `pythonchat.NewClient`。
这正是 DIP 说的"把违反抽象的脏活集中到组装组件"——wire 目录就是全系统最脏、
也最外圈的地方，而且它理应如此。

**3. 依赖规则给了分层图方向。**
README 的分层 `handler -> service -> domain -> repository` 单看像是"高层依赖
低层"，但注意每一层往下依赖的方式：chat 用例对外部 AI 的依赖走 port，Python
回调 Go 走 internal tools 的窄接口。越靠近 `domain/` 的代码越不知道外面的世界——
`domain/hmd` 里没有任何 handler、Gin、HTTP 的影子。业务规则处在依赖箭头的终点，
这正是第 8 章 OCP 埋的伏笔在 DIP 这里收束：**保护谁，就让所有箭头指向谁。**

**4. chat 模块三个方向全部完成了反转。**
（更正：本节初版误写为"`service/chat` 直接依赖 `*repochat.SessionRepository`
具体类型"，细读 `service.go` 后发现不实。）实际情况是 `service/chat` 在包内声明了
**未导出**的消费侧窄接口：

```go
// internal/service/chat/service.go
type sessionRepository interface { FindActiveByUserID(...); Create(...); ... }
type messageRepository interface { Create(...); ListRecentBySessionID(...) }
type runtimeContextStore interface { Get(...); Set(...) }
```

`Service` 持有的全是这些接口，具体的 `*repochat.SessionRepository` 只在 wire 注入时
出现。所以 chat 对 AI（`AIResponder`）、对 Mongo、对 Redis 三个方向全部反转，接口
全部住在消费侧且未导出（纯为自己声明）——Go 社区公认的最佳形态。测试能只靠假实现
把 Send 全链路测完，就是这三次反转的直接红利。

## 一个容易犯的错

把"依赖注入框架"当成 DIP 本身。DIP 是**源码依赖方向**的原则，用不用 wire/DI
框架只是组装手段。哪怕手写 `main()` 里逐个 new 再传参，只要业务代码依赖的是
自己定义的接口，DIP 就成立；反之用了再花哨的 DI 框架，接口定义在实现方包里，
依赖方向照样是错的。

## 思考题

1. 画出 chat 链路的两张图：控制流图和源码依赖图。标出两者方向相反的那条边。
2. 如果把 `RuntimeContextStore` 从 Redis 换成 Mongo，需要动哪些文件？答案里
   有没有 `service/chat` 之外的用例代码？（有就说明某处 DIP 没守住。）
3. `internal/config` 被几乎所有层 import。它算"稳定的具体类"还是该被反转的
   细节？如果 config 从 YAML 换成配置中心，波及面有多大？

---

## 讨论沉淀（2026-07）

### 两张图讲法（解"反转"之惑）

"`service/chat` 调用 `pythonchat`"和"`pythonchat` 实现 `service/chat` 的接口"
不矛盾，因为是两张不同的图：

```text
调用图（运行时）：  service/chat ────────> pythonchat    （由业务决定，永远不变）
依赖图（源码）：    service/chat <──────── pythonchat    （被接口掰反了）
```

天真写法里两图同向（要调谁就 import 谁）；在调用方自己包里声明接口后，调用照旧、
import 反向。**"反转"反的就是源码依赖这一根箭头。**配套动作：new 具体实现的"脏话"
全部赶进 `wire/`（Main 组件）——全项目搜 `pythonchat.NewClient` 只有一处。
DIP 不是消灭对具体类的依赖，是把它圈禁到一个没有业务逻辑可污染的角落。

豁免条款：依赖**稳定的**具体类型无害（标准库；以及第 7 章讨论过的 `pkg/errcode`
合同——成立前提是它像标准库一样稳）。判断标准不是"是不是接口"，
而是"会不会变、谁让它变"。

### 自测题及答案：换成直连 Claude API，diff 落在哪

会出现：`internal/integration/claudechat/`（纯新增）、`wire/providers_chat.go`
（换一行 new）、`internal/config/`（加配置项）。
保证不出现：`service/chat`、`handler/v1/miniapp/chat`、`repository/chat`、所有 domain 包。

**凭什么敢保证：这些包到新实现之间不存在 import 边。**不是"小心点就不会改到"的
软保证，是编译器层面的硬事实——没有依赖边，变更就没有传播路径。架构的本质是
管理依赖边；边不存在，波及就不可能发生。

残留的坑（接第 9 章）：签名照抄就能编译，LSP 行为契约得人守——超时返回 error
而非挂起、失败不留半截状态、output 语义一致。换实现那天拿现有 `service_test.go`
场景对新实现跑一遍（契约测试）才算真正可替换。

### SOLID 五条串成一句话

> **SRP（单一职责）告诉你在哪画边界，ISP（接口隔离）告诉你边界上的门开多窄，
> DIP（依赖反转）告诉你门朝哪边开，LSP（里氏替换）保证走这扇门的人守规矩，
> OCP（开闭）是这一切换来的回报——下次需求来时，老代码不动。**

---

到这里 Part III（SOLID）读完了。这五条原则解决的是"模块层面怎么画依赖"；
第 12~14 章会把同样的思路提升到**组件**（可部署单元）层面。想继续时说一声即可。
