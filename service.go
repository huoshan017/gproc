package gproc

import (
	"errors"
	"time"
)

var ErrClosed = errors.New("gproc: closed service cant request")

const (
	SERVICE_TICK_MS = 10 // 定时器间隔
)

// 请求
type req struct {
	name string
	args interface{}
	peer IReceiver
}

// 本地服务
type LocalService struct {
	ResponseHandler
	chClose    chan struct{}
	handleMap  map[string]func(peer IReceiver, args interface{})
	tickHandle func(tick int32)
}

// 初始化
func (s *LocalService) Init(reqLen int32) {
	s.ResponseHandler.Init(reqLen)
	s.chClose = make(chan struct{})
	s.handleMap = make(map[string]func(IReceiver, interface{}))
}

// 关闭
func (s *LocalService) Close() {
	if s.closed {
		return
	}
	close(s.chClose)
	s.closed = true
}

// 注册处理器
func (s *LocalService) RegisterHandle(reqName string, h func(IReceiver, interface{})) {
	s.handleMap[reqName] = h
}

// 设置定时器处理
func (s *LocalService) SetTickHandle(h func(tick int32)) {
	s.tickHandle = h
}

// 循环处理请求
func (s *LocalService) Run() error {
	var err error
	if s.tickHandle != nil {
		err = s.runProcessReqAndTick()
	} else {
		err = s.runProcessReq()
	}
	return err
}

// 循环处理请求和定时器
func (s *LocalService) runProcessReqAndTick() error {
	run := true
	ticker := time.NewTicker(time.Duration(time.Millisecond * SERVICE_TICK_MS))
	lastTime := time.Now()
	for run {
		select {
		case r, o := <-s.ch:
			if !o {
				return errors.New("gproc: service already closed, break loop")
			}
			s.processReq(r)
		case <-ticker.C:
			now := time.Now()
			tick := now.Sub(lastTime).Milliseconds()
			s.tickHandle(int32(tick))
			lastTime = now
		case <-s.chClose:
			run = false
		}
	}
	return nil
}

// 循环处理请求
func (s *LocalService) runProcessReq() error {
	run := true
	for run {
		select {
		case r, o := <-s.ch:
			if !o {
				return errors.New("gproc: service already closed, break loop")
			}
			s.processReq(r)
		case <-s.chClose:
			run = false
		}
	}
	return nil
}

// 处理请求或者返回的结果
func (s *LocalService) processReq(r *req) {
	// 处理请求
	if s.handle(r.peer, r.name, r.args) {
		return
	}
	// 遍历请求者处理回调
	s.requesterMap.Range(func(key, _ interface{}) bool {
		req, o := key.(*Requester)
		if !o {
			return false
		}
		return req.handle(r.peer, r.name, r.args)
	})
}

// 处理返回的结果
func (s *LocalService) handle(peer IReceiver, reqName string, args interface{}) bool {
	h, o := s.handleMap[reqName]
	if !o {
		return false
	}
	h(peer, args)
	return true
}
