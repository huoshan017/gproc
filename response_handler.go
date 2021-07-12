package gproc

import (
	"errors"
	"sync"
)

// 回应处理器
type ResponseHandler struct {
	chReq        chan *req
	chClose      chan struct{}
	closed       bool
	requesterMap sync.Map
}

// 创建
func NewResponseHandler() *ResponseHandler {
	handler := &ResponseHandler{}
	handler.Init()
	return handler
}

// 初始化
func (h *ResponseHandler) Init() {
	h.chReq = make(chan *req, 100)
	h.chClose = make(chan struct{})
}

// 添加请求者
func (h *ResponseHandler) AddRequester(r *Requester) {
	h.requesterMap.Store(r, true)
}

// 关闭
func (h *ResponseHandler) Close() {
	if h.closed {
		return
	}
	close(h.chClose)
	h.closed = true
}

// 接收请求
func (h *ResponseHandler) Receive(peer IReceiver, reqName string, args interface{}) error {
	// 已关闭，防止重复close造成panic
	if h.closed {
		return ErrClosed
	}
	// 请求写入
	h.chReq <- &req{
		name: reqName,
		args: args,
		peer: peer,
	}
	return nil
}

// 更新处理回调
func (h *ResponseHandler) Update() error {
	select {
	case r, o := <-h.chReq:
		if !o {
			return errors.New("gproc: service already closed, break loop")
		}
		// 处理请求
		h.requesterMap.Range(func(key, _ interface{}) bool {
			req, o := key.(*Requester)
			if !o {
				return false
			}
			return req.handle(r.peer, r.name, r.args)
		})
	default:
	}
	return nil
}
