package gproc

const (
	CHANNEL_LENGTH = 100
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
		chanLen = CHANNEL_LENGTH
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
func (h *handler) Send(sender ISender, reqName string, args interface{}) error {
	// 已关闭，防止重复close造成panic
	if h.closed {
		return ErrClosed
	}
	// 请求写入
	m := &msg{
		name:   reqName,
		args:   args,
		sender: sender,
	}
	h.ch <- m
	return nil
}

// 请求消息处理器
type RequestHandler struct {
	handler   *handler
	handleMap map[string]func(sender ISender, args interface{})
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
}

// 默认初始化
func (h *RequestHandler) InitDefault() {
	h.Init(newDefaultHandler())
}

// 关闭
func (h *RequestHandler) Close() {
	h.handler.Close()
}

// 注册
func (h *RequestHandler) RegisterHandle(reqName string, handle func(ISender, interface{})) {
	h.handleMap[reqName] = handle
}

// 接收消息，实际等于Channel发送消息
func (h *RequestHandler) Recv(sender ISender, msgName string, msgArgs interface{}) error {
	return h.handler.Send(sender, msgName, msgArgs)
}

// 处理接收的消息
func (h *RequestHandler) Run() error {
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
			h.handleReq(m.sender, m.name, m.args)
		case <-h.handler.chClose:
			h.handler.closed = true
			loop = false
		}
	}
	return nil
}

// 处理单个IRequester请求后的回调
func (h *RequestHandler) handleReq(sender ISender, reqName string, args interface{}) bool {
	handle, o := h.handleMap[reqName]
	if !o {
		return false
	}
	handle(sender, args)
	return true
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
func (h *ResponseHandler) AddRequester(req IRequester) {
	h.requesterMap[req] = struct{}{}
}

// 发送
func (h *ResponseHandler) Send(msgName string, msgArgs interface{}) error {
	return h.handler.Send(nil, msgName, msgArgs)
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
func (r *ResponseHandler) handleResp(resp *msg) {
	for k := range r.requesterMap {
		if k.handle(resp.name, resp.args) {
			break
		}
	}
}
