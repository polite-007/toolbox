package ipinfo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// IpinfoClient 封装 IP 信息查询客户端。
type IpinfoClient struct {
	config *Config
	http   *http.Client
}

// NewIpinfoClient 创建一个新的 IpinfoClient，支持通过 Option 函数自定义配置。
func NewIpinfoClient(opts ...Option) (*IpinfoClient, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	transport, err := buildTransport(cfg.Proxy)
	if err != nil {
		return nil, err
	}

	return &IpinfoClient{
		config: cfg,
		http: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
	}, nil
}

// buildTransport 根据代理配置构造 http.Transport。
// 支持 http/https 代理（标准库原生）与 socks5 代理（golang.org/x/net/proxy）。
func buildTransport(proxyStr string) (*http.Transport, error) {
	if proxyStr == "" {
		return &http.Transport{}, nil
	}

	// 无 scheme 时默认按 socks5 处理，与 naabu 封装保持一致
	if !strings.Contains(proxyStr, "://") {
		proxyStr = "socks5://" + proxyStr
	}

	u, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy %q: %w", proxyStr, err)
	}

	switch u.Scheme {
	case "http", "https":
		return &http.Transport{Proxy: http.ProxyURL(u)}, nil

	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("invalid socks5 proxy %q: %w", proxyStr, err)
		}
		return &http.Transport{
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q", u.Scheme)
	}
}
