package fofa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	// FOFA API 基础地址
	BaseURL = "https://fofa.info" // 字符串末尾不带斜杠
	// 默认查询超时时间
	DefaultSearchTimeout = 65 * time.Second
	// 默认聚合查询时间
	DefaultStatsTimeout = 300 * time.Second
	// 默认轻量请求超时时间（account 等）
	DefaultAccountTimeout = 30 * time.Second
	// 默认重试次数（仅针对 FOFA 请求过快 errmsg 重试）
	DefaultRetryCount = 3
	// 默认重试基础间隔（秒）
	DefaultRetryInterval = 1
)

// Client FOFA API 客户端
type Client struct {
	email         string        // FOFA 账号邮箱
	key           string        // FOFA API Key
	baseURL       string        // API 基础地址
	httpClient    *http.Client  // HTTP 客户端（不设全局超时，由 context 控制）
	retryCount    int           // 请求过快时的重试次数，默认 3，设为 0 则不重试
	retryInterval time.Duration // 请求过快时首次重试间隔，后续每次翻倍，默认 1s
}

// NewClient 创建新的 FOFA 客户端，支持通过 Option 函数自定义配置
func NewClient(email, key string, opts ...Option) *Client {
	cfg := defaultConfig()
	cfg.email = email
	cfg.key = key
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = newHTTPClient(cfg.proxy)
	}
	return &Client{
		email:         cfg.email,
		key:           cfg.key,
		baseURL:       cfg.baseURL,
		httpClient:    cfg.httpClient,
		retryCount:    cfg.retryCount,
		retryInterval: cfg.retryInterval,
	}
}

// retryDelays 根据配置生成重试间隔序列，每次翻倍
func (c *Client) retryDelays() []time.Duration {
	if c.retryCount <= 0 || c.retryInterval <= 0 {
		return nil
	}
	delays := make([]time.Duration, c.retryCount)
	for i := range delays {
		delays[i] = c.retryInterval * (1 << i) // 每次翻倍: 1s, 2s, 4s, 8s...
	}
	return delays
}

// apiErrorResponse FOFA API 通用错误响应字段
type apiErrorResponse struct {
	Error  bool   `json:"error"`
	ErrMsg string `json:"errmsg"`
}


// do 发送单次 HTTP 请求，并在 errmsg 提示请求过快时按配置重试
func (c *Client) do(ctx context.Context, fullURL string) ([]byte, error) {
	delays := c.retryDelays()
	maxAttempts := 1
	if len(delays) > 0 {
		maxAttempts = len(delays) + 1 // 含首次请求
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, errors.Wrap(err, "构建请求失败")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "发送请求失败")
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "读取响应失败")
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		}

		var apiResp apiErrorResponse
		if json.Unmarshal(body, &apiResp) == nil && apiResp.Error && isRateLimitErrmsg(apiResp.ErrMsg) {
			lastErr = fmt.Errorf("FOFA API 请求过快: %s", apiResp.ErrMsg)
			if attempt < len(delays) && ctx.Err() == nil {
				time.Sleep(delays[attempt])
				continue
			}
			return nil, lastErr
		}

		return body, nil
	}

	return nil, lastErr
}
