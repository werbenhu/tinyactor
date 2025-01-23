package tinyactor

import (
	"errors"
	"fmt"
	"sync"
)

// Akka 管理Actor系统
type Akka struct {
	actors      map[PID]Actor
	namedActors map[string]PID
	mailboxes   map[PID]chan envelope
	lock        sync.RWMutex
}

// New 创建Actor系统
func New() *Akka {
	return &Akka{
		actors:      make(map[PID]Actor),
		namedActors: make(map[string]PID),
		mailboxes:   make(map[PID]chan envelope),
	}
}

// Spawn 创建匿名Actor
func (as *Akka) Spawn(actor Actor) PID {
	as.lock.Lock()
	defer as.lock.Unlock()

	pid := PID{id: fmt.Sprintf("actor-%d", len(as.actors))}
	return as.registerActor(pid, actor)
}

// SpawnNamed 创建命名Actor
func (as *Akka) SpawnNamed(name string, actor Actor) PID {
	as.lock.Lock()
	defer as.lock.Unlock()

	if _, exists := as.namedActors[name]; exists {
		panic(fmt.Sprintf("Actor with name %s already exists", name))
	}

	pid := PID{id: name}
	return as.registerActor(pid, actor)
}

// registerActor 注册Actor的内部方法
func (as *Akka) registerActor(pid PID, actor Actor) PID {
	as.actors[pid] = actor
	as.namedActors[pid.id] = pid
	as.mailboxes[pid] = make(chan envelope, 100)

	go as.run(pid)
	return pid
}

// GetActorByName 通过名字获取Actor
func (as *Akka) GetActorByName(name string) (PID, bool) {
	as.lock.RLock()
	defer as.lock.RUnlock()

	pid, exists := as.namedActors[name]
	return pid, exists
}

// runActor 处理Actor消息循环
func (as *Akka) run(pid PID) {
	mailbox := as.mailboxes[pid]
	actor := as.actors[pid]

	ctx := &actorContext{
		system: as,
		self:   pid,
	}

	for envelope := range mailbox {
		ctx.currentSender = envelope.sender
		ctx.message = envelope.message
		actor.Receive(ctx)
	}
}

// Stop 停止Actor
func (as *Akka) Stop(pid PID) {
	as.lock.Lock()
	defer as.lock.Unlock()

	if mailbox, exists := as.mailboxes[pid]; exists {
		close(mailbox)
		delete(as.mailboxes, pid)
		delete(as.actors, pid)
		delete(as.namedActors, pid.id)
	}
}

// send 发送消息
func (as *Akka) send(from, to PID, message interface{}) {
	as.lock.RLock()
	defer as.lock.RUnlock()

	mailbox, exists := as.mailboxes[to]
	if !exists {
		return
	}

	mailbox <- envelope{
		message: message,
		sender:  from,
	}
}

// ask 同步请求-响应模式
func (as *Akka) ask(from, to PID, message interface{}) Future {
	f := &defaultFuture{
		result: make(chan interface{}, 1),
		err:    make(chan error, 1),
		done:   make(chan struct{}),
	}

	go func() {
		as.lock.RLock()
		mailbox, exists := as.mailboxes[to]
		as.lock.RUnlock()

		if !exists {
			f.err <- fmt.Errorf("actor %v does not exist", to)
			return
		}

		select {
		case mailbox <- envelope{message: message, sender: from}:
		case <-f.done:
			f.err <- errors.New("future cancelled")
		}
	}()
	return f
}
