package gproc

// 发送者接口
type ISender interface {
	Send(msgName string, msgArgs interface{}) error
}

// 请求者接口
type IRequester interface {
	Request(msgName string, arg interface{}) error
	RegisterCallback(msgName string, callback func(interface{}))
	handle(msgName string, arg interface{}) bool
}

// 请求消息处理器
type IRequestHandler interface {
	RegisterHandle(msgName string, handle func(ISender, interface{}))
	Recv(sender ISender, msgName string, msgArgs interface{}) error
	Run() error
}

// 返回消息处理器
type IResponseHandler interface {
	ISender
	addRequester(req IRequester)
	Update() error
}
