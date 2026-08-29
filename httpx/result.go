package httpx

import (
	"time"

	"github.com/projectdiscovery/httpx/runner"
)

// Result 表示一条 HTTP 探测结果，封装了上游 runner.Result 的常用字段，
// 避免业务代码直接依赖原生 httpx 库类型。
type Result struct {
	Timestamp     time.Time // 探测时间
	Input         string    // 原始输入（host:port 或 URL）
	URL           string    // 探测的 URL
	FinalURL      string    // 重定向后的最终 URL
	Scheme        string    // http/https
	Host          string    // Host 头
	HostIP        string    // 目标 IP
	Port          string    // 端口
	Path          string    // 请求路径
	Method        string    // 请求方法
	StatusCode    int       // HTTP 状态码
	Title         string    // 页面标题
	WebServer     string    // Server 头
	ContentType   string    // Content-Type
	ContentLength int       // 响应体长度
	Location      string    // 重定向 Location
	ResponseTime  string    // 响应耗时
	Technologies  []string  // 识别到的技术栈
	CDN           bool      // 是否 CDN
	CDNName       string    // CDN 名称
	CDNType       string    // CDN 类型
	HTTP2         bool      // 是否 HTTP/2
	WebSocket     bool      // 是否 WebSocket
	Pipeline      bool      // 是否 HTTP pipeline
	VHost         bool      // 是否虚拟主机
	FavIconMMH3   string    // favicon mmh3 hash
	FavIconMD5    string    // favicon md5 hash
	JarmHash      string    // JARM 指纹
	A             []string  // A 记录
	AAAA          []string  // AAAA 记录
	CNAMEs        []string  // CNAME 记录
	Error         string    // 错误信息字符串
	Failed        bool      // 是否探测失败
	Err           error     // 探测错误
	ResponseBody  string    // 响应体（截取前 4KB）
	SubjectCN     string    // TLS 证书 Subject Common Name
	SubjectOrg    []string  // TLS 证书 Subject Organization
}

const maxResponseBodyPreview = 4 * 1024

func toResult(native runner.Result) Result {
	result := Result{
		Timestamp:     native.Timestamp,
		Input:         native.Input,
		URL:           native.URL,
		FinalURL:      native.FinalURL,
		Scheme:        native.Scheme,
		Host:          native.Host,
		HostIP:        native.HostIP,
		Port:          native.Port,
		Path:          native.Path,
		Method:        native.Method,
		StatusCode:    native.StatusCode,
		Title:         native.Title,
		WebServer:     native.WebServer,
		ContentType:   native.ContentType,
		ContentLength: native.ContentLength,
		Location:      native.Location,
		ResponseTime:  native.ResponseTime,
		Technologies:  append([]string(nil), native.Technologies...),
		CDN:           native.CDN,
		CDNName:       native.CDNName,
		CDNType:       native.CDNType,
		HTTP2:         native.HTTP2,
		WebSocket:     native.WebSocket,
		Pipeline:      native.Pipeline,
		VHost:         native.VHost,
		FavIconMMH3:   native.FavIconMMH3,
		FavIconMD5:    native.FavIconMD5,
		JarmHash:      native.JarmHash,
		A:             append([]string(nil), native.A...),
		AAAA:          append([]string(nil), native.AAAA...),
		CNAMEs:        append([]string(nil), native.CNAMEs...),
		Error:         native.Error,
		Failed:        native.Failed,
		Err:           native.Err,
	}

	// 响应体：从 Response.Data 截取前 4KB
	if native.Response != nil && len(native.Response.Data) > 0 {
		data := native.Response.Data
		if len(data) > maxResponseBodyPreview {
			data = data[:maxResponseBodyPreview]
		}
		result.ResponseBody = string(data)
	}

	// TLS 证书 Subject 信息
	if native.TLSData != nil {
		result.SubjectCN = native.TLSData.SubjectCN
		result.SubjectOrg = append([]string(nil), native.TLSData.SubjectOrg...)
	}

	return result
}
