package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Fetcher 提供单次、可取消、流式的 HTTP 请求能力。
//
// 与 Run 的区别：Run 面向批量探测（结果通过回调返回，响应体截断），
// Fetcher 面向需要读取完整响应体、需要精确控制超时的场景，
// 例如目录遍历、文件下载。
//
// 超时分为四层：
//   - DialTimeout：TCP 建连超时
//   - Timeout：TTFB（TLS 握手 + 发送请求 + 等待响应头）超时
//   - StallTimeout：body 传输期间连续无字节的超时（慢速但持续的下载不受影响）
//   - ctx：调用方通过 context 控制整体生命周期
type Fetcher struct {
	config   *FetchConfig
	client   *http.Client
	proxyErr error
}

// FetchConfig 是 Fetcher 的配置
type FetchConfig struct {
	DialTimeout     time.Duration // TCP 建连超时
	Timeout         time.Duration // TTFB 超时（含 TLS 握手、等待响应头）
	StallTimeout    time.Duration // body 连续无字节的超时，<=0 表示不检测
	MaxBodySize     int64         // 响应体上限，<=0 表示不限制
	FollowRedirects bool          // 是否跟随重定向
	MaxRedirects    int           // 最大重定向次数
	InsecureTLS     bool          // 是否跳过 TLS 校验
	Proxy           string        // HTTP/SOCKS5 代理
	Headers         []string      // 自定义请求头，格式 "Key: Value"
	UserAgent       string        // User-Agent
}

// FetchOption 配置函数类型
type FetchOption func(*FetchConfig)

// FetchResponse 是 Fetcher 的单次响应
type FetchResponse struct {
	StatusCode    int
	Header        http.Header
	URL           *url.URL      // 重定向后的最终 URL
	ContentLength int64         // 响应体长度，未知为 -1
	Body          io.ReadCloser // 带 stall 看门狗与大小上限的响应体
}

// ErrBodyTooLarge 表示响应体超过配置的上限
var ErrBodyTooLarge = errors.New("httpx: response body exceeds max body size")

// ErrStalled 表示响应体传输停滞超时
var ErrStalled = errors.New("httpx: response body transfer stalled")

const (
	defaultFetchDialTimeout  = 10 * time.Second
	defaultFetchTTFBTimeout  = 15 * time.Second
	defaultFetchStallTimeout = 30 * time.Second
	defaultFetchMaxBodySize  = int64(50 * 1024 * 1024)
	defaultFetchMaxRedirects = 10
	defaultFetchUserAgent    = "Mozilla/5.0 (compatible; toolbox-opendir)"
)

func defaultFetchConfig() *FetchConfig {
	return &FetchConfig{
		DialTimeout:     defaultFetchDialTimeout,
		Timeout:         defaultFetchTTFBTimeout,
		StallTimeout:    defaultFetchStallTimeout,
		MaxBodySize:     defaultFetchMaxBodySize,
		FollowRedirects: true,
		MaxRedirects:    defaultFetchMaxRedirects,
		InsecureTLS:     true,
		UserAgent:       defaultFetchUserAgent,
	}
}

// WithFetchDialTimeout 设置 TCP 建连超时
func WithFetchDialTimeout(timeout time.Duration) FetchOption {
	return func(c *FetchConfig) {
		c.DialTimeout = timeout
	}
}

// WithFetchTimeout 设置 TTFB 超时（含 TLS 握手、等待响应头）
func WithFetchTimeout(timeout time.Duration) FetchOption {
	return func(c *FetchConfig) {
		c.Timeout = timeout
	}
}

// WithFetchStallTimeout 设置 body 传输停滞超时，<=0 表示不检测
func WithFetchStallTimeout(timeout time.Duration) FetchOption {
	return func(c *FetchConfig) {
		c.StallTimeout = timeout
	}
}

// WithFetchMaxBodySize 设置响应体上限（字节），<=0 表示不限制
func WithFetchMaxBodySize(size int64) FetchOption {
	return func(c *FetchConfig) {
		c.MaxBodySize = size
	}
}

// WithFetchFollowRedirects 设置是否跟随重定向
func WithFetchFollowRedirects(follow bool) FetchOption {
	return func(c *FetchConfig) {
		c.FollowRedirects = follow
	}
}

// WithFetchMaxRedirects 设置最大重定向次数
func WithFetchMaxRedirects(maxRedirects int) FetchOption {
	return func(c *FetchConfig) {
		c.MaxRedirects = maxRedirects
	}
}

// WithFetchInsecureTLS 设置是否跳过 TLS 校验
func WithFetchInsecureTLS(insecure bool) FetchOption {
	return func(c *FetchConfig) {
		c.InsecureTLS = insecure
	}
}

// WithFetchProxy 设置 HTTP/SOCKS5 代理
func WithFetchProxy(proxy string) FetchOption {
	return func(c *FetchConfig) {
		c.Proxy = proxy
	}
}

