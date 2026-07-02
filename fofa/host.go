package fofa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// HostRequest Host 聚合请求参数
type HostRequest struct {
	Ctx    context.Context // 上下文
	Host   string          // host 名，通常是 ip，如 78.48.50.249
	Detail bool            // 是否显示端口详情，true 返回每个端口的协议及产品详情
}

// HostResponse Host 聚合响应结果
//
// 普通模式（Detail=false）会填充 Protocol/Port/Category/Product 等列表字段；
// 详情模式（Detail=true）会填充 Ports 列表（含每个端口的产品详情）。
// 两种模式共享的基础字段（IP/ASN/Org/CountryName/CountryCode/UpdateTime/Host）都会返回。
type HostResponse struct {
	Error       bool   `json:"error"`        // 是否有错误
	ErrMsg      string `json:"errmsg"`       // 错误信息
	Host        string `json:"host"`         // 查询的 host
	IP          string `json:"ip"`           // ip 地址
	ASN         int    `json:"asn"`          // asn 编号
	Org         string `json:"org"`          // asn 组织
	CountryName string `json:"country_name"` // 国家名
	CountryCode string `json:"country_code"` // 国家代码

	// 普通模式（detail=false）字段
	Protocol []string `json:"protocol"` // 协议列表
	Port     []int    `json:"port"`     // 端口列表
	Category []string `json:"category"` // 分类标签列表
	Product  []string `json:"product"`  // 产品标签列表

	// 详情模式（detail=true）字段
	Ports []HostPort `json:"ports"` // 端口详情列表

	UpdateTime string `json:"update_time"` // FOFA 最后更新时间

	// 以下为 API 计费相关字段，按需返回
	ConsumedFpoint  int `json:"consumed_fpoint"`
	RequiredFpoints int `json:"required_fpoints"`
}

// HostPort 详情模式下的单个端口信息
type HostPort struct {
	Port     int           `json:"port"`     // 端口
	Protocol string        `json:"protocol"` // 协议
	Products []HostProduct `json:"products"` // 该端口上的产品详情列表
}

// HostProduct 详情模式下的产品信息
type HostProduct struct {
	Product      string `json:"product"`        // 产品名
	Category     string `json:"category"`       // 产品分类
	Level        int    `json:"level"`          // 产品分层：5 应用层,4 支持层,3 服务层,2 系统层,1 硬件层,0 无组件分层
	SortHardCode int    `json:"sort_hard_code"` // 是否为硬件：1 为硬件，否则为非硬件
}

// Host 根据 host（通常是 ip）获取聚合信息
//
// 接口：GET /api/v1/host/{host}
// 限制：请求并发为 1s/次
func (c *Client) Host(req *HostRequest) (*HostResponse, error) {
	host := strings.TrimSpace(req.Host)
	if host == "" {
		return nil, errors.New("host 不能为空")
	}

	// host 作为 path 的一部分，需要转义以避免特殊字符破坏 URL
	apiURL := fmt.Sprintf("%s/api/v1/host/%s", c.baseURL, url.PathEscape(host))
	params := url.Values{}
	params.Set("key", c.key)
	params.Set("detail", fmt.Sprintf("%t", req.Detail))

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 设置 context 超时
	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultHostTimeout)
	defer cancel()

	// 发送请求
	body, err := c.do(ctx, fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "Host 聚合请求失败")
	}

	// 解析 JSON 响应
	var hostResp HostResponse
	if err := json.Unmarshal(body, &hostResp); err != nil {
		return nil, errors.Wrap(err, "解析响应 JSON 失败")
	}

	// 检查 API 返回的错误
	if hostResp.Error {
		return nil, fmt.Errorf("FOFA API 错误: %s", hostResp.ErrMsg)
	}

	return &hostResp, nil
}
