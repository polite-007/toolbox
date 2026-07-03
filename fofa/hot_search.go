package fofa

import (
	"encoding/json"
	"io"
)

type HotTrend struct {
	DateE  string `json:"date`   // 2026-07-01
	Value  int    `json:"value"` // 搜索次数
	IsMark bool   `json:"is_mark"`
}

type HotSearchData struct {
	App         string      `json:"app"`          // 应用名称
	QueryString string      `json:"query_string"` // 搜索查询语句
	Asset_count int         `json:"asset_count"`  // 资产数量
	IsHot       bool        `json:"is_hot"`       // 是否为热门搜索
	HotTrend    []*HotTrend `json:"hot_trend"`    // 热门搜索趋势
}

type HotSearchResponse struct {
	Code int              `json:"code"`    // 状态码 0 成功 非0 失败
	Msg  string           `json:"message"` // 错误信息
	Data []*HotSearchData `json:"data"`    // 热门搜索数据
}

func (c *Client) HotSearch() (*HotSearchResponse, error) {
	url := "https://api.fofa.info/v1/hotsearch"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var hotSearchResponse HotSearchResponse
	err = json.Unmarshal(body, &hotSearchResponse)
	if err != nil {
		return nil, err
	}
	return &hotSearchResponse, nil
}