// WithFetchHeaders 设置自定义请求头（追加模式）
func WithFetchHeaders(headers []string) FetchOption {
	return func(c *FetchConfig) {
		c.Headers = append(c.Headers, headers...)
	}
}

// WithFetchUserAgent 设置 User-Agent
func WithFetchUserAgent(userAgent string) FetchOption {
	return func(c *FetchConfig) {
		c.UserAgent = userAgent
	}
}

// NewFetcher 创建一个新的 Fetcher
func NewFetcher(opts ...FetchOption) *Fetcher {
	cfg := defaultFetchConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	f := &Fetcher{config: cfg}

	proxyFunc := http.ProxyFromEnvironment
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		switch {
		case err != nil:
			f.proxyErr = fmt.Errorf("httpx: invalid proxy %q: %w", cfg.Proxy, err)
		case proxyURL.Host == "":
			f.proxyErr = fmt.Errorf("httpx: invalid proxy %q", cfg.Proxy)
		default:
			proxyFunc = http.ProxyURL(proxyURL)
		}
	}

	transport := &http.Transport{
		Proxy:                 proxyFunc,
		DialContext:           (&net.Dialer{Timeout: cfg.DialTimeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: cfg.InsecureTLS}, // #nosec G402 -- 与 httpx 默认行为一致
		TLSHandshakeTimeout:   cfg.Timeout,
		ResponseHeaderTimeout: cfg.Timeout,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
	}

	f.client = &http.Client{
		Transport:     transport,
		CheckRedirect: f.checkRedirect,
	}
	return f
}

// Close 关闭空闲连接，不影响进行中的请求
func (f *Fetcher) Close() {
	f.client.CloseIdleConnections()
}

// Get 发起一次 GET 请求并返回流式响应。
// 返回的 Body 必须关闭；关闭时同时释放请求 context。
func (f *Fetcher) Get(ctx context.Context, target string) (*FetchResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.proxyErr != nil {
		return nil, f.proxyErr
	}

	target = normalizeFetchTarget(target)
	reqCtx, cancel := context.WithCancel(ctx)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	if f.config.UserAgent != "" {
		req.Header.Set("User-Agent", f.config.UserAgent)
	}
	for _, header := range f.config.Headers {
		key, value, ok := strings.Cut(header, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		req.Header.Set(key, strings.TrimSpace(value))
	}

	resp, err := f.client.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	body := newFetcherBody(resp.Body, cancel, f.config.StallTimeout, f.config.MaxBodySize)
	return &FetchResponse{
		StatusCode:    resp.StatusCode,
		Header:        resp.Header,
		URL:           resp.Request.URL,
		ContentLength: resp.ContentLength,
		Body:          body,
	}, nil
}

// checkRedirect 只允许同 hostname 的重定向，避免被带到外部站点
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if !f.config.FollowRedirects {
		return http.ErrUseLastResponse
	}
	if f.config.MaxRedirects > 0 && len(via) >= f.config.MaxRedirects {
		return fmt.Errorf("httpx: stopped after %d redirects", f.config.MaxRedirects)
	}
	if len(via) > 0 && !strings.EqualFold(via[0].URL.Hostname(), req.URL.Hostname()) {
		return fmt.Errorf("httpx: cross-host redirect to %q is not allowed", req.URL.Hostname())
	}
	return nil
}

func normalizeFetchTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return target
	}
	if !strings.Contains(target, "://") {
		return "http://" + target
	}
	return target
}

// fetcherBody 组合了大小上限、停滞看门狗和请求取消
type fetcherBody struct {
	rc      io.ReadCloser
	cancel  context.CancelFunc
	timer   *time.Timer
	stall   time.Duration
	stalled atomic.Bool
	max     int64
	read    int64
	once    sync.Once
}

func newFetcherBody(rc io.ReadCloser, cancel context.CancelFunc, stall time.Duration, max int64) *fetcherBody {
	b := &fetcherBody{rc: rc, cancel: cancel, stall: stall, max: max}
	if stall > 0 {
		b.timer = time.AfterFunc(stall, func() {
			b.stalled.Store(true)
			cancel()
		})
	}
	return b
}

func (b *fetcherBody) Read(p []byte) (int, error) {
	if b.max > 0 {
		remaining := b.max + 1 - b.read
		if remaining <= 0 {
			return 0, ErrBodyTooLarge
		}
		if int64(len(p)) > remaining {
			p = p[:remaining]
		}
	}

	n, err := b.rc.Read(p)
	b.read += int64(n)

	if b.timer != nil && n > 0 {
		b.timer.Reset(b.stall)
	}
	if b.max > 0 && b.read > b.max {
		return 0, ErrBodyTooLarge
	}
	if err != nil {
		if b.timer != nil {
			b.timer.Stop()
		}
		if b.stalled.Load() {
			return n, ErrStalled
		}
	}
	return n, err
}

func (b *fetcherBody) Close() error {
	b.once.Do(func() {
		if b.timer != nil {
			b.timer.Stop()
		}
		b.cancel()
	})
	return b.rc.Close()
}
