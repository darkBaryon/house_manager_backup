# 第 7 章 SRP：单一职责原则

## 这章讲什么

SRP 大概是 SOLID 里被误解最深的一条。多数人望文生义地理解成"一个函数/一个类只做
一件事"——那其实是更基础的编程常识，不是 SRP。Martin 在这章专门纠正这个误解，给出
的正式定义是：

> 任何一个软件模块，应该只对**一类行为者**（actor）负责。

"行为者"指的是会要求这个模块修改的一群人：产品经理、财务、DBA、运营……SRP 关心的
不是"这段代码干了几件事"，而是"有几拨人会因为不同的理由来改这段代码"。如果两拨
诉求不同的人都会改同一个模块，他们的修改就会互相踩踏——这才是 SRP 要防的事故。

书里的经典反例是一个 `Employee` 类，同时有三个方法：

- `calculatePay()` —— 财务部门要求改
- `reportHours()` —— 人力部门要求改
- `save()` —— DBA 要求改

三个方法共用一个私有的工时计算函数。某天财务要求改工时算法，改完后人力的报表
悄悄跟着变了——没人意识到，因为"共用代码"看起来是好事。SRP 说：为不同行为者服务
的代码必须**拆开**，哪怕它们现在看起来长得一样。

解法是把三块逻辑拆成三个类，各自面对自己的行为者；如果嫌调用方要认识三个类太麻烦，
可以再加一个 Facade 统一入口。

## 对照本项目

本项目的目录结构本身就是按行为者切的，这一点比大多数项目做得干净：

```text
internal/handler/v1/miniapp/   —— 为小程序端（租客）负责
internal/handler/v1/publish/   —— 为发房端（房东）负责
internal/handler/v1/admin/     —— 为管理后台（运营/员工）负责
internal/service/miniapp/      —— 小程序端用例
internal/service/publish/      —— 发房端用例
internal/service/admin/        —— 后台用例
```

小程序、发房端、后台是三拨完全不同的行为者：租客要搜房，房东要发房，运营要管人。
如果把"查房源"写成一个所有端共用的 service，某天后台要求"列表里带上未上架房源"，
这个改动就可能顺手把小程序端的搜索结果污染了——正是书里 `Employee` 事故的翻版。
项目把它们拆成 `service/miniapp/house` 和后台自己的查询路径，就是 SRP 在起作用。

另一个例子是读写分离的 HMD / HPD：

```text
internal/domain/hmd/          —— 房东维护房源主数据（写侧，行为者：房东）
internal/repository/hpd/      —— 小程序读模型投影（读侧，行为者：租客搜索）
```

"房东怎么录入房源"和"租客搜索时看到什么"是两类完全不同的诉求，变化节奏也不同。
把它们做成两套模型（HMD 写入后投影成 HPD），两边可以各自演化，互不踩踏。最近的
提交历史（"清理 HMD 更新白名单"、"删除 HPD 投影死出口"）说明这两侧确实在被
不同理由驱动着独立修改——这正是当初拆开的回报。

## 一个容易犯的错

SRP 不是"越拆越碎越好"。判断标准始终是行为者：如果两段代码永远因为同一个理由、
被同一拨人要求修改，拆开它们只会增加导航成本。比如 `domain/hmd/` 里楼栋、房型、
房间放在同一个包里，因为它们都服务于"房东维护房源结构"这一件事，不必再拆。

## 思考题

1. `internal/service/chat/` 没有按端拆（没有 `service/miniapp/chat`），而是独立成包。
   它的行为者是谁？这样放合理吗？
2. 假如后台将来要"代房东修改房源"，你会让 admin service 直接调 `domain/hmd`，
   还是复制一份逻辑？用 SRP 的行为者视角论证。
3. `pkg/response`、`pkg/errcode` 被所有端共用，这违反 SRP 吗？（提示：想想它们的
   修改理由是谁提出的。）

---

## 讨论沉淀（2026-07）

以下是围绕本章的对话讨论中沉淀下来的内容，超出初版笔记的部分。

### Facade 模式与 `PublishService`

