package fofa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// AccountResponse 账号信息响应结果
type AccountResponse struct {
	Error           bool   `json:"error"`             // 是否有错误
	ErrMsg          string `json:"errmsg"`            // 错误信息
	Email           string `json:"email"`             // 账号邮箱
	Username        string `json:"username"`          // 用户名
	Category        string `json:"category"`          // 用户类别
	Fcoin           int    `json:"fcoin"`             // F币余额
	FofaPoint       int    `json:"fofa_point"`        // F点数
	RemainFreePoint int    `json:"remain_free_point"` // 剩余免费点数
	RemainAPIQuery  int    `json:"remain_api_query"`  // 剩余API查询次数
	RemainAPIData   int    `json:"remain_api_data"`   // 剩余API数据条数
	IsVip           bool   `json:"isvip"`             // 是否是VIP
	VipLevel        int    `json:"vip_level"`         // VIP等级
	IsVerified      bool   `json:"is_verified"`       // 是否已认证
	Avatar          string `json:"avatar"`            // 头像URL
	Message         string `json:"message"`           // 消息
	FofacliVer      string `json:"fofacli_ver"`       // fofacli版本
	FofaServer      bool   `json:"fofa_server"`       // 是否为FOFA服务端
}

// Account 获取当前账号信息
func (c *Client) Account(ctx context.Context) (*AccountResponse, error) {
	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/api/v1/info/my", c.baseURL)
	params := url.Values{}
	params.Set("key", c.key)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 设置 context 超时
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultAccountTimeout)
	defer cancel()

	// 发送请求
	body, err := c.do(ctx, fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "账号信息请求失败")
	}

	// 解析 JSON 响应
	var accountResp AccountResponse
	if err := json.Unmarshal(body, &accountResp); err != nil {
		return nil, errors.Wrap(err, "解析响应 JSON 失败")
	}

	// 检查 API 返回的错误
	if accountResp.Error {
		return nil, fmt.Errorf("FOFA API 错误: %s", accountResp.ErrMsg)
	}

	return &accountResp, nil
}
