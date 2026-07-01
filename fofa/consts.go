package fofa

import (
	"strings"
)

// ErrCodeRateLimit FOFA 请求过快错误码
const ErrCodeRateLimit = "45012"

// isRateLimitErrmsg 判断 FOFA 业务层 errmsg 是否为请求过快
func isRateLimitErrmsg(msg string) bool {
	return strings.Contains(msg, ErrCodeRateLimit) ||
		strings.Contains(msg, "请求速度过快") ||
		strings.Contains(msg, "请求太快")
}

// isNextExpiredErrmsg 判断 errmsg 是否表示 next 游标已失效
func isNextExpiredErrmsg(msg string) bool {
	return strings.Contains(msg, "过期") ||
		strings.Contains(msg, "失效") ||
		strings.Contains(msg, "游标") ||
		strings.Contains(msg, "nextid")
}

// isRetryableNextErr 判断连续翻页过程中是否值得对同一 next 游标重试
func isRetryableNextErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if isNextExpiredErrmsg(msg) {
		return false
	}
	if isRateLimitErrmsg(msg) {
		return true
	}
	if strings.Contains(msg, "发送请求失败") ||
		strings.Contains(msg, "读取响应失败") ||
		strings.Contains(msg, "连续翻页请求失败") ||
		strings.Contains(msg, "解析响应 JSON 失败") {
		return true
	}
	if strings.Contains(msg, "状态码: 5") {
		return true
	}
	return false
}

// SearchFieldsFree 免费版可用的查询字段（34个）
const SearchFieldsFree = "ip,port,protocol,country,country_name,region,city," +
	"longitude,latitude,asn,org,host,domain,os,server," +
	"icp,title,jarm,header,banner,cert,base_protocol,link," +
	"cert.issuer.org,cert.issuer.cn,cert.subject.org,cert.subject.cn," +
	"tls.ja3s,tls.version,cert.sn,cert.not_before,cert.not_after," +
	"cert.domain,status_code"

// SearchFieldsPersonal 个人版及以上可用的查询字段（37个，含免费版）
const SearchFieldsPersonal = SearchFieldsFree +
	",header_hash,banner_hash,banner_fid"

// SearchFieldsProfessional 专业版及以上可用的查询字段（41个，含个人版）
const SearchFieldsProfessional = SearchFieldsPersonal +
	",cname,lastupdatetime,product,product_category"

// SearchFieldsBusiness 商业版及以上可用的查询字段（48个，含专业版）
const SearchFieldsBusiness = SearchFieldsProfessional +
	",product.version,icon_hash,cert.is_valid,cname_domain," +
	"body,cert.is_match,cert.is_equal"

// SearchFieldsEnterprise 企业版可用的查询字段（51个，含商业版）
const SearchFieldsEnterprise = SearchFieldsBusiness +
	",icon,fid,structinfo"

const SearchFieldsFreeCN = "ip地址,端口,协议名,国家代码,国家名,区域,城市," +
	"地理位置 经度,地理位置 纬度,asn编号,asn组织,主机名,域名,操作系统,网站server," +
	"icp备案号,网站标题,jarm 指纹,网站header,协议 banner,证书,基础协议 比如tcp/udp,资产URL链接," +
	"证书颁发者组织,证书颁发者通用名称,证书持有者组织,证书持有者通用名称," +
	"ja3s指纹信息,tls协议版本,证书序列号,证书生效时间,证书到期时间," +
	"证书中的根域名,http 状态码"

const SearchFieldsPersonalCN = SearchFieldsFreeCN +
	",http/https相应信息计算的hash值,协议相应信息的完整hash值,协议相应信息架构的指纹值,"

const SearchFieldsProfessionalCN = SearchFieldsPersonalCN +
	"域名cname,FOFA最后更新时间,产品名,产品分类"

const SearchFieldsBusinessCN = SearchFieldsProfessionalCN +
	",产品版本号,返回的icon_hash值,证书是否有效,cname的域名," +
	"网站正文内容,证书颁发者和持有者是否相同,证书和域名是否匹配"

const SearchFieldsEnterpriseCN = SearchFieldsBusinessCN +
	",icon 图标,fid,结构化信息"
