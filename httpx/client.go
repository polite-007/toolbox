package httpx

import (
	"github.com/projectdiscovery/httpx/common/customheader"
	customport "github.com/projectdiscovery/httpx/common/customports"
	"github.com/projectdiscovery/httpx/runner"
)

// HttpxClient 封装 projectdiscovery/httpx 的探测客户端
type HttpxClient struct {
	config *Config
}

// NewHttpxClient 创建一个新的 HttpxClient，支持通过 Option 函数自定义配置
func NewHttpxClient(opts ...Option) *HttpxClient {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return &HttpxClient{config: cfg}
}

// toRunnerOptions 将用户 Config 转换为上游 runner.Options
// 集中管理与上游库的映射，上游 API 变动只需修改此函数
func (c *HttpxClient) toRunnerOptions(targets []string, onResult func(r runner.Result)) *runner.Options {
	cfg := c.config

	opts := &runner.Options{
		// 请求配置
		Methods:         cfg.Method,
		RequestURI:      cfg.Path,
		RequestBody:     cfg.Body,
		CustomHeaders:   customheader.CustomHeaders(cfg.Headers),
		Timeout:         cfg.Timeout,
		FollowRedirects: cfg.FollowRedirects,

		// 探测配置
		Threads:    cfg.Threads,
		RateLimit:  cfg.RateLimit,
		Retries:    cfg.Retries,
		TechDetect: cfg.TechDetect,
		Favicon:    cfg.Favicon,

		// 网络配置
		Proxy: cfg.Proxy,

		// 响应体配置
		ResponseInStdout:          true,
		MaxResponseBodySizeToRead: cfg.ResponseBodySize,
		MaxResponseBodySizeToSave: cfg.ResponseBodySize,

		// 默认行为
		MaxRedirects:  10,
		HostMaxErrors: 30,
		RandomAgent:   true,

		// 运行时注入
		InputTargetHost: targets,
		OnResult:        onResult,
		ShowStatistics:  cfg.ShowStatistics,
	}

	if len(cfg.Ports) > 0 {
		opts.CustomPorts = customport.CustomPorts(cfg.Ports)
	}

	return opts
}