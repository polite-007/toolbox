package fofa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// StatsRequest 统计请求参数
type StatsRequest struct {
	Ctx    context.Context // 上下文
	Query  string          // 搜索查询语句
	Size   int
	Fields string // 统计字段，如: protocol,port,country
}

// StatsResponse 统计响应结果
type StatsResponse struct {
	Error    bool                   `json:"error"`    // 是否有错误
	ErrMsg   string                 `json:"errmsg"`   // 错误信息
	Distinct map[string]int         `json:"distinct"` // 唯一计数，key为字段名，value为去重后的数量
	Aggs     map[string]interface{} `json:"aggs"`     // 聚合结果，key为字段名，value为该字段的统计列表
	Field    string                 // 统计字段列表（逗号分隔），用于标识统计的字段
	Size     int
}

// Fields 返回统计字段列表，按逗号拆分并去除空白项
func (s *StatsResponse) Fields() []string {
	return parseFields(s.Field)
}

// parseFields 将逗号分隔的字段字符串拆分为列表
func parseFields(fields string) []string {
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}

// StatsResult 单条统计结果
type StatsResult struct {
	Name    string         `json:"name"`              // 字段值
	Count   int            `json:"count"`             // 该值出现的数量
	Regions []StatsSubItem `json:"regions,omitempty"` // country 聚合时下钻的区域/城市统计
}

// StatsSubItem 子聚合条目，用于 country 下钻的 region/city
type StatsSubItem struct {
	Name  string `json:"name"`  // 子项名称（区域/城市）
	Count int    `json:"count"` // 子项数量
}

// aggsKey 把请求字段名映射为 FOFA aggs 中实际的 key。
// FOFA 在查询 country 时，aggs 中的 key 实际为 countries（复数），
// 其余字段与请求字段同名。
func aggsKey(field string) string {
	switch field {
	case "country":
		return "countries"
	default:
		return field
	}
}

// Stats 执行统计查询
func (c *Client) Stats(req *StatsRequest) (*StatsResponse, error) {
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	if req.Fields == "" {
		return nil, errors.New("统计字段不能为空")
	}

	if req.Size < 1 || req.Size > 10000 {
		return nil, errors.New("统计数量必须在1到10000之间")
	}

	// 对查询语句进行 base64 编码
	qbase64 := base64.StdEncoding.EncodeToString([]byte(req.Query))

	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/api/v1/search/stats", c.baseURL)
	params := url.Values{}
	params.Set("key", c.key)
	params.Set("qbase64", qbase64)
	params.Set("size", fmt.Sprintf("%d", req.Size))
	params.Set("fields", req.Fields)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 设置 context 超时
	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultStatsTimeout)
	defer cancel()

	// 发送请求
	body, err := c.do(ctx, fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "统计请求失败")
	}

	// 解析 JSON 响应
	var statsResp StatsResponse
	if err := json.Unmarshal(body, &statsResp); err != nil {
		return nil, errors.Wrap(err, "解析响应 JSON 失败")
	}

	// 检查 API 返回的错误
	if statsResp.Error {
		return nil, fmt.Errorf("FOFA API 错误: %s", statsResp.ErrMsg)
	}

	// 保存字段名列表，用于后续的 GetResults() 方法
	statsResp.Field = req.Fields

	return &statsResp, nil
}

// GetResults 从 Aggs 中获取统计结果
func (s *StatsResponse) GetResults(field string) []*StatsResult {
	// 确定查询字段
	queryField := s.Field
	if field != "" {
		queryField = field
	}

	// FOFA 部分字段在 aggs 中的 key 与请求字段名不一致（如 country -> countries）
	aggKey := aggsKey(queryField)
	aggsData, exists := s.Aggs[aggKey]
	if !exists {
		return nil
	}

	// 解析 JSON 数组
	aggsArray, ok := aggsData.([]interface{})
	if !ok {
		return nil
	}

	results := make([]*StatsResult, 0, len(aggsArray))
	for _, item := range aggsArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		result := &StatsResult{}
		if v, ok := itemMap["name"]; ok {
			result.Name = fmt.Sprintf("%v", v)
		}
		if v, ok := itemMap["count"]; ok {
			if count, ok := v.(float64); ok {
				result.Count = int(count)
			}
		}

		// 解析 country 下钻的 regions（区域/城市）子聚合
		if v, ok := itemMap["regions"]; ok {
			if regions, ok := v.([]interface{}); ok {
				for _, region := range regions {
					regionMap, ok := region.(map[string]interface{})
					if !ok {
						continue
					}
					sub := StatsSubItem{}
					if rv, ok := regionMap["name"]; ok {
						sub.Name = fmt.Sprintf("%v", rv)
					}
					if rv, ok := regionMap["count"]; ok {
						if c, ok := rv.(float64); ok {
							sub.Count = int(c)
						}
					}
					if sub.Name != "" {
						result.Regions = append(result.Regions, sub)
					}
				}
			}
		}

		if result.Name != "" {
			results = append(results, result)
		}
	}

	return results
}

// GetAllResults 一次性返回所有统计字段的统计结果，key 为字段名，value 为该字段的统计列表。
// 当请求时传入多个字段（如 "protocol,port"）时，可一次取回全部聚合数据。
func (s *StatsResponse) GetAllResults() map[string][]*StatsResult {
	result := make(map[string][]*StatsResult, len(s.Aggs))
	for _, field := range s.Fields() {
		if r := s.GetResults(field); len(r) > 0 {
			result[field] = r
		}
	}
	return result
}

// GetAllDistinct 返回所有字段的唯一计数，直接返回 Distinct 映射。
// 调用方应当只读，不应修改返回的 map。
func (s *StatsResponse) GetAllDistinct() map[string]int {
	return s.Distinct
}

// GetDistinct 获取唯一计数
func (s *StatsResponse) GetDistinct(field string) int {
	aggData, exists := s.Distinct[field]
	if !exists {
		return 0
	}

	return aggData
}
