package naabu

import "time"

// Config 封装 naabu 端口扫描的常用配置，与上游 runner.Options 解耦。
type Config struct {
	// 目标配置
	TargetsFile string

	// 端口配置
	Ports    string
	TopPorts string

	// 扫描配置
	ScanType            string
	Rate                int
	Retries             int
	Timeout             time.Duration
	Threads             int
	WarmUpTime          int
	Verify              bool
	HostPortConcurrency int

	// 功能开关
	WithHostDiscovery bool
	Passive           bool
	ServiceDiscovery  bool
	ServiceVersion    bool
	ExcludeCDN        bool

	// 网络配置
	Proxy          string
	ProxyAuth      string
	Resolvers      string
	SystemResolver bool

	// 输出行为
	Silent bool
}

// Option 配置函数类型。
type Option func(*Config)

// defaultConfig 返回默认配置。
func defaultConfig() *Config {
	return &Config{
		TopPorts:            "100",
		ScanType:            "c",
		Rate:                1000,
		Retries:             3,
		Timeout:             time.Second,
		Threads:             25,
		WarmUpTime:          2,
		HostPortConcurrency: 10,
		Silent:              true,
	}
}

// WithTargetsFile 设置目标文件路径。
func WithTargetsFile(file string) Option {
	return func(c *Config) {
		c.TargetsFile = file
	}
}

// WithPorts 设置扫描端口表达式，例如 "80,443,100-200,u:53"。
func WithPorts(ports string) Option {
	return func(c *Config) {
		c.Ports = ports
	}
}

// WithSynScan 使用 SYN 扫描，通常需要管理员/root 权限。
func WithSynScan() Option {
	return func(c *Config) {
		c.ScanType = "s"
	}
}

// WithRate 设置每秒发包数量。
func WithRate(rate int) Option {
	return func(c *Config) {
		c.Rate = rate
	}
}

// WithRetries 设置扫描重试次数。
func WithRetries(retries int) Option {
	return func(c *Config) {
		c.Retries = retries
	}
}

// WithTimeout 设置端口扫描超时时间。
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithThreads 设置内部工作线程数。
func WithThreads(threads int) Option {
	return func(c *Config) {
		c.Threads = threads
	}
}

// WithWarmUpTime 设置扫描阶段之间的等待秒数。
func WithWarmUpTime(seconds int) Option {
	return func(c *Config) {
		c.WarmUpTime = seconds
	}
}

// WithVerify 启用 TCP 二次验证。
func WithVerify() Option {
	return func(c *Config) {
		c.Verify = true
	}
}

// WithHostPortConcurrency 设置 RunHostPorts 按 host 分组扫描时的最大并发数。
func WithHostPortConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.HostPortConcurrency = concurrency
	}
}

// WithHostDiscovery 启用端口扫描前的主机发现。
func WithHostDiscovery() Option {
	return func(c *Config) {
		c.WithHostDiscovery = true
	}
}

// WithPassive 启用被动端口枚举。
func WithPassive() Option {
	return func(c *Config) {
		c.Passive = true
	}
}

// WithServiceDiscovery 启用基于端口号的服务识别。
func WithServiceDiscovery() Option {
	return func(c *Config) {
		c.ServiceDiscovery = true
	}
}

// WithServiceVersion 启用主动服务版本探测。
func WithServiceVersion() Option {
	return func(c *Config) {
		c.ServiceVersion = true
	}
}

// WithExcludeCDN 对 CDN/WAF 目标跳过全端口扫描，仅扫描常见 Web 端口。
func WithExcludeCDN() Option {
	return func(c *Config) {
		c.ExcludeCDN = true
	}
}

// WithProxy 设置 SOCKS5 代理，可选 auth 格式为 "username:password"。
func WithProxy(proxy string, auth ...string) Option {
	return func(c *Config) {
		c.Proxy = proxy
		if len(auth) > 0 {
			c.ProxyAuth = auth[0]
		}
	}
}

// WithResolvers 设置自定义 DNS resolver，支持逗号分隔或文件路径。
func WithResolvers(resolvers string) Option {
	return func(c *Config) {
		c.Resolvers = resolvers
	}
}

// WithSystemResolver 启用系统 DNS 作为 fallback resolver。
func WithSystemResolver() Option {
	return func(c *Config) {
		c.SystemResolver = true
	}
}

// WithSilent 设置是否静默输出。
func WithSilent(silent bool) Option {
	return func(c *Config) {
		c.Silent = silent
	}
}
