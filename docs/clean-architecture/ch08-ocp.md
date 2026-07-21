# 第 8 章 OCP：开闭原则（精读版）

## 这章讲什么

开闭原则的口号是：软件实体应当**对扩展开放，对修改关闭**。也就是说，当需求变化时，
理想状态是通过"新增代码"来满足，而不是"改动已有代码"。这个说法最早由 Bertrand
Meyer 在 1988 年提出，比这本书早了近三十年。

这句话单独听很玄：不改代码怎么可能加功能？Martin 在这章给出了可操作的版本——OCP
的实现手段不是什么魔法，而是**依赖方向的设计**。衡量一个架构好坏的标准之一，就是
"一个需求变更需要改动多少已有代码"；好的架构把这个量压到最小，理想值是零。

## 书中案例：财务报表

全章围绕一个例子展开，值得完整过一遍，因为它就是后面"整洁架构"同心圆的雏形。

需求是这样的：一个系统在网页上展示财务摘要——数据可以滚动、负数标红。随后
利益相关方提出：同样的信息要能打印成黑白纸质报表——要分页、有页眉页脚、
负数用括号表示。

显然要写新代码，但除了新代码之外，**老代码要改多少**？

Martin 的做法分两步：

1. **先用 SRP 拆职责。**"计算报表数据"和"呈现报表"是两个不同的变更理由，
   拆成三块：生成报表数据的 *Interactor*（业务核心）、网页呈现的 *Presenter/View*、
   打印呈现的 *Presenter/View*。
2. **再用 DIP 组织依赖方向。**在组件边界上放接口，让所有依赖箭头都**指向
   Interactor**：Controller 依赖 Interactor 的入口接口；两个 Presenter 依赖
   Interactor 的输出边界接口；View 依赖 Presenter。

于是新增"导出 PDF"这种呈现方式，只是加一组新的 Presenter/View 实现，
Interactor 一行都不用动。这就是"对扩展开放（随便加呈现方式），对修改关闭
（业务核心不动）"。

### 关键洞见：保护层级

这章最值得背下来的一句话是：

> 如果 A 组件不想被 B 组件的变化波及，就让 B 依赖 A。

依赖箭头指向谁，谁就是被保护的一方。由此推出一个**保护层级**（protection
hierarchy）：越高层的策略越该被保护，越低层的细节越可以"牺牲"。Interactor
承载业务规则，是最高层策略，所以它处在所有依赖箭头的终点，被保护得最好；
View 是最低层的细节，谁都不依赖它，改起来最随意。

"层级"的判断标准不是调用顺序，而是**离输入输出的距离**：离 IO 越远、越接近
"这个系统为什么存在"的核心逻辑，层级越高。这其实已经预告了全书的同心圆架构图
——OCP 是那张图背后的驱动力之一。

### 两个容易被略过的细节

**方向控制（directional control）。**案例的类图里之所以出现那么多接口，
唯一目的就是把某些依赖箭头"掰"成反方向。没有接口时，调用方自然依赖被调用方；
插入接口后，被调用方反过来依赖（实现）调用方声明的抽象。接口不是仪式，
是转向器。

**信息隐藏（information hiding）。**Controller 访问 Interactor 时要经过一个
`FinancialReportRequester` 接口，这个接口的作用不是转依赖方向（方向本来就对），
而是**限制 Controller 能知道多少**——防止 Controller 传递性地依赖 Interactor 的
内部细节。没有它，Controller 的重编译/重部署会被 Interactor 的内部改动波及。
换句话说：接口既管"箭头指向哪"，也管"透过箭头能看见多少"。

## 对照本项目

### 1. AIResponder：一次完整的"箭头掰转"

最标准的例子是 AI chat 的回复方。看 `internal/service/chat/ports.go`：

```go
type AIResponder interface {
    Respond(ctx context.Context, input AIRespondInput) (AIRespondOutput, error)
}
```

三个事实拼起来，正好是书里那张图：

- **接口声明在用例侧**，不在实现侧。`service/chat`（Interactor）自己声明
  "我需要一个能回复的东西"，输入输出类型（`AIRespondInput/Output`）也定义在
  `service/chat/types.go`——这是输出边界接口的标准姿势。
- **实现在外圈**。`integration/pythonchat/client.go` 里有一行
  `var _ chatservice.AIResponder = (*Client)(nil)`，编译期断言自己实现了内圈的
  接口。依赖箭头：`integration/pythonchat -> service/chat`，指向被保护方。
- **组装在 Main 组件**。`wire/providers_chat.go` 的 `newChatAIResponder`
  把 `pythonchat.Client` 装进 `chatsvc.AIResponder`，再由 `newChatService`
  注入用例。只有 wire 层同时知道抽象和实现。

假设哪天把 Python 服务换成直接调 Claude API，或者加一个压测用的假 AI：新增一个
实现 `AIResponder` 的类型，改 `wire/providers_chat.go` 一处注入。`service/chat`
的会话管理、消息落库、超时控制——一行都不用改，它们的测试也一行不用改。

