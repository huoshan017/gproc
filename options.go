package gproc

// 请求选项结构
type RequestOptions struct {
	requestTimeout int32
}

// 请求超时
func (options *RequestOptions) SetRequestTimeout(timeout int32) {
	options.requestTimeout = timeout
}

// 请求选项
type RequestOption func(RequestOptions)

func RequestTimeout(timeout int32) RequestOption {
	return func(options RequestOptions) {
		options.SetRequestTimeout(timeout)
	}
}
