package gproc

import (
	"errors"
)

var ErrClosed = errors.New("gproc: closed service cant request")

// 请求
type req struct {
	name string
	args interface{}
	peer IReceiver
}

// 本地服务
type LocalService struct {
	ResponseHandler
	handle func(IReceiver, string, interface{}) bool
}

// 设置处理器
func (s *LocalService) SetHandle(f func(IReceiver, string, interface{}) bool) {
	s.handle = f
}

// 循环处理请求
func (s *LocalService) Run() error {
	run := true
	for run {
		select {
		case r, o := <-s.chReq:
			if !o {
				return errors.New("gproc: service already closed, break loop")
			}
			if s.handle == nil {
				panic("s.Handle is nil ")
			}
			// 处理请求
			if !s.handle(r.peer, r.name, r.args) {
				s.requesterMap.Range(func(key, _ interface{}) bool {
					req, o := key.(*Requester)
					if !o {
						return false
					}
					return req.handle(r.peer, r.name, r.args)
				})
			}
		case <-s.chClose:
			run = false
		}
	}
	return nil
}
