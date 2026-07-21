# 第 8 章 OCP：开闭原则

## 这章讲什么

开闭原则的口号是：软件实体应当**对扩展开放，对修改关闭**。也就是说，当需求变化时，
理想状态是通过"新增代码"来满足，而不是"改动已有代码"。

这句话单独听很玄：不改代码怎么可能加功能？Martin 在这章给出了可操作的版本——OCP
的实现手段是**依赖方向的设计**。他用一个"财务报表"的例子：同一份报表数据，既要在
网页上显示，又要打印成纸质版。做法是：

1. 按 SRP 把"计算报表"和"呈现报表"拆开；
2. 让呈现侧依赖计算侧的**接口**，而不是反过来；
3. 于是新增一种呈现方式（比如导出 PDF）只需要加一个新实现，计算逻辑一行都不用动。

关键洞见：**如果 A 组件不想被 B 的变化波及，就让 B 依赖 A**。依赖箭头指向谁，
谁就是被保护的一方。整个系统里最该被保护的是业务逻辑，所以业务逻辑应该处在所有
依赖箭头的终点——这其实已经预告了全书"同心圆"的架构图。

## 对照本项目

最标准的例子是 AI chat 的回复方。看 `internal/service/chat/ports.go`：

```go
type AIResponder interface {
    Respond(ctx context.Context, input AIRespondInput) (AIRespondOutput, error)
}
```

chat 用例（`service/chat.Service`）只依赖这个接口。目前的实现是
`integration/pythonchat.Client`，通过 HTTP 调 Python 服务。README 里明文写了约束：

> `service/chat` 只依赖 `AIResponder` port，不直接 import `integration/pythonchat`。

假设哪天要把 Python 服务换成直接调 Claude API，或者加一个"压测用的假 AI"，
做法都是**新增**一个实现 `AIResponder` 的类型，然后在 `wire/providers_chat.go`
里换掉注入的实现。`service/chat` 的会话管理、消息落库、超时控制——一行都不用改。
这就是"对扩展开放（随便加新 responder），对修改关闭（chat 用例不动）"。

反过来想更清楚：如果当初 `service/chat` 直接 import 了 `pythonchat`，那么换 AI
供应商就要动用例代码，改动会波及会话逻辑的测试和稳定性。依赖箭头一转
（`integration -> service` 的接口，而不是 `service -> integration`），
波及就被挡在了圈外。

另一个例子是 route 注册。新增一个端点时，改动集中在新 handler 文件和
`wire/providers_registrars.go` 的注册处，`internal/app` 的 Gin 骨架不需要动——
应用骨架对"新增路由"这类扩展是关闭修改的。

## 一个容易犯的错

OCP 不是要你为每个类都预先抽接口。抽象是有成本的（间接层、导航负担），只应该
架在**预期会变**的轴上。本项目的选择很克制：`AIResponder` 抽了接口，因为
"AI 实现会换"是明确预期；而 `SessionRepository` 直接用具体类型注入，因为
"换掉 Mongo"并不在预期里。100% 的封闭既不可能也不划算——这章原文也承认，
架构师只能针对**最可能的变化**做封闭，这需要判断，也会猜错。

## 思考题

1. 如果要给 chat 加"多模型路由"（简单问题走便宜模型），你会改 `AIResponder`
   接口，还是在接口后面加一个路由实现？哪种符合 OCP？
2. `handler/internaltools/house` 是 Python 反向调 Go 的入口。它的存在保护了谁？
   依赖箭头怎么画？
3. 找一处你觉得"下次需求一来必须改老代码"的地方，想想值不值得现在就抽接口。

---

## 讨论沉淀（2026-07）

### 一秒钟自测法

> **"下一个 X 来的时候，git diff 会落在哪？"**

diff 全在新文件 + 一行注册 → 这条轴 OCP 达成；diff 落进核心函数/老逻辑 →
这条轴开着，再问"X 会经常来吗"决定值不值得抽接口。OCP 说穿了 =
"让下一次需求的 diff 尽量只有绿色（新增行），没有红色（改动行）" + "挑对下注的轴"。

### 走查："对新增读模型关闭修改"

`listingprojection.Service` 持有 `projectors []projector` 切片，`Apply` 循环只认识
`projector` 接口。假想需求"接 Elasticsearch 读模型"：

- **新建** `searchindex_projector.go` / `searchindex_mapper.go`；
- **修改**只有 `NewService` 切片登记一行（组装代码，非逻辑代码）。

`Apply` 循环、`domain/hmd`、六个 publish 子 service 碰都不碰。反例：若 `Apply`
硬编码 `s.miniapp.Refresh(); s.publisher.Refresh(); ...`，每加一个端就要打开核心循环
改一次。两种写法功能相同，差别只在**下次需求的 diff 落在哪**。

