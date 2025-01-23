package tinyactor

// PID 表示Actor进程标识
type PID struct {
	id string
}

// Actor 接口定义Actor行为
type Actor interface {
	Receive(context Context)
}

// Context 提供Actor间通信上下文
type Context interface {
	Self() PID
	Sender() PID
	Send(to PID, message interface{})
	Reply(message interface{})
	Ask(to PID, message interface{}) Future
}

// envelope 封装消息和发送者信息
type envelope struct {
	message interface{}
	sender  PID
}

// actorContext 实现Context接口
type actorContext struct {
	system        *Akka
	self          PID
	currentSender PID
	message       interface{}
}

// Context接口实现
func (ctx *actorContext) Self() PID {
	return ctx.self
}

func (ctx *actorContext) Sender() PID {
	return ctx.currentSender
}

func (ctx *actorContext) Send(to PID, message interface{}) {
	ctx.system.send(ctx.self, to, message)
}

func (ctx *actorContext) Reply(message interface{}) {
	ctx.system.send(ctx.self, ctx.currentSender, message)
}

func (ctx *actorContext) Ask(to PID, message interface{}) Future {
	return ctx.system.ask(ctx.self, to, message)
}
