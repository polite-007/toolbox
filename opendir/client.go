package opendir

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/polite-007/toolbox/httpx"
)

// OpenDirClient 开放目录采集客户端
type OpenDirClient struct {
	config  *Config
	fetcher *httpx.Fetcher
}

// NewOpenDirClient 创建一个新的 OpenDirClient
func NewOpenDirClient(opts ...Option) *OpenDirClient {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	fetcher := httpx.NewFetcher(
		httpx.WithFetchDialTimeout(time.Duration(cfg.DialTimeout)*time.Second),
		httpx.WithFetchTimeout(time.Duration(cfg.Timeout)*time.Second),
		httpx.WithFetchStallTimeout(time.Duration(cfg.StallTimeout)*time.Second),
		// 大小限制由 opendir 按目录页/文件分别控制
		httpx.WithFetchMaxBodySize(0),
		httpx.WithFetchFollowRedirects(cfg.FollowRedirects),
		httpx.WithFetchInsecureTLS(cfg.InsecureTLS),
		httpx.WithFetchProxy(cfg.Proxy),
		httpx.WithFetchHeaders(cfg.Headers),
		httpx.WithFetchUserAgent(cfg.UserAgent),
	)

	return &OpenDirClient{config: cfg, fetcher: fetcher}
}

// normalizeTarget 补全 scheme 并校验目标
func normalizeTarget(target string) (*url.URL, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, ErrInvalidTarget
	}
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%w: unsupported scheme %q", ErrInvalidTarget, u.Scheme)
	}
	if u.Host == "" {
		return nil, ErrInvalidTarget
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}
	return u, nil
}