反过来想更清楚：如果当初 `service/chat` 直接 import 了 `pythonchat`，那么换 AI
供应商就要动用例代码，改动会波及会话逻辑的测试和稳定性。依赖箭头一转，
波及就被挡在了圈外。

### 2. internaltools/house：消费方接口 = 信息隐藏

`internal/handler/internaltools/house/handler.go` 是 Python 服务反向调 Go 的入口，
它没有直接持有 `*housesvc.HouseService`，而是自己声明了一个窄接口：

```go
type houseService interface {
    Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error)
    GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}
```

`HouseService` 上还有别的方法，但这个 handler 只看得见这两个。这正是书里
`FinancialReportRequester` 的角色：不是为了转依赖方向（handler 依赖 service
方向本来就对），而是**信息隐藏**——handler 不会传递性地依赖 service 的其他
能力，service 加减无关方法不波及这个 handler 及其测试（`handler_test.go`
可以用一个小 fake 实现这个接口，不用拉起真 service）。

### 3. RouteRegistrar：对"新增路由"关闭修改

`internal/handler/router.go` 只有一个 6 行的接口：

```go
type RouteRegistrar interface {
    RegisterRoutes(rg *gin.RouterGroup)
}
```

每个 handler 实现它完成自注册；`wire/providers_registrars.go` 按分组
（公开 / 小程序登录态 / 管理端 / 发布端 / 内部工具）把 registrar 收集成切片。
新增一个端点时，改动集中在新 handler 文件和 registrar 分组的一行追加，
`internal/app` 的 Gin 骨架不需要动——应用骨架对"新增路由"这类扩展是关闭修改的。

### 4. 本项目的保护层级

把书里的"保护层级"套到本仓库，箭头全部指向内圈：

```text
被保护最弱  handler/v1、integration/pythonchat、repository/*   （细节，随时可换）
    ↓ 依赖
中间        service/*                                          （用例）
    ↓ 依赖
被保护最强  domain/*                                           （业务规则）

组装：wire/* 谁都认识，但谁也不认识它。
```

`domain/hmd`、`domain/hpd` 不 import 任何 handler/repository——它们是所有
箭头的终点，正如 Interactor 在案例里的位置。

## 一个容易犯的错

OCP 不是要你为每个类都预先抽接口。抽象是有成本的（间接层、导航负担），只应该
架在**预期会变**的轴上。本项目的选择很克制：`AIResponder` 抽了接口，因为
"AI 实现会换"是明确预期；而 `wire/providers_chat.go` 里 `SessionRepository`、
`MessageRepository` 直接用具体类型 `*repochat.XxxRepository` 注入，因为
"换掉 Mongo"并不在预期里。100% 的封闭既不可能也不划算——这章原文也承认，
架构师只能针对**最可能的变化**做封闭，这需要判断，也会猜错。猜错的代价是
不对称的：少抽一个接口，将来补上是机械劳动；多抽一堆用不上的接口，读代码的人
每天都在付利息。

## 思考题（附参考思路）

1. **如果要给 chat 加"多模型路由"（简单问题走便宜模型），你会改 `AIResponder`
   接口，还是在接口后面加一个路由实现？哪种符合 OCP？**

   加路由实现。写一个 `RoutingResponder`，它自己也实现 `AIResponder`，内部持有
   若干个 `AIResponder` 并按规则分发——这是组合模式，`service/chat` 和现有的
   `pythonchat.Client` 都不用动，只改 wire 里的一处注入。反之，如果往接口上加
   `RespondCheap()` 之类的方法，所有实现方和用例都要跟着改，正是 OCP 想避免的
   "改动波及"。判断口诀：**变化能不能被表达成"接口的又一个实现"？能，就别动接口。**

2. **`handler/internaltools/house` 是 Python 反向调 Go 的入口。它的存在保护了谁？
   依赖箭头怎么画？**

   保护了 `service/miniapp/house` 和它背后的领域逻辑。箭头是
   `internaltools/house（适配器，外圈）-> service/miniapp/house（用例，内圈）`，
   Python 侧的协议细节（路由路径、请求响应 JSON 形状、内部鉴权）全被挡在
   handler 这一层。Python 服务改契约，波及止步于这个 handler；反过来 Go 的
   房源搜索逻辑改内部实现，Python 侧无感。另外注意它的 `houseService` 窄接口
   还叠加了一层信息隐藏（见上文第 2 节）。

3. **找一处你觉得"下次需求一来必须改老代码"的地方，想想值不值得现在就抽接口。**

   开放题。一个候选：`wire/providers_registrars.go` 里每个分组的构造函数——
   每加一个 handler 都要改一行。但按这章的标准，这**不值得**再抽象：改动发生在
   Main 组件（最外圈、本来就最脏的地方），一行追加不波及任何业务代码，为它上
   反射式自动注册反而把依赖关系变得不可静态追踪。OCP 保护的是高层策略，
   不是消灭一切改动。
