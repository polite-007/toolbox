package opendir

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

const defaultMaxRedirects = 10

var (
	// errBodyTooLarge 表示响应体超过配置的上限
	errBodyTooLarge = errors.New("opendir: response body exceeds max body size")
	// errStalled 表示响应体传输停滞超时
	errStalled = errors.New("opendir: response body transfer stalled")
)

// fetchConfig 是内部 HTTP 抓取器的配置
type fetchConfig struct {
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

// fetchResponse 是单次响应
type fetchResponse struct {
	StatusCode    int
	Header        http.Header
	URL           *url.URL      // 重定向后的最终 URL
	ContentLength int64         // 响应体长度，未知为 -1
	Body          io.ReadCloser // 带 stall 看门狗与大小上限的响应体
}

// fetcher 提供单次、可取消、流式的 HTTP 请求能力。
//
// 超时分为三层：
//   - DialTimeout：TCP 建连超时
//   - Timeout：TTFB（TLS 握手 + 发送请求 + 等待响应头）超时
//   - StallTimeout：body 传输期间连续无字节的超时（慢速但持续的下载不受影响）
//
// 整体生命周期由调用方通过 context 控制。
type fetcher struct {
	client       *http.Client
	proxyErr     error
	userAgent    string
	headers      []string
	stallTimeout time.Duration
	maxBodySize  int64
}

func newFetcher(cfg fetchConfig) *fetcher {
	f := &fetcher{}

	proxyFunc := http.ProxyFromEnvironment
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		switch {
		case err != nil:
			f.proxyErr = fmt.Errorf("opendir: invalid proxy %q: %w", cfg.Proxy, err)
		case proxyURL.Host == "":
			f.proxyErr = fmt.Errorf("opendir: invalid proxy %q", cfg.Proxy)
		default:
			proxyFunc = http.ProxyURL(proxyURL)
		}
	}

	checkRedirect := func(req *http.Request, via []*http.Request) error {
		if !cfg.FollowRedirects {
			return http.ErrUseLastResponse
		}
		if cfg.MaxRedirects > 0 && len(via) >= cfg.MaxRedirects {
			return fmt.Errorf("opendir: stopped after %d redirects", cfg.MaxRedirects)
		}
		if len(via) > 0 && !strings.EqualFold(via[0].URL.Hostname(), req.URL.Hostname()) {
			return fmt.Errorf("opendir: cross-host redirect to %q is not allowed", req.URL.Hostname())
		}
		return nil
	}

	transport := &http.Transport{
		Proxy:                 proxyFunc,
		DialContext:           (&net.Dialer{Timeout: cfg.DialTimeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: cfg.InsecureTLS}, // #nosec G402 -- 默认跳过校验，便于采集自签证书目标
		TLSHandshakeTimeout:   cfg.Timeout,
		ResponseHeaderTimeout: cfg.Timeout,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
	}

	f.client = &http.Client{
		Transport:     transport,
		CheckRedirect: checkRedirect,
	}
	f.userAgent = cfg.UserAgent
	f.headers = cfg.Headers
	f.stallTimeout = cfg.StallTimeout
	f.maxBodySize = cfg.MaxBodySize
	return f
}

// Get 发起一次 GET 请求并返回流式响应。
// 返回的 Body 必须关闭；关闭时同时释放请求 context。
func (f *fetcher) Get(ctx context.Context, target string) (*fetchResponse, error) {
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
	if f.userAgent != "" {
		req.Header.Set("User-Agent", f.userAgent)
	}
	for _, header := range f.headers {
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

	body := newFetcherBody(resp.Body, cancel, f.stallTimeout, f.maxBodySize)
	return &fetchResponse{
		StatusCode:    resp.StatusCode,
		Header:        resp.Header,
		URL:           resp.Request.URL,
		ContentLength: resp.ContentLength,
		Body:          body,
	}, nil
}

// Close 关闭空闲连接，不影响进行中的请求
func (f *fetcher) Close() {
	f.client.CloseIdleConnections()
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
			return 0, errBodyTooLarge
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
		return 0, errBodyTooLarge
	}
	if err != nil {
		if b.timer != nil {
			b.timer.Stop()
		}
		if b.stalled.Load() {
			return n, errStalled
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
