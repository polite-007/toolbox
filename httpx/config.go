package httpx

import (
	"github.com/projectdiscovery/httpx/common/customheader"
	customport "github.com/projectdiscovery/httpx/common/customports"
	"github.com/projectdiscovery/httpx/runner"
)


// Option 配置函数类型，用于自定义 runner.Options
type Option func(*runner.Options)

// WithPath 设置请求路径
func WithPath(path string) Option {
	return func(opts *runner.Options) {
		opts.RequestURI = path
	}
}

// WithPorxy 设置代理
func WithPorxy(proxy string) Option {
	return func(opts *runner.Options) {
		opts.Proxy = proxy
	}
}

// WithCustomHost 设置自定义Host
func WithCustomHost(host string) Option {
	return func(opts *runner.Options) {
		opts.CustomHeaders = customheader.CustomHeaders{"Host: " + host}
	}
}

// WithHeaders 设置自定义Headers
func WithHeaders(headers []string) Option {
	return func(opts *runner.Options) {
		opts.CustomHeaders = customheader.CustomHeaders(headers)
	}
}

// WithBody 设置请求Body
func WithBody(body string) Option {
	return func(opts *runner.Options) {
		opts.RequestBody = body
	}
}

// WithPorts 设置请求Ports
func WithPorts(ports []string) Option {
	return func(opts *runner.Options) {
		opts.CustomPorts = customport.CustomPorts(ports)
	}
}

// WithFollowRedirects 设置是否跟随重定向
func WithFollowRedirects() Option {
	return func(opts *runner.Options) {
		opts.FollowRedirects = true
	}
}

// WithThreads 设置线程数
func WithThreads(threads int) Option {
	return func(opts *runner.Options) {
		opts.Threads = threads
	}
}

// WithFaviconMMH3 设置是否获取faviconMMH3
func WithFaviconMMH3() Option {
	return func(opts *runner.Options) {
		opts.Favicon = true
	}
}

// WithShowStatistics 设置是否显示统计信息
func WithShowStatistics() Option {
	return func(opts *runner.Options) {
		opts.ShowStatistics = true
	}
}

// WithTechDetect 设置是否检测技术
func WithTechDetect() Option {
	return func(opts *runner.Options) {
		opts.TechDetect = true
	}
}

// 设置请求速率限制
func WithRateLimit(rateLimit int) Option {
	return func(opts *runner.Options) {
		opts.RateLimit = rateLimit
	}
}