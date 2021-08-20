package gproc

// 请求者，发起请求到IRequesterHandler，除创建初始化外整个生命周期在同一个goroutine中
// 一般跟IRequestHandler不在同一个goroutine
type Requester struct {
	owner       IResponseHandler             // Requester的持有者
	receiver    IRequestHandler              // Requester请求的接收者
	callbackMap map[string]func(interface{}) // 之所以不用线程安全的sync.Map，是因为Requester只在一个goroutine中使用
	options     RequestOptions
	key         interface{} // requester的key，告诉对面的receiver唯一标识自己，用于转发和通知
}

// 创建请求者
func NewRequester(owner IResponseHandler, receiver IRequestHandler, key interface{}, options ...RequestOption) IRequester {
	if owner == nil || receiver == nil {
		panic("owner or receiver is nil")
	}
	req := &Requester{
		owner:       owner,
		receiver:    receiver,
		key:         key,
		callbackMap: make(map[string]func(interface{})),
	}
	owner.addRequester(req)
	for _, option := range options {
		option(req.options)
	}
	return req
}

// 请求
func (r *Requester) Request(msgName string, arg interface{}) error {
	// 相当于RequestHandler接收消息
	err := r.receiver.Recv(r.owner, msgName, arg)
	return err
}

// 注册回调
func (r *Requester) RegisterCallback(msgName string, callback func(interface{})) {
	r.callbackMap[msgName] = callback
}

// 请求带回调，这个方法肯定在handle函数同一goroutine中使用，不存在callbackMap线程安全问题
func (r *Requester) RequestWithCallback(msgName string, arg interface{}, callback func(interface{})) error {
	r.RegisterCallback(msgName, callback)
	return r.Request(msgName, arg)
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
