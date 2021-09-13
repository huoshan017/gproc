package gproc

import (
	"time"
)

const (
	ChannelLength       = 100                                  // 通道长度
	ServiceTickDuration = time.Duration(10 * time.Millisecond) // 定时器间隔
)

// 消息处理器
type handler struct {
	ch      chan *msg
	closed  bool
	chClose chan struct{}
}

// 新的处理器
func newHandler(chanLen int32) *handler {
	h := &handler{}
	h.Init(chanLen)
	return h
}

// 新的默认处理器
func newDefaultHandler() *handler {
	return newHandler(0)
}

// 创建消息处理器
func (h *handler) Init(chanLen int32) {
	if chanLen <= 0 {
		chanLen = ChannelLength
	}
	h.ch = make(chan *msg, chanLen)
	h.chClose = make(chan struct{})
}

// 关闭
func (h *handler) Close() {
	if h.closed {
		return
	}
	close(h.chClose)
	h.closed = true
}

// 是否关闭
func (h *handler) IsClosed() bool {
	return h.closed
}

// 内部发送函数
func (h *handler) Send(m *msg) error {
	// 已关闭，防止重复close造成panic
	if h.closed {
		return ErrClosed
	}
	h.ch <- m
	return nil
}

// 请求消息处理器
type RequestHandler struct {
	handler               *handler
	handleMap             map[string]func(sender ISender, args interface{})
	signUpMap             map[interface{}]ISender
	tickHandle            func(tick time.Duration)
	forwardNoTargetHandle map[string]func(sender ISender, toKey interface{}, args interface{})
	tick                  time.Duration
}

// 创建RequestHandler
func NewRequestHandler(handler *handler) *RequestHandler {
	h := &RequestHandler{}
	h.Init(handler)
	return h
}

// 创建默认的请求处理器
func NewDefaultRequestHandler() *RequestHandler {
	return NewRequestHandler(newDefaultHandler())
}

// 初始化
func (h *RequestHandler) Init(handler *handler) {
	h.handler = handler
	h.handleMap = make(map[string]func(sender ISender, args interface{}))
	h.signUpMap = make(map[interface{}]ISender)
	h.forwardNoTargetHandle = make(map[string]func(sender ISender, toKey interface{}, args interface{}))
}

// 默认初始化
func (h *RequestHandler) InitDefault() {
	h.Init(newDefaultHandler())
}

// 关闭
func (h *RequestHandler) Close() {
	h.handler.Close()
}

// 设置定时器处理
func (h *RequestHandler) SetTickHandle(handle func(tick time.Duration), tick time.Duration) {
	h.tickHandle = handle
	if tick <= 0 {
		tick = ServiceTickDuration
	}
	h.tick = tick
}

// 注册
func (h *RequestHandler) RegisterHandle(msg string, handle func(ISender, interface{})) {
	h.handleMap[msg] = handle
}

// 注册无目标转发时的处理器
func (h *RequestHandler) RegisterForward4NoTarget(msgName string, handle func(ISender, interface{}, interface{})) {

}

// 接收消息，实际等于Channel发送消息
func (h *RequestHandler) recv(m *msg) error {
	return h.handler.Send(m)
}

// 通知
func (h *RequestHandler) Notify(toKey interface{}, name string, args interface{}) error {
	s, o := h.signUpMap[toKey]
	if !o {
		return ErrNotFoundRequesterKey
	}
	return s.Send(name, args)
}

// 处理接收的消息
func (h *RequestHandler) Run() error {
	if h.handler.closed {
		return ErrClosed
	}

	var lastTime time.Time
	var ticker *time.Ticker
	if h.tick > 0 && h.tickHandle != nil {
		ticker = time.NewTicker(h.tick)
		lastTime = time.Now()
	}

	loop := true
	if ticker == nil {
		for loop {
			select {
			case m, o := <-h.handler.ch:
				if !o {
					return ErrClosed
				}
				h.handleMsg(m)
			case <-h.handler.chClose:
				h.handler.closed = true
				loop = false
			}
		}
	} else {
		for loop {
			select {
			case m, o := <-h.handler.ch:
				if !o {
					return ErrClosed
				}
				h.handleMsg(m)
			case <-ticker.C:
				now := time.Now()
				tick := now.Sub(lastTime)
				h.tickHandle(tick)
				lastTime = now
			case <-h.handler.chClose:
				h.handler.closed = true
				loop = false
			}
		}
	}
	return nil
}

