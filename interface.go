package gproc

// 处理请求或者回调接口
type IHandler interface {
	Handle(peer IReceiver, name string, args interface{}) bool
}

// 接收者接口（内部接口）
type IReceiver interface {
	Receive(reqPeer IReceiver, reqName string, arg interface{}) error
}
