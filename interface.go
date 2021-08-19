package gproc

// 发送者接口
type ISender interface {
	// 发送
	Send(msgName string, msgArgs interface{}) error
}

// 请求者接口
type IRequester interface {
	// 请求
	Request(msgName string, arg interface{}) error
	// 注册请求回调
	RegisterCallback(msgName string, callback func(interface{}))
	// 带回调的请求
	RequestWithCallback(msgName string, arg interface{}, callback func(interface{})) error
	// 处理返回
	handle(msgName string, arg interface{}) bool
}

// 请求消息处理器
type IRequestHandler interface {
	// 注册处理函数
	RegisterHandle(msgName string, handle func(ISender, interface{}))
	// 接收IRequester发来的数据
	Recv(sender ISender, msgName string, msgArgs interface{}) error
	// 运行
	Run() error
}

// 返回消息处理器
type IResponseHandler interface {
	ISender
	// 添加请求者
	addRequester(req IRequester)
	// 更新
	Update() error
}
