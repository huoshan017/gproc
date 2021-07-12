package gproc

// 接收者接口（内部接口）
type IReceiver interface {
	Receive(reqPeer IReceiver, reqName string, arg interface{}) error
}
