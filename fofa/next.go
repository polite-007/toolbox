package fofa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
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
// 容错处理，因为fofa接口非常干，非常不稳定，因此增加冗余的错误处理，来对抗fofa的不稳定
func (c *Client) SearchNextAll(req *NextRequest, fn func(*NextResponse) error) error {
	if req == nil {
		return errors.New("请求参数不能为空")
	}
	if fn == nil {
		return errors.New("回调函数不能为空")
	}

	nextToken := req.Next
	for {
		pageReq := *req
		pageReq.Next = nextToken

		resp, err := c.SearchNext(&pageReq)
		if err != nil {
			return err
		}
		if len(resp.Results) == 0 {
			return nil
		}
		if err := fn(resp); err != nil {
			return err
		}
		// 下一页游标为空，则表示没有更多数据
		if !resp.HasMore() {
			return nil
		}
		// 数据数量小于预期配置的数量
		if len(resp.Results) < req.Size {
			return nil
		}
		nextToken = resp.Next
	}
}
