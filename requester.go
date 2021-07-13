package gproc

// 请求者，一般跟IReceiver不在同一个goroutine
type Requester struct {
	receiver    IReceiver
	callbackMap map[string]func(interface{})
	owner       IReceiver
}

// 创建请求者
func NewRequester(handler IHandler, receiver IReceiver) IRequester {
	if handler == nil || receiver == nil {
		panic("handler or receiver is nil")
	}
	req := &Requester{
		receiver:    receiver,
		callbackMap: make(map[string]func(interface{})),
		owner:       handler,
	}
	handler.AddRequester(req)
	return req
}

// 请求
func (r *Requester) Request(reqName string, arg interface{}) error {
	return r.receiver.Receive(r.owner, reqName, arg)
}

// 注册回调
func (r *Requester) RegisterCallback(respName string, callback func(interface{})) {
	r.callbackMap[respName] = callback
}

// 请求带回调
func (r *Requester) RequestWithCallback(reqName string, arg interface{}, respName string, callback func(interface{})) error {
	err := r.Request(reqName, arg)
	if err == nil {
		r.callbackMap[respName] = callback
	}
	return err
}

// 处理回调
func (r *Requester) handle(peer IReceiver, respName string, arg interface{}) bool {
	callback, o := r.callbackMap[respName]
	if !o {
		return false
	}
	callback(arg)
	return true
}
