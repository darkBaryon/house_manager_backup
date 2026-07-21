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

1. `service/chat` 依赖 `*repochat.SessionRepository` 具体类型而非接口。按 ISP
   要不要抽？什么时机抽？（提示：现在有几个消费者、几个实现？）
2. 如果 Python 侧提出"还想要房东联系方式"，你是扩宽现有 internal tool 的返回，
   还是加新端点？用 ISP 分析两种做法各让谁背了什么。
3. 数 `AIRespondInput` 的字段：有没有 Python 实际不消费的字段？多传的每个字段
   都是一条隐形依赖。
