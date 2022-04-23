package gproc

// 发送者接口
type ISender interface {
	// 发送普通消息
	Send(msgId uint32, args interface{}) error
	// 转发消息
	forward(fromSender ISender, fromKey interface{}, msgId uint32, args interface{}) error
}

// 请求者接口
type IRequester interface {
	// 请求
	Request(msgId uint32, args interface{}) error
	// 请求转发
	RequestForward(toKey interface{}, msgId uint32, args interface{}) error
	// 注册请求回调
	RegisterCallback(msgId uint32, callback func(interface{}))
	// 注册通知处理器
	RegisterNotify(msgId uint32, handler func(interface{}))
	// 注册转发处理器
	RegisterForward(msgId uint32, handle func(fromKey interface{}, args interface{}))
	// 处理返回
	handle(m *msg) bool
}

// 请求消息处理器
type IRequestHandler interface {
	// 注册处理函数
	RegisterHandle(msgId uint32, handle func(ISender, interface{}))
	// 注册无法找到目标的转发处理器
	RegisterForward4NoTarget(msgId uint32, handle func(sender ISender, toKey interface{}, args interface{}))
	// 运行
	Run() error
	// 接收IRequester发来的数据
	recv(m *msg) error
}

// 返回消息处理器
type IResponseHandler interface {
	ISender
	// 创建请求者
	CreateRequester(receiver IRequestHandler, key interface{}, options ...RequestOption) IRequester
	// 更新
	Update() error
	// 添加请求者
	addRequester(req IRequester)
}
