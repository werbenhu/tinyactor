package tinyactor

import (
	"errors"
	"time"
)

// Future 异步操作结果
type Future interface {
	Result() (interface{}, error)
	Wait(timeout time.Duration) bool
	Cancel()
}

// defaultFuture 实现Future接口
type defaultFuture struct {
	result chan interface{}
	err    chan error
	done   chan struct{}
}

// Future接口实现
func (f *defaultFuture) Result() (interface{}, error) {
	select {
	case result := <-f.result:
		return result, nil
	case err := <-f.err:
		return nil, err
	case <-f.done:
		return nil, errors.New("future cancelled")
	}
}

func (f *defaultFuture) Wait(timeout time.Duration) bool {
	select {
	case <-f.result:
		return true
	case <-f.err:
		return true
	case <-time.After(timeout):
		return false
	case <-f.done:
		return false
	}
}

func (f *defaultFuture) Cancel() {
	close(f.done)
}
