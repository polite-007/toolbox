package dig

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DigResult DNS查询结果
type DigResult struct {
	Domain     string   `json:"domain"`          // 查询的域名
	RecordType string   `json:"record_type"`     // 记录类型：A, AAAA, MX, CNAME, NS, TXT等
	Records    []string `json:"records"`         // 查询结果列表
	TTL        int      `json:"ttl"`             // TTL值（如果可用）
	Nameserver string   `json:"nameserver"`      // 使用的DNS服务器
	QueryTime  int      `json:"query_time"`      // 查询耗时（毫秒）
	Error      string   `json:"error,omitempty"` // 错误信息
}

// Dig DNS查询工具
type Dig struct {
	Nameserver string        // DNS服务器地址，如 "8.8.8.8:53"，为空则使用系统默认
	Timeout    time.Duration // 查询超时时间，默认5秒
}

// NewDig 创建Dig实例
func NewDig(nameserver string, timeout time.Duration) *Dig {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Dig{
		Nameserver: nameserver,
		Timeout:    timeout,
	}
}

// getNameserver 获取DNS服务器地址
func (d *Dig) getNameserver() string {
	if d.Nameserver != "" {
		return d.Nameserver
	}
	return "system-default"
}

// normalizeNameserver 规范化DNS服务器地址，确保包含端口
func (d *Dig) normalizeNameserver() string {
	if d.Nameserver == "" {
		return ""
	}
	// 如果没有端口，添加默认端口53
	if !strings.Contains(d.Nameserver, ":") {
		return d.Nameserver + ":53"
	}
	return d.Nameserver
}

// queryWithDNS 使用指定的DNS服务器查询（使用miekg/dns库）
func (d *Dig) queryWithDNS(domain string, qtype uint16) (*dns.Msg, error) {
	nameserver := d.normalizeNameserver()
	if nameserver == "" {
		return nil, fmt.Errorf("DNS服务器地址未指定")
	}

	// 创建DNS客户端
	c := new(dns.Client)
	c.Timeout = d.Timeout

	// 创建DNS查询消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	// 发送查询
	r, _, err := c.Exchange(m, nameserver)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// QueryA 查询A记录（IPv4地址）
func (d *Dig) QueryA(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "A",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeA)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var ipv4s []string
		for _, answer := range msg.Answer {
			if a, ok := answer.(*dns.A); ok {
				ipv4s = append(ipv4s, a.A.String())
				if result.TTL == 0 {
					result.TTL = int(a.Hdr.Ttl)
				}
			}
		}
		result.Records = ipv4s
		return result, nil
	}

	// 否则使用系统默认DNS
	ips, err := net.LookupIP(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	var ipv4s []string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip.String())
		}
	}
	result.Records = ipv4s

	return result, nil
}

// QueryAAAA 查询AAAA记录（IPv6地址）
func (d *Dig) QueryAAAA(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "AAAA",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeAAAA)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var ipv6s []string
		for _, answer := range msg.Answer {
			if aaaa, ok := answer.(*dns.AAAA); ok {
				ipv6s = append(ipv6s, aaaa.AAAA.String())
				if result.TTL == 0 {
					result.TTL = int(aaaa.Hdr.Ttl)
				}
			}
		}
		result.Records = ipv6s
		return result, nil
	}

	// 否则使用系统默认DNS
	ips, err := net.LookupIP(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	var ipv6s []string
	for _, ip := range ips {
		if ip.To4() == nil {
			ipv6s = append(ipv6s, ip.String())
		}
	}
	result.Records = ipv6s

	return result, nil
}

// QueryMX 查询MX记录（邮件交换记录）
func (d *Dig) QueryMX(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "MX",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeMX)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var records []string
		for _, answer := range msg.Answer {
			if mx, ok := answer.(*dns.MX); ok {
				records = append(records, fmt.Sprintf("%d %s", mx.Preference, mx.Mx))
				if result.TTL == 0 {
					result.TTL = int(mx.Hdr.Ttl)
				}
			}
		}
		result.Records = records
		return result, nil
	}

	// 否则使用系统默认DNS
	mxRecords, err := net.LookupMX(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	var records []string
	for _, mx := range mxRecords {
		records = append(records, fmt.Sprintf("%d %s", mx.Pref, mx.Host))
	}
	result.Records = records

	return result, nil
}

