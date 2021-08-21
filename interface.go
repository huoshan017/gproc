package gproc

// 发送者接口
type ISender interface {
	// sender对应的key
	//GetKey() interface{}
	// 发送普通消息
	Send(name string, args interface{}) error
	// 转发消息
	Forward(fromSender ISender, fromKey interface{}, name string, args interface{}) error
	// 通知
	//Notify(name string, args interface{}) error
}

// 请求者接口
type IRequester interface {
	// 请求
	Request(msgName string, arg interface{}) error
	// 注册请求回调
	RegisterCallback(msgName string, callback func(interface{}))
	// 注册通知处理器
	RegisterNotify(name string, handler func(interface{}))
	// 请求转发给
	RequestForward(toKey interface{}, name string, arg interface{}) error
	// 注册转发处理器
	RegisterForward(name string, handle func(sender ISender, fromKey interface{}, args interface{}))
	// 处理返回
	handle(m *msg) bool
}

// 请求消息处理器
type IRequestHandler interface {
	// 注册处理函数
	RegisterHandle(msgName string, handle func(ISender, interface{}))
	// 接收IRequester发来的数据
	recv(m *msg) error
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
