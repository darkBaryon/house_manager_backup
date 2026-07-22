# 第 10 章 ISP：接口隔离原则

## 这章讲什么

接口隔离原则：**不要强迫调用方依赖它用不到的东西。**

书里的例子：三个用户各自只用 `OPS` 类的一个方法（op1/op2/op3）。虽然 User1 只调
op1，但在静态类型语言里，op2 的实现一改，User1 也得跟着重新编译、重新部署——
它被自己根本不用的代码牵连了。解法是拆成三个小接口，每个用户只依赖自己那份。

Martin 把它推广到架构层：不只是类的接口，**任何依赖都不该超过实际需要**。
系统 S 为了一个小功能引入了框架 F，F 又绑死了数据库 D——于是 D 的每次升级故障
都能波及 S，尽管 S 压根不关心 D。结论一句话：**依赖了不需要的东西，就背上了
不需要的风险。**

对 Go 读者有个彩蛋：这章特别指出 ISP 的痛感与语言相关，动态语言里问题小得多。
而 Go 的接口是隐式实现、按调用方声明的，等于把 ISP 内置进了语言习惯——
Go 谚语"接口属于消费者，不属于实现者"就是 ISP。

## 对照本项目

**1. `AIResponder` 是教科书级的 ISP。**
`pythonchat.Client` 作为一个 HTTP client，实际拥有的能力不止"回复"（配置、
超时管理、token 处理……）。但 `service/chat` 声明的端口只有一个方法：

```go
type AIResponder interface {
    Respond(ctx context.Context, input AIRespondInput) (AIRespondOutput, error)
}
```

接口定义在**消费者包**（`service/chat`）而不是实现方包里，宽度正好等于消费者的
需要。`pythonchat` 内部怎么重构、加多少方法，只要 `Respond` 不变，用例层就毫无
感知。如果反过来把接口定义在 `integration/pythonchat` 里、把 client 的全部方法
都塞进去，`service/chat` 就得依赖一堆它不用的东西——那是 Java 时代的常见错法。

**2. repository 按模块切分。**
`internal/repository/` 下是 `chat`、`hpd`、`hmd`、`favorite`、`history` 等十来个
小包，而不是一个巨大的 `Repository` 接口。收藏服务只依赖 `repository/favorite`，
足迹逻辑改动与它无关。如果做成一个全能 `Store` 接口，每个 service 都被迫依赖
全部数据操作，任何一处改动全体重编——正是书里 OPS 反例的规模化版本。

**3. 内部工具接口只暴露两个端点。**
Python 服务调 Go 只有 `house/search` 和 `house/public_detail` 两个 internal
tools。Go 明明还有几十个能力，但只按 Python 的实际需要开口子，且约定"只返回
确定性业务事实"。跨服务边界上的 ISP 就是：**API 面越窄，对方背的风险越小。**

## 一个容易犯的错

预防性地定义"将来可能用得上"的大接口。ISP 的方向恰恰相反：接口应该由已经存在的
调用需求**倒逼**出来，一个消费者一个视角。Go 里的实操习惯：先写具体类型，等出现
第二个实现或测试需要时，再在消费者侧提炼最小接口——本项目 chat 模块就是这么长的。

## 思考题

1. ~~`service/chat` 依赖 `*repochat.SessionRepository` 具体类型而非接口~~
   （此题前提有误，见第 11 章更正：service.go 内已有消费侧窄接口。）
2. 如果 Python 侧提出"还想要房东联系方式"，你是扩宽现有 internal tool 的返回，
   还是加新端点？用 ISP 分析两种做法各让谁背了什么。
3. 数 `AIRespondInput` 的字段：有没有 Python 实际不消费的字段？多传的每个字段
   都是一条隐形依赖。

---

## 讨论沉淀（2026-07）

### `publishaccess` 包是干什么的（讨论中补的背景）

access = 访问权限。回答"当前登录的房东有权看到/操作哪些房源"，即发房端多房东
之间的数据隔离。机制是一张根级归属关系表（`HpdRootScopeRelation`）：

1. **写入归属**：房东创建根实体（集中式项目/分散式小区）时，
   `UpsertRootScopeForPrincipal` 记录"这个 Principal（Redis session 里的结构化
   登录身份）拥有这个根"；
2. **查询过滤**：拉列表先 `ListAccessibleProjectIDs` 查名下项目 ID 再过滤；
3. **操作校验**：改楼栋/房间前 `CanAccessProjectForPrincipal` 验"这楼是你的吗"
   （即各子 service 开头的 `requireProjectAccess`）。

设计亮点：**权限只挂根实体**，楼栋/房型/房间不单独记权限，顺着树找到根再验——
权限表只有根级条目，量小且不随房间数膨胀。
（注意：这是鉴权代码，炸了是"房东看到别人房源"级别的安全事故，属于必须亲自
验收的承重墙。）

### 五方法共享接口的裁决：轻微违例，不拆

`publishAccessService`（5 方法）被六个子 service 共用；楼栋 service 只用得到
project 系方法，却依赖了 community 系——按 ISP 严格标准是轻微违例。它的实际代价
不在运行时，在**测试**：mock 这个接口要把五个方法都实现一遍。ISP 事故在 Go 里最
常见的表现形式：**不是编译慢，是 mock 肥**。

裁决不拆的理由：五个方法语义高度内聚（全是"根级权限"）、变化节奏一致，拆的收益
只是每个 mock 少两三个空方法。同文件里 domain 接口按消费者拆成了十几个
（三十多个方法不拆会爆炸）、access 共享（五个方法拆了嫌碎）——粒度按成本定，
不一致不是毛病。若将来要拆，正确切法是按集中式/分散式两组
（`projectAccess` / `communityAccess`），对齐消费者的真实需要。

**拆分信号两条：mock 开始肥、或两组方法开始因不同理由变化。信号没亮，别拆。**
这条规则适用于一切原则驱动的重构：原则告诉你往哪拆，成本信号告诉你何时拆。

### lint 补课

lint（静态检查）= 不运行代码、只"读"代码挑毛病的工具，管编译器不管的事。
本项目 CI 已有 `go vet` / `gofmt -l`。社区标配 **golangci-lint**（聚合器），
相关检查器：

- `exhaustruct`：构造指定结构体漏填字段即报——治"mapper 忘抄新字段"
  "refreshFuncs 漏填函数"这类编译放行、上线才炸的问题；
- `depguard` / `go-arch-lint`：把 README 里"service/chat 不许 import pythonchat"
  这类架构军规写成机器规则，谁违反（包括未来的 CC）CI 就挂。

定位：测试验证"行为对不对"，lint 验证"写法有没有已知坑"，**架构约束也能 lint 化**。
写在文档里的架构约束会腐烂，写进工具里的才能活下来——写代码的越来越多是 AI 时，
机器可执行的约束比注释里的君子协定值钱得多。
