package httpx

// Config 封装 httpx 探测的所有可配置项，与上游 runner.Options 解耦
type Config struct {
	// 请求配置
	Method          string
	Path            string
	Body            string
	Headers         []string
	CustomHost      string
	Timeout         int
	FollowRedirects bool

	// 探测配置
	Threads   int
	RateLimit int
	Retries   int
	Ports     []string
	TechDetect bool
	Favicon    bool

	// 网络配置
	Proxy string

	// 响应体配置
	ResponseBodySize int

	// 输出配置
	ShowStatistics bool
}

// Option 配置函数类型
type Option func(*Config)

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		Method:           "GET",
		Timeout:          15,
		Retries:          1,
		Threads:          50,
		RateLimit:        150,
		FollowRedirects:  false,
		TechDetect:       false,
		Favicon:          false,
		ShowStatistics:   false,
		ResponseBodySize: 50 * 1024 * 1024, // 50MB
	}
}

// WithMethod 设置请求方法（GET/POST/HEAD 等）
func WithMethod(method string) Option {
	return func(c *Config) {
		c.Method = method
	}
}

// WithPath 设置请求路径
func WithPath(path string) Option {
	return func(c *Config) {
		c.Path = path
	}
}

// WithBody 设置请求 Body
func WithBody(body string) Option {
	return func(c *Config) {
		c.Body = body
	}
}

// WithHeaders 设置自定义请求头（追加模式，可与 WithCustomHost 共存）
func WithHeaders(headers []string) Option {
	return func(c *Config) {
		c.Headers = append(c.Headers, headers...)
	}
}

// WithCustomHost 设置自定义 Host 头（追加模式，可与 WithHeaders 共存）
func WithCustomHost(host string) Option {
	return func(c *Config) {
		c.Headers = append(c.Headers, "Host: "+host)
	}
}

// WithTimeout 设置超时时间（秒）
func WithTimeout(timeout int) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithFollowRedirects 启用跟随重定向
func WithFollowRedirects() Option {
	return func(c *Config) {
		c.FollowRedirects = true
	}
}

// WithThreads 设置并发线程数
func WithThreads(threads int) Option {
	return func(c *Config) {
		c.Threads = threads
	}
}

// WithRateLimit 设置每秒最大请求数
func WithRateLimit(rateLimit int) Option {
	return func(c *Config) {
		c.RateLimit = rateLimit
	}
}

// WithRetries 设置重试次数
func WithRetries(retries int) Option {
	return func(c *Config) {
		c.Retries = retries
	}
}

// WithPorts 设置探测端口列表
func WithPorts(ports []string) Option {
	return func(c *Config) {
		c.Ports = ports
	}
}

// WithTechDetect 启用 Web 指纹识别
func WithTechDetect() Option {
	return func(c *Config) {
		c.TechDetect = true
	}
}

// WithFaviconMMH3 启用 favicon hash 计算
func WithFaviconMMH3() Option {
	return func(c *Config) {
		c.Favicon = true
	}
}

// WithProxy 设置 HTTP/SOCKS5 代理
func WithProxy(proxy string) Option {
	return func(c *Config) {
		c.Proxy = proxy
	}
}

// WithResponseBodySize 设置响应体最大读取和保存大小（字节）
func WithResponseBodySize(size int) Option {
	return func(c *Config) {
		c.ResponseBodySize = size
	}
}

// WithShowStatistics 启用扫描统计信息输出
func WithShowStatistics() Option {
	return func(c *Config) {
		c.ShowStatistics = true
	}
}