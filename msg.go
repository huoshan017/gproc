package gproc

import (
	"sync"
)

type msgType uint8

const (
	msgNormal  msgType = 0 // 普通
	msgSignup  msgType = 1 // 报名
	msgForward msgType = 2 // 转发
	//msgNotify msgType = 3 // 通知
)

// 消息
type msg struct {
	typ     msgType
	fromKey interface{}
	toKey   interface{}
	name    string
	args    interface{}
	sender  ISender
}

// 重置
func (m *msg) reset() {
	m.typ = 0
	m.fromKey = nil
	m.toKey = nil
	m.name = ""
	m.args = nil
	m.sender = nil
}

// 消息池结构
type msgPool struct {
	pool *sync.Pool
}

// 创建消息池
func newMsgPool() *msgPool {
	mp := &msgPool{}
	mp.init()
	return mp
}

// 初始化
func (p *msgPool) init() {
	p.pool = &sync.Pool{
		New: func() interface{} {
			return &msg{}
		},
	}
}

// 获取消息对象
func (p *msgPool) get() *msg {
	return p.pool.Get().(*msg)
}

// 返还消息对象
func (p *msgPool) put(m *msg) {
	p.pool.Put(m)
}

// 消息池对象
var msgpool *msgPool = newMsgPool()

// 是否使用消息池
var usePool bool = true

// 取出消息对象
func getMsg() *msg {
	var m *msg
	if usePool {
		m = msgpool.get()
	} else {
		m = &msg{}
	}
	return m
}

// 放回消息对象
func putMsg(m *msg) {
	if usePool {
		m.reset()
		msgpool.put(m)
	} else {
		m = nil
	}
}
