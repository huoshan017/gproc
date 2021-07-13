package gproc

// 请求者接口
type IRequester interface {
	Request(reqName string, arg interface{}) error
	RegisterCallback(respName string, callback func(interface{}))
	handle(peer IReceiver, respName string, arg interface{}) bool
}

// 接收者接口
type IReceiver interface {
	Receive(reqPeer IReceiver, reqName string, arg interface{}) error
}

// 处理器接口
type IHandler interface {
	IReceiver
	AddRequester(r IRequester)
}

// 请求返回处理器
type IResponseHandler interface {
	IHandler
	Update() error
}

// 本地服务接口
type ILocalService interface {
	IHandler
	RegisterHandle(reqName string, h func(IReceiver, interface{}))
	SetTickHandle(h func(tick int32))
	Run() error
}
