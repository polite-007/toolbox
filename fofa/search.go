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

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Ctx    context.Context // 上下文
	Query  string          // 搜索查询语句
	Size   int             // 返回记录数，默认100，最大10000
	Page   int             // 页码，默认1
	Fields string          // 返回字段，多个字段用逗号分隔，如: ip,port,host
}

// SearchResponse 搜索响应结果
type SearchResponse struct {
	Error   bool       `json:"error"`   // 是否有错误
	ErrMsg  string     `json:"errmsg"`  // 错误信息
	Size    int        `json:"size"`    // 当前返回的记录数
	Page    int        `json:"page"`    // 当前页码
	Mode    string     `json:"mode"`    // 查询模式
	Query   string     `json:"query"`   // 查询语句
	Results [][]string `json:"results"` // 搜索结果，二维数组
	Fields  []string   // 字段名列表，用于映射 Results 的列
}

// SearchResult 单条搜索结果，直接映射 FOFA API 返回的所有字段
type SearchResult struct {
	IP              string `json:"ip"`               // ip地址
	Port            string `json:"port"`             // 端口
	Protocol        string `json:"protocol"`         // 协议名
	Country         string `json:"country"`          // 国家代码
	CountryName     string `json:"country_name"`     // 国家名
	Region          string `json:"region"`           // 区域
	City            string `json:"city"`             // 城市
	Longitude       string `json:"longitude"`        // 地理位置 经度
	Latitude        string `json:"latitude"`         // 地理位置 纬度
	ASN             string `json:"asn"`              // asn编号
	FID             string `json:"fid"`              // 资产指纹
	Org             string `json:"org"`              // asn组织
	Host            string `json:"host"`             // 主机名
	Domain          string `json:"domain"`           // 域名
	OS              string `json:"os"`               // 操作系统
	Server          string `json:"server"`           // 网站server
	ICP             string `json:"icp"`              // icp备案号
	Title           string `json:"title"`            // 网站标题
	Jarm            string `json:"jarm"`             // jarm 指纹
	Header          string `json:"header"`           // 网站header
	Banner          string `json:"banner"`           // 协议 banner
	Cert            string `json:"cert"`             // 证书
	BaseProtocol    string `json:"base_protocol"`    // 基础协议，比如tcp/udp
	Link            string `json:"link"`             // 资产的URL链接
	CertIssuerOrg   string `json:"cert.issuer.org"`  // 证书颁发者组织
	CertIssuerCN    string `json:"cert.issuer.cn"`   // 证书颁发者通用名称
	CertSubjectOrg  string `json:"cert.subject.org"` // 证书持有者组织
	CertSubjectCN   string `json:"cert.subject.cn"`  // 证书持有者通用名称
	TLSJa3s         string `json:"tls.ja3s"`         // ja3s指纹信息
	TLSVersion      string `json:"tls.version"`      // tls协议版本
	CertSN          string `json:"cert.sn"`          // 证书的序列号
	CertNotBefore   string `json:"cert.not_before"`  // 证书生效时间
	CertNotAfter    string `json:"cert.not_after"`   // 证书到期时间
	CertDomain      string `json:"cert.domain"`      // 证书中的根域名
	HeaderHash      string `json:"header_hash"`      // http/https相应信息计算的hash值
	BannerHash      string `json:"banner_hash"`      // 协议相应信息的完整hash值
	BannerFid       string `json:"banner_fid"`       // 协议相应信息架构的指纹值
	Cname           string `json:"cname"`            // 域名cname
	Lastupdatetime  string `json:"lastupdatetime"`   // FOFA最后更新时间
	Product         string `json:"product"`          // 产品名
	ProductCategory string `json:"product_category"` // 产品分类
	ProductVersion  string `json:"product.version"`  // 产品版本号
	IconHash        string `json:"icon_hash"`        // 返回的icon_hash值
	CertIsValid     string `json:"cert.is_valid"`    // 证书是否有效
	CnameDomain     string `json:"cname_domain"`     // cname的域名
	Body            string `json:"body"`             // 网站正文内容
	CertIsMatch     string `json:"cert.is_match"`    // 证书颁发者和持有者是否相同
	CertIsEqual     string `json:"cert.is_equal"`    // 证书和域名是否匹配
	Icon            string `json:"icon"`             // icon 图标
	Fid2            string `json:"fid2"`             // fid2
	Structinfo      string `json:"structinfo"`       // 结构化信息
	Org2            string `json:"org2"`             // 归属机构
	Ptag            string `json:"ptag"`             // 产品标签
}

// GetContentByFields 根据字段名尝试获取内容
// fields: 字段名，多个字段用逗号分隔，如: "ip,port" 或 "ip"
func (s *SearchResult) GetContentByFields(fields string) []string {
	if fields == "" {
		return nil
	}

	// 分割字段名
	fieldList := strings.Split(fields, ",")
	values := make([]string, 0, len(fieldList))

	for _, field := range fieldList {
		field = strings.TrimSpace(field) // 去除空格
		value := getFieldValue(s, field)
		values = append(values, value)
	}

	return values
}


// GetResults 将Results转化成可直接使用的结构体
func (s *SearchResponse) GetResults() []*SearchResult {
	records := make([]*SearchResult, 0, len(s.Fields))
	for _, res := range s.Results {
		records = append(records, newRecordFromStrings(res, s.Fields))
	}
	return records
}


// Search 执行搜索查询
func (c *Client) Search(req *SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return nil, errors.New("查询语句不能为空")
	}

	// 设置默认值
	if req.Size <= 0 {
		req.Size = 500
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Fields == "" {
		req.Fields = "ip,port,host"
	}

	// 检查查询字段不得少于2个
	if len(strings.Split(req.Fields, ",")) < 2 {
		if req.Fields == "ip" {
			req.Fields += ",port"
		} else {
			req.Fields += ",ip"
		}
	}

	// 对查询语句进行 base64 编码
	qbase64 := base64.StdEncoding.EncodeToString([]byte(req.Query))

	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/api/v1/search/all", c.baseURL)
	params := url.Values{}
	params.Set("key", c.key)
	params.Set("qbase64", qbase64)
	params.Set("size", fmt.Sprintf("%d", req.Size))
	params.Set("page", fmt.Sprintf("%d", req.Page))
	params.Set("fields", req.Fields)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 设置 context 超时
	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	// 发送请求
	body, err := c.do(ctx, fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "搜索请求失败")
	}

	// 解析 JSON 响应
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, errors.Wrap(err, "解析响应 JSON 失败")
	}

	// 检查 API 返回的错误
	if searchResp.Error {
		return nil, fmt.Errorf("FOFA API 错误: %s", searchResp.ErrMsg)
	}

	// 保存字段名列表，用于后续的 GetResults() 方法
	fields := strings.Split(req.Fields, ",")
	searchResp.Fields = make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			searchResp.Fields = append(searchResp.Fields, field)
		}
	}

	return &searchResp, nil
}

