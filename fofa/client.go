package fofa

import (
	"net/http"
	"time"
)

const (
	// FOFA API 基础地址
	BaseURL = "https://fofa.info/api/v1/"
	// 默认超时时间
	DefaultTimeout = 30 * time.Second
)

// Client FOFA API 客户端
type Client struct {
	email      string        // FOFA 账号邮箱
	key        string        // FOFA API Key
	baseURL    string        // API 基础地址
	timeout    time.Duration // 请求超时时间
	httpClient *http.Client  // HTTP 客户端
}

// NewClient 创建新的 FOFA 客户端
func NewClient(email, key string) *Client {
	return &Client{
		email:   email,
		key:     key,
		baseURL: BaseURL,
		timeout: DefaultTimeout,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

