package gproc

// 请求者，一般跟IReceiver不在同一个goroutine
type Requester struct {
	receiver    IReceiver
	callbackMap map[string]func(interface{})
	owner       IReceiver
}

// 创建请求者，第一个参数为ResponseHandler
func NewRequester4ResponseHandler(handler *ResponseHandler, receiver IReceiver) *Requester {
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

// 创建请求者，第一个参数为LocalService
func NewRequester4LocalService(service *LocalService, receiver IReceiver) *Requester {
	if service == nil || receiver == nil {
		panic("service or receiver is nil")
	}
	req := &Requester{
		receiver:    receiver,
		callbackMap: make(map[string]func(interface{})),
		owner:       service,
	}
	service.AddRequester(req)
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

// 处理回调
func (r *Requester) handle(peer IReceiver, respName string, arg interface{}) bool {
	callback, o := r.callbackMap[respName]
	if !o {
		return false
	}
	callback(arg)
	return true
}
