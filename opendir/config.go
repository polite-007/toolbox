package opendir

import "time"

// Config 封装目录采集的所有可配置项
type Config struct {
	// 采集配置
	Depth       int   // 采集深度，1 表示只采集根目录
	Threads     int   // 共享 worker 数，RunTargets 中所有目标共用
	MaxURLs     int   // 单个目标最多发出的请求数
	MaxFileSize int64 // 单个文件大小上限（字节），<=0 表示不限制

	// 超时配置
	Timeout      int           // TTFB 超时（秒）
	DialTimeout  int           // TCP 建连超时（秒）
	StallTimeout int           // body 停滞超时（秒），<=0 表示不检测
	MaxDuration  time.Duration // 整次运行的总预算（多目标时为整批），0 表示不限制

	// 跳过规则
	SkipNames    []string // 文件名包含任一项则跳过（大小写敏感）
	SkipSuffixes []string // 文件名以任一项结尾则跳过（大小写敏感，自动补 "."）

	// 回调
	OnFile       FileHandler
	OnTargetDone func(TargetResult) // 单个目标完成时回调；在 worker 中并发执行

	// 网络配置
	Proxy           string
	Headers         []string
	UserAgent       string
	InsecureTLS     bool
	FollowRedirects bool

	// 输出配置
	MaxErrors int // Stats.Errors 最多保留的条数
}

// Option 配置函数类型
type Option func(*Config)

const defaultUserAgent = "Mozilla/5.0 (compatible; toolbox-opendir)"

func defaultConfig() *Config {
	return &Config{
		Depth:           1,
		Threads:         50,
		MaxURLs:         1000,
		MaxFileSize:     50 * 1024 * 1024,
		Timeout:         15,
		DialTimeout:     10,
		StallTimeout:    30,
		MaxDuration:     0,
		InsecureTLS:     true,
		FollowRedirects: true,
		UserAgent:       defaultUserAgent,
		MaxErrors:       100,
	}
}

// WithDepth 设置采集深度，1 表示只采集根目录
func WithDepth(depth int) Option {
	return func(c *Config) {
		c.Depth = depth
	}
}

// WithThreads 设置并发 worker 数
func WithThreads(threads int) Option {
	return func(c *Config) {
		c.Threads = threads
	}
}

// WithMaxURLs 设置单个目标最多发出的请求数
func WithMaxURLs(maxURLs int) Option {
	return func(c *Config) {
		c.MaxURLs = maxURLs
	}
}

// WithMaxFileSize 设置单个文件大小上限（字节），<=0 表示不限制
func WithMaxFileSize(size int64) Option {
	return func(c *Config) {
		c.MaxFileSize = size
	}
}

// WithTimeout 设置 TTFB 超时（秒）
func WithTimeout(timeout int) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithDialTimeout 设置 TCP 建连超时（秒）
func WithDialTimeout(timeout int) Option {
	return func(c *Config) {
		c.DialTimeout = timeout
	}
}

// WithStallTimeout 设置 body 停滞超时（秒），<=0 表示不检测
func WithStallTimeout(timeout int) Option {
	return func(c *Config) {
		c.StallTimeout = timeout
	}
}

// WithMaxDuration 设置单目标总预算，0 表示不限制
func WithMaxDuration(d time.Duration) Option {
	return func(c *Config) {
		c.MaxDuration = d
	}
}

// WithSkipNames 设置文件名跳过关键词（大小写敏感的子串匹配）
func WithSkipNames(names []string) Option {
	return func(c *Config) {
		c.SkipNames = append(c.SkipNames, names...)
	}
}

// WithSkipSuffixes 设置文件名跳过后缀（大小写敏感，自动补前导 "."）
func WithSkipSuffixes(suffixes []string) Option {
	return func(c *Config) {
		c.SkipSuffixes = append(c.SkipSuffixes, suffixes...)
	}
}

// WithOnFile 设置文件内容回调
func WithOnFile(handler FileHandler) Option {
	return func(c *Config) {
		c.OnFile = handler
	}
}

// WithOnTargetDone 设置单个目标完成时的回调。
// 该回调在 worker 中并发执行，必须保证 goroutine-safe。
func WithOnTargetDone(fn func(TargetResult)) Option {
	return func(c *Config) {
		c.OnTargetDone = fn
	}
}

// WithProxy 设置 HTTP/SOCKS5 代理
func WithProxy(proxy string) Option {
	return func(c *Config) {
		c.Proxy = proxy
	}
}

// WithHeaders 设置自定义请求头（追加模式）
func WithHeaders(headers []string) Option {
	return func(c *Config) {
		c.Headers = append(c.Headers, headers...)
	}
}

// WithUserAgent 设置 User-Agent
func WithUserAgent(userAgent string) Option {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

// WithInsecureTLS 设置是否跳过 TLS 校验
func WithInsecureTLS(insecure bool) Option {
	return func(c *Config) {
		c.InsecureTLS = insecure
	}
}

// WithFollowRedirects 设置是否跟随重定向
func WithFollowRedirects(follow bool) Option {
	return func(c *Config) {
		c.FollowRedirects = follow
	}
}

// WithMaxErrors 设置 Stats.Errors 最多保留的条数
func WithMaxErrors(maxErrors int) Option {
	return func(c *Config) {
		c.MaxErrors = maxErrors
	}
}
