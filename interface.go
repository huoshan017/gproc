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