// QueryCNAME 查询CNAME记录
func (d *Dig) QueryCNAME(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "CNAME",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeCNAME)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var records []string
		for _, answer := range msg.Answer {
			if cname, ok := answer.(*dns.CNAME); ok {
				records = append(records, cname.Target)
				if result.TTL == 0 {
					result.TTL = int(cname.Hdr.Ttl)
				}
			}
		}
		result.Records = records
		return result, nil
	}

	// 否则使用系统默认DNS
	cname, err := net.LookupCNAME(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Records = []string{cname}
	return result, nil
}

// QueryNS 查询NS记录（域名服务器记录）
func (d *Dig) QueryNS(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "NS",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeNS)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var records []string
		for _, answer := range msg.Answer {
			if ns, ok := answer.(*dns.NS); ok {
				records = append(records, ns.Ns)
				if result.TTL == 0 {
					result.TTL = int(ns.Hdr.Ttl)
				}
			}
		}
		result.Records = records
		return result, nil
	}

	// 否则使用系统默认DNS
	nsRecords, err := net.LookupNS(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	var records []string
	for _, ns := range nsRecords {
		records = append(records, ns.Host)
	}
	result.Records = records

	return result, nil
}

// QueryTXT 查询TXT记录
func (d *Dig) QueryTXT(domain string) (*DigResult, error) {
	start := time.Now()
	result := &DigResult{
		Domain:     domain,
		RecordType: "TXT",
		Nameserver: d.getNameserver(),
	}

	// 如果指定了DNS服务器，使用miekg/dns库查询
	if d.Nameserver != "" {
		msg, err := d.queryWithDNS(domain, dns.TypeTXT)
		queryTime := int(time.Since(start).Milliseconds())
		result.QueryTime = queryTime

		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if msg.Rcode != dns.RcodeSuccess {
			result.Error = dns.RcodeToString[msg.Rcode]
			return result, fmt.Errorf("DNS查询失败: %s", result.Error)
		}

		var records []string
		for _, answer := range msg.Answer {
			if txt, ok := answer.(*dns.TXT); ok {
				records = append(records, strings.Join(txt.Txt, " "))
				if result.TTL == 0 {
					result.TTL = int(txt.Hdr.Ttl)
				}
			}
		}
		result.Records = records
		return result, nil
	}

	// 否则使用系统默认DNS
	txtRecords, err := net.LookupTXT(domain)
	queryTime := int(time.Since(start).Milliseconds())
	result.QueryTime = queryTime

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Records = txtRecords
	return result, nil
}

