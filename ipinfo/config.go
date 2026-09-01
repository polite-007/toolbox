package ipinfo

import "time"

// Config 封装 ipinfo 查询的常用配置。
type Config struct {
	// Timeout 单次 HTTP 请求超时时间
	Timeout time.Duration
	// Proxy 代理地址，支持 http://、https://、socks5:// 或无 scheme（默认按 socks5 处理）
	Proxy string
	// Source 指定数据源；为空时按默认 failover 链依次尝试，非空时仅使用该源
	Source string
}

// Option 配置函数类型。
type Option func(*Config)

// defaultConfig 返回默认配置。
func defaultConfig() *Config {
	return &Config{
		Timeout: 10 * time.Second,
	}
}

// WithTimeout 设置单次请求超时时间。
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithProxy 设置代理地址。支持 "http://host:port"、"https://host:port"、"socks5://host:port"，
// 无 scheme 时默认按 socks5 处理。
func WithProxy(proxy string) Option {
	return func(c *Config) {
		c.Proxy = proxy
	}
}

// WithSource 指定数据源（如 "ipinfo.io"、"ipwho.is"、"ip9.com.cn"、"ipdata.info"），
// 禁用 failover，仅使用该源。
func WithSource(source string) Option {
	return func(c *Config) {
		c.Source = source
	}
}
