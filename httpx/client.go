package httpx


import "github.com/projectdiscovery/httpx/runner"

type HttpxClient struct {
	Options *runner.Options
}

func NewHttpxClient(opts ...Option) *HttpxClient {
	options := &runner.Options{
		Methods: "GET", // 请求方法
		Timeout: 15,    // 超时时间
		Retries: 1,     // 重试次数
		// Favicon:       true,  // 获取favicon
		MaxRedirects:  10,   // 最大重定向次数
		HostMaxErrors: 30,   // 最大错误次数
		RateLimit:     150,  // 请求速率限制
		RandomAgent:   true, // 随机Agent
		// 设置body获取最大值
		// JSONOutput:                true,
		// 须落库响应正文，否则 blacklink 关键词/HTML 检测拿不到 body（Title 仍会单独解析，但 script 等结构依赖 body）
		ResponseInStdout:          true,
		MaxResponseBodySizeToRead: 50 * 1024 * 1024,
		MaxResponseBodySizeToSave: 50 * 1024 * 1024,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &HttpxClient{
		Options: options,
	}
}