// QueryPTR 查询PTR记录（反向DNS查询）
func (d *Dig) QueryPTR(ip string) (*DigResult, error) {
	start := time.Now()
	names, err := net.LookupAddr(ip)
	queryTime := int(time.Since(start).Milliseconds())

	result := &DigResult{
		Domain:     ip,
		RecordType: "PTR",
		QueryTime:  queryTime,
		Nameserver: d.getNameserver(),
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Records = names
	return result, nil
}

// Query 通用查询方法，根据记录类型自动选择查询方法
func (d *Dig) Query(domain string, recordType string) (*DigResult, error) {
	recordType = strings.ToUpper(recordType)

	switch recordType {
	case "A":
		return d.QueryA(domain)
	case "AAAA":
		return d.QueryAAAA(domain)
	case "MX":
		return d.QueryMX(domain)
	case "CNAME":
		return d.QueryCNAME(domain)
	case "NS":
		return d.QueryNS(domain)
	case "TXT":
		return d.QueryTXT(domain)
	default:
		return nil, fmt.Errorf("不支持的记录类型: %s", recordType)
	}
}

// FormatResult 格式化输出查询结果（类似dig命令的输出格式）
func (result *DigResult) FormatResult() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("; <<>> Dig Query <<>> %s %s\n", result.RecordType, result.Domain))
	sb.WriteString(";; global options: +cmd\n")
	sb.WriteString(";; Got answer:\n")

	if result.Error != "" {
		sb.WriteString(fmt.Sprintf(";; ERROR: %s\n", result.Error))
		return sb.String()
	}

	if len(result.Records) == 0 {
		sb.WriteString(";; No answer found\n")
		return sb.String()
	}

	sb.WriteString(";; ->>HEADER<<- opcode: QUERY, status: NOERROR\n")
	sb.WriteString(fmt.Sprintf(";; Query time: %d msec\n", result.QueryTime))
	if result.Nameserver != "" {
		if result.Nameserver == "system-default" {
			sb.WriteString(fmt.Sprintf(";; SERVER: %s (system default resolver)\n", result.Nameserver))
		} else {
			sb.WriteString(fmt.Sprintf(";; SERVER: %s\n", result.Nameserver))
		}
	} else {
		sb.WriteString(";; SERVER: system default resolver\n")
	}
	sb.WriteString(";;\n")

	for _, record := range result.Records {
		sb.WriteString(fmt.Sprintf("%s\t\tIN\t%s\t%s\n", result.Domain, result.RecordType, record))
	}

	return sb.String()
}

// GetMXIPs 获取MX记录对应的IP地址
func (d *Dig) GetMXIPs(domain string) ([]string, error) {
	mxResult, err := d.QueryMX(domain)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, mxRecord := range mxResult.Records {
		// 解析MX记录格式：优先级 域名
		parts := strings.Fields(mxRecord)
		if len(parts) < 2 {
			continue
		}
		mxDomain := parts[len(parts)-1]
		// 去掉末尾的点
		mxDomain = strings.TrimSuffix(mxDomain, ".")

		// 查询MX域名的A记录
		aResult, err := d.QueryA(mxDomain)
		if err != nil {
			continue
		}
		ips = append(ips, aResult.Records...)
	}

	return ips, nil
}

// CDNCheckResult CDN检测结果
type CDNCheckResult struct {
	IsCDN      bool     `json:"is_cdn"`     // 是否使用CDN
	Confidence float64  `json:"confidence"` // 置信度 0-100
	IPs        []string `json:"ips"`        // 解析到的IP列表
	Reasons    []string `json:"reasons"`    // 判断原因
	CDNType    string   `json:"cdn_type"`   // CDN类型（如果可识别）
}

// 已知CDN厂商的CNAME特征
var cdnCNAMEPatterns = map[string]string{
	"cloudflare": "cloudflare",
	"cloudfront": "cloudfront",
	"fastly":     "fastly",
	"akamai":     "akamai",
	"aliyun":     "aliyuncs",
	"tencent":    "qcloudcdn",
	"baidu":      "baidubce",
	"wangsu":     "wscloudcdn",
	"chinacache": "chinacache",
}

