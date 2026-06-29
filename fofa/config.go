package fofa

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// config 封装 FOFA 客户端的所有可配置项
type config struct {
	email         string
	key           string
	baseURL       string
	proxy         string
	httpClient    *http.Client
	retryCount    int
	retryInterval time.Duration
}

// Option 配置函数类型
type Option func(*config)

func defaultConfig() *config {
	return &config{
		baseURL:       BaseURL,
		retryCount:    DefaultRetryCount,
		retryInterval: DefaultRetryInterval * time.Second,
	}
}

// WithBaseURL 设置 API 基础地址（用于自定义或测试）
func WithBaseURL(baseURL string) Option {
	return func(c *config) {
		c.baseURL = baseURL
	}
}

// WithProxy 设置 HTTP/HTTPS/SOCKS5 代理，如 http://127.0.0.1:7890 或 socks5://127.0.0.1:7890；默认无代理
func WithProxy(proxyURL string) Option {
	return func(c *config) {
		c.proxy = strings.TrimSpace(proxyURL)
	}
}

// WithRetryCount 设置重试次数，设为 0 则不重试
func WithRetryCount(count int) Option {
	return func(c *config) {
		c.retryCount = count
	}
}

// WithRetryInterval 设置首次重试间隔，后续每次翻倍
func WithRetryInterval(d time.Duration) Option {
	return func(c *config) {
		c.retryInterval = d
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端（优先级高于 WithProxy）
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

func newHTTPClient(proxyURL string) *http.Client {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err == nil {
			switch u.Scheme {
			case "http", "https":
				transport.Proxy = http.ProxyURL(u)
			case "socks5", "socks5h":
				dialer, dialErr := proxy.FromURL(u, proxy.Direct)
				if dialErr == nil {
					transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						if cd, ok := dialer.(proxy.ContextDialer); ok {
							return cd.DialContext(ctx, network, addr)
						}
						return dialer.Dial(network, addr)
					}
				}
			}
		}
	}

	return &http.Client{Transport: transport}
}
