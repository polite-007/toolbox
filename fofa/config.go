package fofa

import (
	"net/http"
	"time"
)

// config 封装 FOFA 客户端的所有可配置项
type config struct {
	email         string
	key           string
	baseURL       string
	httpClient    *http.Client
	retryCount    int
	retryInterval time.Duration
}

// Option 配置函数类型
type Option func(*config)

func defaultConfig() *config {
	return &config{
		baseURL:       BaseURL,
		httpClient:    &http.Client{},
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

// WithHTTPClient 设置自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}
