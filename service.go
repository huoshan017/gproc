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
	handleMap map[string]func(peer IReceiver, args interface{})
}

// 初始化
func (s *LocalService) Init() {
	s.ResponseHandler.Init()
	s.handleMap = make(map[string]func(IReceiver, interface{}))
}

// 注册处理器
func (s *LocalService) RegisterHandle(reqName string, h func(IReceiver, interface{})) {
	s.handleMap[reqName] = h
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

// 处理
func (s *LocalService) handle(peer IReceiver, reqName string, args interface{}) bool {
	h, o := s.handleMap[reqName]
	if !o {
		return false
	}
	h(peer, args)
	return true
}
