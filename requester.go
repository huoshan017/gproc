package gproc

// 请求者，发起请求到IRequesterHandler，除创建初始化外整个生命周期在同一个goroutine中
// 一般跟IRequestHandler不在同一个goroutine
type Requester struct {
	owner       IResponseHandler                                       // Requester的持有者
	receiver    IRequestHandler                                        // Requester请求的接收者
	callbackMap map[uint32]func(interface{})                           // 之所以不用线程安全的sync.Map，是因为Requester只在一个goroutine中使用
	forwardMap  map[uint32]func(fromKey interface{}, args interface{}) // 转发消息到处理函数的映射
	options     RequestOptions                                         // 请求选项
	key         interface{}                                            // requester的key，告诉对面的receiver唯一标识自己，用于转发和通知
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
		callbackMap: make(map[uint32]func(interface{})),
		forwardMap:  make(map[uint32]func(interface{}, interface{})),
	}
	owner.addRequester(req)
	for _, option := range options {
		option(req.options)
	}
	req.signUp()
	return req
}

// 请求
func (r *Requester) Request(msgId uint32, args interface{}) error {
	var m *msg = getMsg()
	m.typ = msgNormal
	m.sender = r.owner
	m.id = msgId
	m.args = args

	// 相当于RequestHandler接收消息
	return r.receiver.recv(m)
}

// 注册回调
func (r *Requester) RegisterCallback(msgId uint32, callback func(interface{})) {
	r.callbackMap[msgId] = callback
}

// 注册通知回调，与普通回调共享
func (r *Requester) RegisterNotify(msgId uint32, notify func(interface{})) {
	r.RegisterCallback(msgId, notify)
}

// 请求带回调，这个方法肯定在handle函数同一goroutine中使用，不存在callbackMap线程安全问题
func (r *Requester) RequestWithCallback(msgId uint32, arg interface{}, callback func(interface{})) error {
	r.RegisterCallback(msgId, callback)
	return r.Request(msgId, arg)
}

// 转发请求
func (r *Requester) RequestForward(toKey interface{}, msgId uint32, args interface{}) error {
	m := getMsg()
	m.typ = msgForward
	m.fromKey = r.key
	m.toKey = toKey
	m.id = msgId
	m.args = args
	return r.receiver.recv(m)
}

// 注册转发处理器
func (r *Requester) RegisterForward(msgId uint32, handle func(interface{}, interface{})) {
	r.forwardMap[msgId] = handle
}

// 处理回调
func (r *Requester) handle(m *msg) bool {
	if m.typ == msgNormal {
		callback, o := r.callbackMap[m.id]
		if !o {
			return false
		}
		callback(m.args)
	} else if m.typ == msgForward {
		handle, o := r.forwardMap[m.id]
		if !o {
			return false
		}
		handle(m.fromKey, m.args)
	} else {
		return false
	}
	return true
}

// 报名
func (r *Requester) signUp() error {
	m := getMsg()
	m.typ = msgSignup
	m.fromKey = r.key
	m.sender = r.owner
	return r.receiver.recv(m)
}
