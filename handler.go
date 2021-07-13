package gproc

import (
	"errors"
)

const (
	REQUEST_LIST_LENGTH = 100
)

// 处理器
type Handler struct {
	ch           chan *req
	closed       bool
	requesterMap map[IRequester]struct{}
}

// 回应处理器
type ResponseHandler struct {
	Handler
}

// 初始化
func (h *Handler) Init(reqLen int32) {
	if reqLen <= 0 {
		reqLen = REQUEST_LIST_LENGTH
	}
	h.ch = make(chan *req, reqLen)
	h.requesterMap = make(map[IRequester]struct{})
}

// 关闭
func (h *Handler) Close() {
	if h.closed {
		return
	}
	close(h.ch)
	h.closed = true
}

// 添加请求者
func (h *Handler) AddRequester(r IRequester) {
	h.requesterMap[r] = struct{}{}
}

// 接收请求
func (h *Handler) Receive(peer IReceiver, reqName string, args interface{}) error {
	// 已关闭，防止重复close造成panic
	if h.closed {
		return ErrClosed
	}
	// 请求写入
	h.ch <- &req{
		name: reqName,
		args: args,
		peer: peer,
	}
	return nil
}

// 更新处理IRequester的回调
func (h *Handler) Update() error {
	select {
	case r, o := <-h.ch:
		if !o {
			return errors.New("gproc: handler already closed, break loop")
		}
		// 处理请求
		for k, _ := range h.requesterMap {
			if k.handle(r.peer, r.name, r.args) {
				break
			}
		}
	default:
	}
	return nil
}