// IsCDN 综合判断是否使用CDN
// 使用多种方法综合判断，提高准确性
// 新增：使用国内和国际DNS汇总IP，如果IP数量>1则判定为CDN
// 返回值：是否CDN、置信度(0-100)、IP列表、判断原因
func (d *Dig) IsCDN(domain string) (*CDNCheckResult, error) {
	result := &CDNCheckResult{
		IPs:     []string{},
		Reasons: []string{},
	}

	score := 0.0 // 置信度分数

	// 1. 检查CNAME记录（最可靠的指标）
	cnameResult, err := d.QueryCNAME(domain)
	if err == nil && len(cnameResult.Records) > 0 {
		cname := strings.ToLower(cnameResult.Records[0])
		for cdnType, pattern := range cdnCNAMEPatterns {
			if strings.Contains(cname, pattern) {
				result.IsCDN = true
				result.CDNType = cdnType
				result.Reasons = append(result.Reasons, fmt.Sprintf("CNAME指向CDN域名: %s", cname))
				score += 50.0 // CNAME匹配是最强指标
				break
			}
		}
	}

	// 2. 使用多个DNS服务器（国内+国际）查询A记录并汇总IP
	allIPs := make(map[string]bool) // 使用map去重
	dnsServersUsed := []string{}

	// 获取国内DNS服务器
	chinaServers := GetDNSServersByRegion("china")
	// 获取国际DNS服务器
	globalServers := GetDNSServersByRegion("global")

	// 合并DNS服务器列表（限制数量，避免查询时间过长）
	allDNSServers := make(map[string]string)
	count := 0
	maxServers := 8 // 使用8个DNS服务器（4个国内+4个国际）

	// 添加国内DNS
	for name, addr := range chinaServers {
		if count >= maxServers/2 {
			break
		}
		allDNSServers[name] = addr
		count++
	}

	// 添加国际DNS
	for name, addr := range globalServers {
		if count >= maxServers {
			break
		}
		allDNSServers[name] = addr
		count++
	}

	// 使用多个DNS服务器查询并汇总IP
	successCount := 0
	for name, dnsAddr := range allDNSServers {
		// 创建临时Dig实例使用指定DNS服务器
		tempDig := NewDig(dnsAddr, 3*time.Second) // 使用较短的超时时间
		aResult, err := tempDig.QueryA(domain)
		if err != nil {
			// 查询失败，跳过
			continue
		}

		successCount++
		dnsServersUsed = append(dnsServersUsed, fmt.Sprintf("%s(%s)", name, dnsAddr))

		// 汇总IP地址
		for _, ip := range aResult.Records {
			allIPs[ip] = true
		}
	}

	// 将map转换为slice
	for ip := range allIPs {
		result.IPs = append(result.IPs, ip)
	}

	// 3. 多DNS服务器多IP检测（强指标）
	if len(result.IPs) > 1 {
		score += 40.0 // 提高权重，因为多DNS服务器返回不同IP是CDN的强指标
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("通过%d个DNS服务器查询，汇总得到%d个不同IP地址: %v",
				successCount, len(result.IPs), result.IPs))
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("使用的DNS服务器: %v", dnsServersUsed))
		result.IsCDN = true
	} else if len(result.IPs) == 1 {
		// 单IP，但可能是CDN（如果CNAME指向CDN）
		if !result.IsCDN {
			score -= 10.0
			result.Reasons = append(result.Reasons,
				fmt.Sprintf("通过%d个DNS服务器查询，均返回相同IP: %s",
					successCount, result.IPs[0]))
		}
	}

	// 4. IP数量异常多（通常CDN会有很多IP）
	if len(result.IPs) > 5 {
		score += 15.0
		result.Reasons = append(result.Reasons, "解析到异常多的IP地址")
	}

	// 5. 检查NS记录（CDN通常使用CDN厂商的NS服务器）
	nsResult, err := d.QueryNS(domain)
	if err == nil {
		for _, ns := range nsResult.Records {
			nsLower := strings.ToLower(ns)
			for cdnType, pattern := range cdnCNAMEPatterns {
				if strings.Contains(nsLower, pattern) {
					score += 10.0
					result.Reasons = append(result.Reasons, fmt.Sprintf("NS记录包含CDN特征: %s", ns))
					if result.CDNType == "" {
						result.CDNType = cdnType
					}
					break
				}
			}
		}
	}

	// 计算最终置信度
	result.Confidence = score
	if score >= 30.0 {
		result.IsCDN = true
	} else if score < 0 {
		result.Confidence = 0
	}

	// 如果通过多DNS服务器查询得到多个IP，即使其他指标不明显，也判定为CDN
	if len(result.IPs) > 1 && successCount >= 2 {
		result.IsCDN = true
		if result.Confidence < 30.0 {
			result.Confidence = 35.0 // 设置最低置信度
		}
		result.Reasons = append(result.Reasons,
			"多DNS服务器返回不同IP，判定为CDN")
	}

	return result, nil
}
