package gproc

// 发送者接口
type ISender interface {
	Send(msgName string, msgArgs interface{}) error
}

// 请求者接口
type IRequester interface {
	Request(reqName string, arg interface{}) error
	RegisterCallback(reqName string, respName string, callback func(interface{}))
	Handle(respName string, arg interface{}) bool
}

// 请求消息处理器
type IRequestHandler interface {
	RegisterHandle(reqName string, handle func(ISender, interface{}))
	Recv(sender ISender, msgName string, msgArgs interface{}) error
}

// 返回消息处理器
type IResponseHandler interface {
	ISender
	AddRequester(req IRequester)
}