整条投影链的依赖方向也是本章正面教材：`domain/hmd` 写完只返回 `HmdChange` 纯数据
回执，完全不知道投影的存在——这是一个用返回值（而非消息队列）实现的进程内领域事件
模式，写侧业务规则处在依赖箭头终点。

### 变化轴 A vs 轴 B（本章最重要的发现）

本设计有两条变化轴，只封闭了一条：

- **轴 A：新增读模型**——封闭得很好（见上），成本≈一个新文件；
- **轴 B：新增实体类型**（如"车位""合同"）——完全未封闭：`hmd_change.go` 枚举、
  `dispatchRefresh` switch、`refreshFuncs` 字段、三个 projector 各自的 refresh/mapper/
  fanout……横跨两包七八个文件。

作者确认：当初是按端侧（SRP 的行为者）切的，不是按 OCP 变化轴切的；且判断轴 B
更可能变。这暴露一个真实的原则冲突：**SRP 按行为者切出的结构恰好封闭轴 A，
而业务上更可能变的是轴 B**——原则之间要做交易是常态。缓解方向（新实体真来了再做，
半天工作量）：scope→refresh 从结构体字段改成注册表 map，各 projector
`register(scope, fn)` 只加不改，switch 和 refreshFuncs 同时消失。

### `refreshFuncs` 的 nil 字段隐患（已确认，暂不修）

`refresh_dispatch.go` 注释称"编译器保证签名匹配"——只对一半：**签名**保证，
**字段填没填**不保证。Go 结构体字面量漏填字段不报错，缺省为 nil；
`dispatchRefresh` 里 `return funcs.Building(ctx, ...)` 若该字段未填 = nil 函数调用
panic，且发生在房东提交房源的写链路上。今天三处字面量填全了没事；风险在按轴 B
加新实体时六个填写点漏一个。修法便宜：调用前判 nil 降级 Warn；或上 `exhaustruct`
lint 强制字面量填全。

### 同步投影链的完整语义与 outbox 升级

`service/publish/publisher.go` 的 `resolveHmdMutation`：HMD 写成功后**同步**调
`publisher.Apply` 刷三个投影，投影失败则整个请求失败；创建操作还会经
`create_compensation.go` 的 `rollbackCreateFailure` 把刚写入的 HMD 实体回滚。
完整语义：**发房 = HMD 写入 + 三个读模型全部刷新，全成才算成，否则补偿回退**。

含义：代码层面扩展免费（OCP 达成），运行时不免费——每登记一个 projector，
写操作多一个同步故障点和延迟；`Apply` 循环内投影间无事务（第二个失败时第一个已刷完），
靠 `Refresh` 幂等 + 重放补救。同步投影对当前规模是合理选择（强一致、无 MQ 运维）；
已规划的 **outbox（发件箱模式）** 升级路径：写一个 `outboxApplier` 实现
`listingProjectionApplier`（把 `HmdChange` 写 outbox 表、后台消费），组装处替换注入——
六个子 service 零改动。这个一方法宽的消费侧接口等于提前为升级修好了插座。

### 字段涟漪：隔离的房租

"加一个字段整个链路都要动"是分层架构最著名的痛，**OCP 管不了它**（接口封闭的是
行为的新增实现，字段是数据形状变化，天生垂直穿透所有层）。要区分两种字段：

- **业务规则关心的字段**：每层"过一遍"都是真实决策（校验规则、各端投影形态），
  涟漪正当——十个地方在回答十个不同的问题；
- **纯透传字段**：十处改动是纯仪式，才是该省的。

省法按激进程度：
1. **审计每条边界是否"挣到了它的拷贝"**：两层结构体字段一一对应、mapper 纯抄写
   → 边界上形状没分化，可共用类型砍掉一次拷贝。HMD↔三端 HPD 的分化是真实的
   （拷贝挣到了）；handler request → service input → domain dto 一段值得审；
2. **公共字段打包**：透传字段聚成子结构体整体嵌入（已有 `commonmodel.CommonFields`
   先例），加字段只改一处定义；
3. **认下剩余成本，交给 AI 干**：字段涟漪机械、模式明确、编译器兜底，恰是 CC 干得
   最好的活。AI 时代重隔离架构更划算——仪式成本被压掉，保护一分没少。

真风险不是打字累，是**静默丢字段**：十个 mapper 有一个忘抄，Go 不报错，字段无声
丢在半路。防法：mapper 全字段对拍测试（`*_mapper_test.go` 已在做，加字段时同步补
断言）、`exhaustruct` lint。

一句话：**字段涟漪 = 隔离的房租。真实分化的层租金该付（可让 AI 代付）；
形状不分化的层，退租。**
