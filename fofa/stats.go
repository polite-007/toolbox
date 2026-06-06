package fofa

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// StatsRequest 统计请求参数
type StatsRequest struct {
	Query  string // 搜索查询语句
	Size   int
	Fields string // 统计字段，如: protocol,port,country
}

// StatsResponse 统计响应结果
type StatsResponse struct {
	Error    bool                   `json:"error"`    // 是否有错误
	ErrMsg   string                 `json:"errmsg"`   // 错误信息
	Distinct map[string]int         `json:"distinct"` // 统计结果，key为字段值，value为数量
	Aggs     map[string]interface{} `json:"aggs"`     // 聚合结果（如果有）
	Field    string                 // 统计字段列表，用于标识统计的字段
	Size     int
}

// StatsResult 单条统计结果
type StatsResult struct {
	Name  string `json:"name"`  // 字段值
	Count int    `json:"count"` // 该值出现的数量
}

// Stats 执行统计查询
func (c *Client) Stats(req *StatsRequest) (*StatsResponse, error) {
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	if req.Fields == "" {
		return nil, errors.New("统计字段不能为空")
	}

	// 对查询语句进行 base64 编码
	qbase64 := base64.StdEncoding.EncodeToString([]byte(req.Query))

	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/search/stats", c.baseURL)
	params := url.Values{}
	// params.Set("email", c.email)
	params.Set("size", fmt.Sprintf("%d", req.Size))
	params.Set("key", c.key)
	params.Set("qbase64", qbase64)
	params.Set("fields", req.Fields)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 重试间隔：5秒、10秒、20秒
	retryDelays := []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second}
	var lastErr error

	// 执行请求，最多重试3次
	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		// 发送 HTTP 请求
		resp, err := c.httpClient.Get(fullURL)
		if err != nil {
			lastErr = errors.Wrap(err, "发送统计请求失败")
			if attempt < len(retryDelays) {
				time.Sleep(retryDelays[attempt])
				continue
			}
			return nil, lastErr
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = errors.Wrap(err, "读取响应失败")
			if attempt < len(retryDelays) {
				time.Sleep(retryDelays[attempt])
				continue
			}
			return nil, lastErr
		}

		// 检查 HTTP 状态码
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
			if attempt < len(retryDelays) {
				time.Sleep(retryDelays[attempt])
				continue
			}
			return nil, lastErr
		}

		// 解析 JSON 响应
		var statsResp StatsResponse
		if err := json.Unmarshal(body, &statsResp); err != nil {
			lastErr = errors.Wrap(err, "解析响应 JSON 失败")
			if attempt < len(retryDelays) {
				time.Sleep(retryDelays[attempt])
				continue
			}
			return nil, lastErr
		}

		// 检查 API 返回的错误
		if statsResp.Error {
			lastErr = fmt.Errorf("FOFA API 错误: %s", statsResp.ErrMsg)
			if attempt < len(retryDelays) {
				time.Sleep(retryDelays[attempt])
				continue
			}
			return nil, lastErr
		}

		// 保存字段名列表，用于后续的 GetResults() 方法
		statsResp.Field = req.Fields
		statsResp.Size = req.Size

		return &statsResp, nil
	}

	return nil, lastErr
}

// GetResults 从 Aggs 中获取统计结果
func (s *StatsResponse) GetResults(field string) []*StatsResult {
	// 确定查询字段
	queryField := s.Field
	if field != "" {
		queryField = field
	}

	// 从 Aggs 获取数据
	aggsData, exists := s.Aggs[queryField]
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

		if result.Name != "" {
			results = append(results, result)
		}
	}

	return results
}

func (s *StatsResponse) GetDistinct(field string) int {
	aggData, exists := s.Distinct[field]
	if !exists {
		return 0
	}

	return aggData
}
