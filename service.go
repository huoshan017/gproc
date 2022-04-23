package gproc

import (
	"time"
)

// 本地服务，处理Requester的请求作为RequestHandler，又可作为Requester的持有者ResponseHandler发送请求
type LocalService struct {
	handler         *handler
	requestHandler  *RequestHandler
	responseHandler *ResponseHandler
}

// 创建本地服务
func NewLocalService(chanLen int32) *LocalService {
	service := &LocalService{}
	service.Init(chanLen)
	return service
}

// 创建默认本地服务
func NewDefaultLocalService() *LocalService {
	return NewLocalService(ChannelLength)
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
	s.Init(ChannelLength)
}

// 关闭
func (s *LocalService) Close() {
	s.requestHandler.Close()
	s.responseHandler.Close()
}

// 设置定时器处理
func (s *LocalService) SetTickHandle(h func(tick time.Duration), tick time.Duration) {
	s.requestHandler.SetTickHandle(h, tick)
}

// 注册请求处理器
func (s *LocalService) RegisterHandle(msgId uint32, handle func(ISender, interface{})) {
	s.requestHandler.RegisterHandle(msgId, handle)
}

// 注册无法找到目标的转发处理器
func (s *LocalService) RegisterForward4NoTarget(msgId uint32, handle func(ISender, interface{}, interface{})) {
	s.requestHandler.RegisterForward4NoTarget(msgId, handle)
}

// 通知
func (s *LocalService) Notify(toKey interface{}, msgId uint32, args interface{}) error {
	return s.requestHandler.Notify(toKey, msgId, args)
}

// 创建请求者
func (s *LocalService) NewRequester(receiver IRequestHandler, key interface{}, options ...RequestOption) IRequester {
	return NewRequester(s.responseHandler, receiver, key, options...)
}

// 循环处理请求
func (s *LocalService) Run() error {
	var err error
	if s.requestHandler.tickHandle != nil {
		err = s.runProcessMsgAndTick()
	} else {
		err = s.runProcessMsg()
	}
	return err
}

// 接收消息
func (s *LocalService) recv(m *msg) error {
	return s.requestHandler.recv(m)
}

// 循环处理消息和定时器
func (s *LocalService) runProcessMsgAndTick() error {
	ticker := time.NewTicker(time.Duration(time.Millisecond * time.Duration(s.requestHandler.tick)))
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
			s.requestHandler.tickHandle(tick)
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
	if !s.requestHandler.handleMsg(r) {
		// 遍历内部IRequester处理返回结果
		s.responseHandler.handleResp(r)
	}
}