SRP 把逻辑拆碎后有个副作用：调用方要认识一堆小类。补救手段是 **Facade（门面）模式**：
拆归拆，对外再合成一个统一入口。`internal/service/publish/service.go` 用 Go 的
struct embedding 无意中实现了它：

```go
type PublishService struct {
    *centralizedProjectService
    *buildingService
    *roomTypeService
    ...
}
```

实现视角是六个各管一摊的小 service，调用方（handler）视角是一个大 service——
六个子 service 的方法通过 embedding 全部"透出"到门面上。SRP 拆分 + Facade 合拢
是配套动作。

### `chat.Service.Send` 为什么不违反 SRP

`Send` 是一个约 200 行的"事务脚本"：校验、session 生命周期、seq 分配、消息落库、
runtime context、调 AI、日志计时全在里面。看似"做了七件事"，但用行为者标准检验：
七件事全部服务于"小程序 chat 产品"一个行为者，需求一变整条链一起变——**按 SRP
不违规**。真正的问题是方法级内聚差、难测试，那是 Clean Code 层面的事，不是架构病。

对应的药也是方法级的：不拆包、不抽接口，把 200 行整理成"20 行目录 + 阶段私有方法"：

```go
func (s *Service) Send(ctx, input) (*SendResult, error) {
    req, err  := s.parseSendInput(ctx, input)
    sess, err := s.resolveActiveSession(ctx, req)
    userMsg, err := s.appendUserMessage(ctx, sess, req)
    aiOut, err := s.requestAIReply(ctx, sess, userMsg)
    return s.persistReply(ctx, sess, aiOut)
}
```

教训：**架构级的刀（拆包、抽接口）留给架构级的病**。用 SOLID 的名义做其实只是
代码美容的重构，是常见的浪费。

### errcode 事故推演（本章最重要的发现）

现状的依赖链：

- `pkg/errcode` 里 `AlreadyExists = New(10006, "资源已存在")`——10006 是前端联调
  契约，"资源已存在"是用户可见文案；
- `internal/domain/hmd/errors.go` import 并使用它：领域校验代码在**直接挑选**线上
  JSON 响应里的数字和文案（`pkg/response.Err` 只透传）；
- 全项目 `domain` 三个包 + `service` 下二十多个文件都是这个模式。

事故剧本：前端提出"错误码规范化，10006 改成六位数"——一个纯展示层需求，波及面
却覆盖 domain/service 层及其测试（`assertErrCode(t, err, errcode.AlreadyExists.Code)`
这类断言全挂）。两拨行为者（前端/产品 vs 业务规则维护者）共用了一个包。

解法（引信出现时再做，不必现在做）：**内层只命名错误，外层翻译码 + 文案**——

```go
// domain 层：只有名字
var ErrDuplicateBuilding = errors.New("duplicate building")

// 出口处：翻译表，可按端给不同文案/码
var hmdErrorMap = map[error]*errcode.Error{
    hmd.ErrDuplicateBuilding: errcode.AlreadyExists,
}
```

这与 `listingprojection` 三端各配 mapper 是同一个直觉：那边投影**数据**，
这边投影**错误**。实操细节：匹配用 `errors.Is`（错误会被 wrap）；翻译表要有兜底
（漏登记统一落 `InternalError`，不能把内部细节漏给前端）。

当前结论：一人项目、契约自己说了算，这个耦合实际成本≈0，**留着**；但要意识到
自己签了"`pkg/errcode` 必须像标准库一样稳定"的合同。架构决策可以推迟，
但不能是无意识的。

### chat 独立成包的定位

作者的理由："这是一整套 AI 控制流程，独立于本项目存在的功能模块"——这已经不是
SRP 的语言（谁会来改它），而是**组件**的语言（能否独立复用），对应第 12 章 REP
（复用发布等价原则）。诚实检验：把 `service/chat` + `repository/chat` +
`integration/pythonchat` 拷进另一个项目能编译吗？差一点——`model/chat` 算它自己的，
但 `pkg/errcode` 把它拴在本项目上。目前是"80% 独立"：边界画了，定位未兑现。
两条路都成立：兑现（挪 `pkg/` 或独立 module）或改口（按 CCP"变化节奏不同"留在原地）。
