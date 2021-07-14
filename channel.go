package gproc

const (
	CHANNEL_LENGTH = 500
)

// 通道
type Channel struct {
	ch      chan *msg
	closed  bool
	chClose chan struct{}
}

// 创建通道
func NewChannel(length int32) *Channel {
	channel := &Channel{}
	channel.Init(length)
	return channel
}

// 初始化
func (c *Channel) Init(length int32) {
	if length <= 0 {
		length = CHANNEL_LENGTH
	}
	c.ch = make(chan *msg, length)
	c.chClose = make(chan struct{})
}

// 关闭
func (c *Channel) Close() {
	if c.closed {
		return
	}
	close(c.chClose)
	c.closed = true
}

// 发送
func (c *Channel) Send(sender ISender, reqName string, args interface{}) error {
	// 已关闭，防止重复close造成panic
	if c.closed {
		return ErrClosed
	}
	// 请求写入
	m := &msg{
		name:   reqName,
		args:   args,
		sender: sender,
	}
	c.ch <- m
	return nil
}

// 接收
func (c *Channel) Recv() (*msg, error) {
	select {
	case m, o := <-c.ch:
		if !o {
			return nil, ErrClosed
		}
		return m, nil
	case <-c.chClose:
		c.closed = true
	default:
	}
	return nil, nil
}

// 是否已经关闭
func (c *Channel) IsClosed() bool {
	return c.closed
}
