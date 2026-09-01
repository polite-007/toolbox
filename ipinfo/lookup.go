package ipinfo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Lookup 查询单个 IP 的网络信息。
//   - ctx: 请求上下文
//   - ip: 目标 IP 地址
//
// 非法 IP 直接返回错误；多个源按 failover 顺序依次尝试，全部失败时返回聚合错误。
func (c *IpinfoClient) Lookup(ctx context.Context, ip string) (*Info, error) {
	if !isIP(ip) {
		return nil, fmt.Errorf("invalid ip %q", ip)
	}

	sources := c.resolveSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("unknown source %q", c.config.Source)
	}

	var errs []error
	for _, src := range sources {
		info, err := c.lookup(ctx, src, ip)
		if err == nil {
			return &info, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", src.name, err))
	}

	return nil, fmt.Errorf("all sources failed for %s: %s", ip, errors.Join(errs...))
}

// resolveSources 根据配置返回要尝试的源列表：指定 Source 时仅返回该源，否则返回默认 failover 链。
func (c *IpinfoClient) resolveSources() []source {
	if c.config.Source != "" {
		if src, ok := sourceByName[c.config.Source]; ok {
			return []source{src}
		}
		return nil
	}
	return sourceList
}

// lookup 请求单个源并解析响应。
func (c *IpinfoClient) lookup(ctx context.Context, src source, ip string) (Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.url(ip), nil)
	if err != nil {
		return Info{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Info{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Info{}, err
	}

	return src.parse(body)
}
