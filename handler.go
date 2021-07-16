package gproc

// 消息处理器
type Handler struct {
	channel *Channel
}

// 创建消息处理器
func (h *Handler) Init(createChannel bool) {
	if createChannel {
		h.channel = NewChannel(CHANNEL_LENGTH)
	}
}

// 关闭
func (h *Handler) Close() {
	h.channel.Close()
}

// 绑定Channel
func (h *Handler) BindChannel(channel *Channel) {
	h.channel = channel
}

// 是否关闭
func (h *Handler) IsClosed() bool {
	return h.channel.closed
}

// 请求消息处理器
type RequestHandler struct {
	Handler
	handleMap map[string]func(sender ISender, args interface{})
}

// 创建RequestHandler
func NewRequestHandler(createChannel bool) *RequestHandler {
	handler := &RequestHandler{}
	handler.Init(createChannel)
	return handler
}

// 初始化
func (h *RequestHandler) Init(createChannel bool) {
	if createChannel {
		h.channel = NewChannel(CHANNEL_LENGTH)
	}
	h.handleMap = make(map[string]func(sender ISender, args interface{}))
}

// 注册
func (h *RequestHandler) RegisterHandle(reqName string, handle func(ISender, interface{})) {
	h.handleMap[reqName] = handle
}

// 接收消息，实际等于Channel发送消息
func (h *RequestHandler) Recv(sender ISender, msgName string, msgArgs interface{}) error {
	return h.channel.Send(sender, msgName, msgArgs)
}

// 处理接收的消息
func (h *RequestHandler) Update() error {
	if h.channel.closed {
		return ErrClosed
	}
	m, err := h.channel.Recv()
	if err != nil {
		return err
	}
	if m != nil {
		h.handleReq(m.sender, m.name, m.args)
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
	Handler
	requesterMap map[IRequester]struct{}
}

// 创建ResponseHandler
func NewResponseHandler(createChannel bool) *ResponseHandler {
	handler := &ResponseHandler{}
	handler.Init(createChannel)
	return handler
}

// 初始化
func (h *ResponseHandler) Init(createChannel bool) {
	if createChannel {
		h.channel = NewChannel(CHANNEL_LENGTH)
	}
	h.requesterMap = make(map[IRequester]struct{})
}

// 添加请求者
func (r *ResponseHandler) AddRequester(req IRequester) {
	r.requesterMap[req] = struct{}{}
}

// 发送
func (r *ResponseHandler) Send(msgName string, msgArgs interface{}) error {
	return r.channel.Send(nil, msgName, msgArgs)
}

// 更新处理IRequester的回调
func (r *ResponseHandler) Update() error {
	if r.channel.closed {
		return ErrClosed
	}
	resp, err := r.channel.Recv()
	if err != nil {
		return err
	}
	// 处理请求
	if resp != nil {
		r.handleResp(resp)
	}
	return nil
}

// 处理返回
func (r *ResponseHandler) handleResp(resp *msg) {
	for k, _ := range r.requesterMap {
		if k.handle(resp.name, resp.args) {
			break
		}
	}
}
