package gproc

import (
	"errors"
	"time"
)

var ErrClosed = errors.New("gproc: closed service cant request")

const (
	SERVICE_TICK_MS = 10 * time.Millisecond // 定时器间隔
)

// 消息
type msg struct {
	name   string
	args   interface{}
	sender ISender
}

// 本地服务，处理Requester的请求，返回ResponseHandler
type LocalService struct {
	handler         *handler
	requestHandler  *RequestHandler
	responseHandler *ResponseHandler
	tickHandle      func(tick time.Duration)
	tick            time.Duration
}

// 创建本地服务
func NewLocalService(chanLen int32) *LocalService {
	service := &LocalService{}
	service.Init(chanLen)
	return service
}

// 创建默认本地服务
func NewDefaultLocalService() *LocalService {
	return NewLocalService(CHANNEL_LENGTH)
}

// 初始化
func (s *LocalService) Init(chanLen int32) {
	s.handler = &handler{}
	s.handler.Init(chanLen)
	s.requestHandler = NewRequestHandler(s.handler)
	s.responseHandler = NewResponseHandler(s.handler)
}

// 默认初始化
func (s *LocalService) InitDefault() {
	s.Init(CHANNEL_LENGTH)
}

// 关闭
func (s *LocalService) Close() {
	s.requestHandler.Close()
	s.responseHandler.Close()
}

// 设置定时器处理
func (s *LocalService) SetTickHandle(h func(tick time.Duration), tick time.Duration) {
	s.tickHandle = h
	if tick <= 0 {
		tick = SERVICE_TICK_MS
	}
	s.tick = tick
}

// 注册请求处理器
func (s *LocalService) RegisterHandle(reqName string, handle func(ISender, interface{})) {
	s.requestHandler.RegisterHandle(reqName, handle)
}

// 接收消息
func (s *LocalService) Recv(sender ISender, msgName string, msgArgs interface{}) error {
	return s.requestHandler.Recv(sender, msgName, msgArgs)
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
	run := true
	for run {
		select {
		case r, o := <-s.handler.ch:
			if !o {
				return ErrClosed
			}
			s.processMsg(r)
		case <-ticker.C:
			now := time.Now()
			tick := now.Sub(lastTime)
			s.tickHandle(tick)
			lastTime = now
		case <-s.handler.chClose:
			run = false
		}
	}
	return nil
}

// 循环处理请求
func (s *LocalService) runProcessMsg() error {
	run := true
	for run {
		select {
		case r, o := <-s.handler.ch:
			if !o {
				return ErrClosed
			}
			s.processMsg(r)
		case <-s.handler.chClose:
			run = false
		}
	}
	return nil
}

// 处理消息，包括请求和返回的结果
func (s *LocalService) processMsg(r *msg) {
	// 处理外部请求
	if s.requestHandler.handleReq(r.sender, r.name, r.args) {
		return
	}
	// 遍历内部IRequester处理返回结果
	s.responseHandler.handleResp(r)
}