func (h *RequestHandler) handleMsg(m *msg) bool {
	result := true
	switch m.typ {
	case msgNormal:
		result = h.handleReq(m.sender, m.name, m.args)
	case msgSignup:
		h.signUpMap[m.fromKey] = m.sender
	case msgForward:
		err := h.handleForward(m.fromKey, m.toKey, m.name, m.args)
		// todo 错误处理先放着
		if err != nil {
			result = false
		}
	default:
		result = false
	}
	putMsg(m)
	return result
}

// 处理单个IRequester请求后的回调
func (h *RequestHandler) handleReq(sender ISender, name string, args interface{}) bool {
	handle, o := h.handleMap[name]
	if !o {
		return false
	}
	handle(sender, args)
	return true
}

// 处理转发
func (h *RequestHandler) handleForward(fromKey, toKey interface{}, name string, args interface{}) error {
	s, o := h.signUpMap[fromKey]
	// 找不到请求者
	if !o {
		return ErrNotFoundRequesterKey
	}
	r, o := h.signUpMap[toKey]
	// 找不到toKey对应的目标，转到无目标的转发处理器
	if !o {
		handle, o := h.forwardNoTargetHandle[name]
		if !o {
			return ErrNotFoundNoTargetForwardHandle
		}
		handle(s, toKey, args)
		return nil
	}
	return r.forward(s, fromKey, name, args)
}

// 返回消息处理器
type ResponseHandler struct {
	handler      *handler
	requesterMap map[IRequester]struct{}
}

// 创建返回Handler
func NewResponseHandler(handler *handler) *ResponseHandler {
	h := &ResponseHandler{}
	h.Init(handler)
	return h
}

// 创建默认返回处理器
func NewDefaultResponseHandler() *ResponseHandler {
	return NewResponseHandler(newDefaultHandler())
}

// 初始化
func (h *ResponseHandler) Init(handler *handler) {
	h.handler = handler
	h.requesterMap = make(map[IRequester]struct{})
}

// 默认初始化
func (h *ResponseHandler) InitDefault() {
	h.Init(newDefaultHandler())
}

// 关闭
func (h *ResponseHandler) Close() {
	h.handler.Close()
}

// 添加请求者
func (h *ResponseHandler) addRequester(req IRequester) {
	h.requesterMap[req] = struct{}{}
}

// 发送
func (h *ResponseHandler) Send(name string, args interface{}) error {
	m := getMsg()
	m.typ = msgNormal
	m.name = name
	m.args = args
	return h.handler.Send(m)
}

// 转发消息
func (h *ResponseHandler) forward(fromSender ISender, fromKey interface{}, name string, args interface{}) error {
	m := getMsg()
	m.typ = msgForward
	m.name = name
	m.sender = fromSender
	m.fromKey = fromKey
	m.args = args
	return h.handler.Send(m)
}

// 更新处理IRequester的回调
func (h *ResponseHandler) Update() error {
	if h.handler.closed {
		return ErrClosed
	}
	loop := true
	for loop {
		select {
		case m, o := <-h.handler.ch:
			if !o {
				return ErrClosed
			}
			h.handleResp(m)
		case <-h.handler.chClose:
			h.handler.closed = true
			loop = false
		default:
			loop = false
		}
	}
	return nil
}

// 处理返回
func (r *ResponseHandler) handleResp(m *msg) {
	for k := range r.requesterMap {
		if k.handle(m) {
			break
		}
	}
	putMsg(m)
}
