package gproc

// 发送者接口
type ISender interface {
	// 发送普通消息
	Send(msgName string, args interface{}) error
	// 转发消息
	forward(fromSender ISender, fromKey interface{}, msgName string, args interface{}) error
}

// 请求者接口
type IRequester interface {
	// 请求
	Request(msgName string, arg interface{}) error
	// 请求转发
	RequestForward(toKey interface{}, msgName string, arg interface{}) error
	// 注册请求回调
	RegisterCallback(msgName string, callback func(interface{}))
	// 注册通知处理器
	RegisterNotify(msgName string, handler func(interface{}))
	// 注册转发处理器
	RegisterForward(msgName string, handle func(fromKey interface{}, args interface{}))
	// 处理返回
	handle(m *msg) bool
}

// 请求消息处理器
type IRequestHandler interface {
	// 注册处理函数
	RegisterHandle(msgName string, handle func(ISender, interface{}))
	// 注册无法找到目标的转发处理器
	RegisterForward4NoTarget(msgName string, handle func(sender ISender, toKey interface{}, args interface{}))
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
