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
type msg struct {
	name   string
	args   interface{}
	sender ISender
}

// 本地服务，处理Requester的请求，通知支持ResponseHandler
type LocalService struct {
	RequestHandler
	ResponseHandler
	tickHandle func(tick int32)
	tick       int32
}

// 初始化
func (s *LocalService) Init() {
	s.RequestHandler.Init(true)
	s.ResponseHandler.Init(false)
	// 绑定同一个Channel
	s.ResponseHandler.BindChannel(s.RequestHandler.channel)
}

// 关闭
func (s *LocalService) Close() {
	s.RequestHandler.Close()
	s.ResponseHandler.Close()
}

// 设置定时器处理
func (s *LocalService) SetTickHandle(h func(tick int32), tick int32) {
	s.tickHandle = h
	if tick <= 0 {
		tick = SERVICE_TICK_MS
	}
	s.tick = tick
}

// 循环处理请求
func (s *LocalService) Run() error {
	var err error
	if s.tickHandle != nil {
		err = s.runProcessMsgAndTick()
	} else {
		err = s.runProcessMsg()
	}
	return err
}

// 循环处理消息和定时器
func (s *LocalService) runProcessMsgAndTick() error {
	ticker := time.NewTicker(time.Duration(time.Millisecond * time.Duration(s.tick)))
	lastTime := time.Now()
	channel := s.RequestHandler.channel
	run := true
	for run {
		select {
		case r, o := <-channel.ch:
			if !o {
				return ErrClosed
			}
			s.processMsg(r)
		case <-ticker.C:
			now := time.Now()
			tick := now.Sub(lastTime).Milliseconds()
			s.tickHandle(int32(tick))
			lastTime = now
		case <-channel.chClose:
			run = false
		}
	}
	return nil
}

// 循环处理请求
func (s *LocalService) runProcessMsg() error {
	channel := s.RequestHandler.channel
	run := true
	for run {
		select {
		case r, o := <-channel.ch:
			if !o {
				return ErrClosed
			}
			s.processMsg(r)
		case <-channel.chClose:
			run = false
		}
	}
	return nil
}

// 处理消息，包括请求和返回的结果
func (s *LocalService) processMsg(r *msg) {
	// 处理外部请求
	if s.RequestHandler.handleReq(r.sender, r.name, r.args) {
		return
	}
	// 遍历内部IRequester处理返回结果
	s.ResponseHandler.handleResp(r)
}
