package gproc

import (
	"fmt"
)

// 请求者，发起请求到IReceiver，除创建初始化外整个生命周期在同一个goroutine中
// 一般跟IRequestHandler不在同一个goroutine
type Requester struct {
	owner       IResponseHandler // Requester的持有者
	receiver    IRequestHandler  // Requester请求的接收者
	req2RespMap map[string]string
	callbackMap map[string]func(interface{}) // 之所以不用线程安全的sync.Map，是因为callbackMap初始化时还未开始执行handle
	options     RequestOptions
}

// 创建请求者
func NewRequester(owner IResponseHandler, receiver IRequestHandler, options ...RequestOption) IRequester {
	if owner == nil || receiver == nil {
		panic("owner or receiver is nil")
	}
	req := &Requester{
		owner:       owner,
		receiver:    receiver,
		req2RespMap: make(map[string]string),
		callbackMap: make(map[string]func(interface{})),
	}
	owner.AddRequester(req)
	for _, option := range options {
		option(req.options)
	}
	return req
}

// 发送
func (r *Requester) Send(msgName string, msgArgs interface{}) error {
	// 相当于RequestHandler接收消息
	return r.receiver.Recv(nil, msgName, msgArgs)
}

// 请求，需要加上计时器处理超时（计时器用时间堆来实现）
func (r *Requester) Request(reqName string, arg interface{}) error {
	if _, o := r.req2RespMap[reqName]; !o {
		return fmt.Errorf("gproc: no request %s map to response", reqName)
	}
	// 相当于RequestHandler接收消息
	err := r.receiver.Recv(r.owner, reqName, arg)
	return err
}

// 注册回调
func (r *Requester) RegisterCallback(reqName string, respName string, callback func(interface{})) {
	r.req2RespMap[reqName] = respName
	r.callbackMap[respName] = callback
}

// 请求带回调，这个方法肯定在handle函数同一goroutine中使用，不存在callbackMap线程安全问题
func (r *Requester) RequestWithCallback(reqName string, arg interface{}, respName string, callback func(interface{})) error {
	err := r.Request(reqName, arg)
	if err == nil {
		r.req2RespMap[reqName] = respName
		r.callbackMap[respName] = callback
	}
	return err
}

// 处理回调
func (r *Requester) handle(respName string, arg interface{}) bool {
	callback, o := r.callbackMap[respName]
	if !o {
		return false
	}
	callback(arg)
	return true
}
