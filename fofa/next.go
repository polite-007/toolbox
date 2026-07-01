package fofa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	// DefaultNextTokenTTL FOFA 连续翻页 next 游标有效期
	DefaultNextTokenTTL = 15 * time.Minute
	// nextTokenSafetyMargin 游标过期安全余量，提前结束避免临界失效
	nextTokenSafetyMargin = 30 * time.Second
)

// NextRequest 连续翻页请求参数
// 文档: https://fofa.info/api/batches_pages
type NextRequest struct {
	Ctx    context.Context // 上下文
	Query  string          // 搜索查询语句
	Fields string          // 返回字段，多个字段用逗号分隔
	Size   int             // 每页条数，默认 100，最大 10000
	Next   string          // 上一页响应中的 next 游标；首页留空
	Full   bool            // 是否搜索全部数据，默认 false（近一年）
}

// NextResponse 连续翻页响应结果
type NextResponse struct {
	Error   bool       `json:"error"`   // 是否有错误
	ErrMsg  string     `json:"errmsg"`  // 错误信息
	Size    int        `json:"size"`    // 匹配总数
	Mode    string     `json:"mode"`    // 查询模式
	Query   string     `json:"query"`   // 查询语句
	Next    string     `json:"next"`    // 下一页游标，空表示没有更多数据
	Results [][]string `json:"results"` // 搜索结果，二维数组
	Fields  []string   // 字段名列表，用于映射 Results 的列
}

// HasMore 是否还有下一页
func (r *NextResponse) HasMore() bool {
	return r.Next != ""
}

// GetResults 将 Results 转化成可直接使用的结构体
func (r *NextResponse) GetResults() []*SearchResult {
	records := make([]*SearchResult, 0, len(r.Results))
	for _, res := range r.Results {
		records = append(records, newRecordFromStrings(res, r.Fields))
	}
	return records
}

// SearchNext 执行连续翻页查询（Search After）
// 不传 Next 时返回第一页；后续请求需传入上一页响应中的 Next
func (c *Client) SearchNext(req *NextRequest) (*NextResponse, error) {
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	if req.Size <= 0 {
		req.Size = 100
	}
	if req.Size > 10000 {
		req.Size = 10000
	}
	if req.Fields == "" {
		req.Fields = "ip,port"
	}
	req.Fields = ensureMinFields(req.Fields)

	qbase64 := base64.StdEncoding.EncodeToString([]byte(req.Query))

	apiURL := fmt.Sprintf("%s/api/v1/search/next", c.baseURL)
	params := url.Values{}
	params.Set("key", c.key)
	params.Set("qbase64", qbase64)
	params.Set("size", fmt.Sprintf("%d", req.Size))
	params.Set("fields", req.Fields)
	params.Set("full", strconv.FormatBool(req.Full))
	if req.Next != "" {
		params.Set("next", req.Next)
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	body, err := c.do(ctx, fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "连续翻页请求失败")
	}

	var nextResp NextResponse
	if err := json.Unmarshal(body, &nextResp); err != nil {
		return nil, errors.Wrap(err, "解析响应 JSON 失败")
	}
	if nextResp.Error {
		return nil, fmt.Errorf("FOFA API 错误: %s", nextResp.ErrMsg)
	}

	nextResp.Fields = splitFields(req.Fields)
	return &nextResp, nil
}

// SearchNextAll 连续翻页拉取同一查询的全部结果
// fn 每收到一页数据调用一次；返回 error 可中止后续翻页
// 单页失败时对同一 next 游标重试；next 游标有效期约 15 分钟，超时后需重新发起查询
func (c *Client) SearchNextAll(req *NextRequest, fn func(*NextResponse) error) error {
	if req == nil {
		return errors.New("请求参数不能为空")
	}
	if fn == nil {
		return errors.New("回调函数不能为空")
	}

	nextToken := req.Next
	nextTokenAt := time.Time{}
	if nextToken != "" {
		nextTokenAt = time.Now()
	}

	for {
		if err := contextErr(req.Ctx); err != nil {
			return err
		}
		if err := checkNextTokenExpired(nextToken, nextTokenAt); err != nil {
			return err
		}

		resp, err := c.searchNextWithRetry(req, nextToken)
		if err != nil {
			return err
		}

		// 收到 next 游标后立即记录时间，避免 fn 处理耗时挤占 15 分钟有效期
		pageNext := resp.Next
		pageNextAt := time.Time{}
		if resp.HasMore() {
			pageNextAt = time.Now()
		}

		if err := fn(resp); err != nil {
			return err
		}
		if len(resp.Results) == 0 {
			return nil
		}
		if !resp.HasMore() {
			return nil
		}
		if len(resp.Results) < req.Size {
			return nil
		}

		nextToken = pageNext
		nextTokenAt = pageNextAt
	}
}

// searchNextWithRetry 对同一 next 游标重试，应对 FOFA 偶发网络/5xx/限速等问题
func (c *Client) searchNextWithRetry(req *NextRequest, nextToken string) (*NextResponse, error) {
	delays := c.retryDelays()
	maxAttempts := 1
	if len(delays) > 0 {
		maxAttempts = len(delays) + 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := contextErr(req.Ctx); err != nil {
			return nil, err
		}

		pageReq := *req
		pageReq.Next = nextToken

		resp, err := c.SearchNext(&pageReq)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if isNextExpiredErrmsg(err.Error()) {
			return nil, fmt.Errorf("next 游标已失效（有效期 %v）: %w", DefaultNextTokenTTL, err)
		}
		if !isRetryableNextErr(err) || attempt >= maxAttempts-1 {
			return nil, err
		}
		if attempt < len(delays) {
			time.Sleep(delays[attempt])
		}
	}

	return nil, lastErr
}

func checkNextTokenExpired(nextToken string, nextTokenAt time.Time) error {
	if nextToken == "" || nextTokenAt.IsZero() {
		return nil
	}
	ttl := DefaultNextTokenTTL - nextTokenSafetyMargin
	if time.Since(nextTokenAt) >= ttl {
		return fmt.Errorf("next 游标已过期（有效期 %v），请重新发起查询", DefaultNextTokenTTL)
	}
	return nil
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
