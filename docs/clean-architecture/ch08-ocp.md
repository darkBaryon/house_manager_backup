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
