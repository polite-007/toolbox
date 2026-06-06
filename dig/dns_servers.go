package dig

// PublicDNSServers 公共DNS服务器列表
// 包含全球常用的公共DNS服务器地址
var PublicDNSServers = map[string]string{
	// Google DNS
	"google-primary":   "8.8.8.8",
	"google-secondary": "8.8.4.4",
	"google-ipv6-1":    "2001:4860:4860::8888",
	"google-ipv6-2":    "2001:4860:4860::8844",

	// Cloudflare DNS
	"cloudflare-primary":   "1.1.1.1",
	"cloudflare-secondary": "1.0.0.1",
	"cloudflare-ipv6-1":    "2606:4700:4700::1111",
	"cloudflare-ipv6-2":    "2606:4700:4700::1001",

	// Quad9 DNS
	"quad9-primary":   "9.9.9.9",
	"quad9-secondary": "149.112.112.112",
	"quad9-ipv6-1":    "2620:fe::fe",
	"quad9-ipv6-2":    "2620:fe::9",

	// OpenDNS (Cisco)
	"opendns-primary":   "208.67.222.222",
	"opendns-secondary": "208.67.220.220",
	"opendns-ipv6-1":    "2620:0:ccc::2",
	"opendns-ipv6-2":    "2620:0:ccd::2",

	// AdGuard DNS
	"adguard-default":     "94.140.14.14",
	"adguard-family":      "94.140.14.15",
	"adguard-nofilter":    "94.140.14.140",
	"adguard-family-ipv6": "2a10:50c0::ad1:ff",

	// 114 DNS (中国)
	"114-primary":   "114.114.114.114",
	"114-secondary": "114.114.115.115",

	// 阿里云DNS (中国)
	"aliyun-primary":   "223.5.5.5",
	"aliyun-secondary": "223.6.6.6",
	"aliyun-ipv6-1":    "2400:3200::1",
	"aliyun-ipv6-2":    "2400:3200:baba::1",

	// 腾讯DNS (中国)
	"tencent-primary":   "119.29.29.29",
	"tencent-secondary": "182.254.116.116",
	"tencent-ipv6":      "2402:4e00::",

	// 百度DNS (中国)
	"baidu-primary": "180.76.76.76",

	// 360 DNS (中国)
	"360-primary":   "101.226.4.6",
	"360-secondary": "123.125.81.6",

	// DNSPod (中国)
	"dnspod-primary":   "119.29.29.29",
	"dnspod-secondary": "182.254.116.116",

	// Level3 DNS
	"level3-primary":   "4.2.2.1",
	"level3-secondary": "4.2.2.2",

	// Comodo Secure DNS
	"comodo-primary":   "8.26.56.26",
	"comodo-secondary": "8.20.247.20",

	// Verisign DNS
	"verisign-primary":   "64.6.64.6",
	"verisign-secondary": "64.6.65.6",

	// Yandex DNS (俄罗斯)
	"yandex-primary":   "77.88.8.8",
	"yandex-secondary": "77.88.8.1",
	"yandex-safe":      "77.88.8.88",
	"yandex-family":    "77.88.8.2",

	// OpenNIC DNS
	"opennic-primary":   "185.121.177.177",
	"opennic-secondary": "169.239.202.202",

	// UncensoredDNS (丹麦)
	"uncensored-primary":   "91.239.100.100",
	"uncensored-secondary": "89.233.43.71",
	"uncensored-ipv6-1":    "2001:67c:28a4::",
	"uncensored-ipv6-2":    "2a01:3a0:53:53::",

	// NextDNS (需要配置)
	"nextdns-primary":   "45.90.28.0",
	"nextdns-secondary": "45.90.30.0",

	// CleanBrowsing (安全DNS)
	"cleanbrowsing-family":   "185.228.168.168",
	"cleanbrowsing-adult":    "185.228.168.10",
	"cleanbrowsing-security": "185.228.168.9",

	// Neustar DNS
	"neustar-primary":   "156.154.70.1",
	"neustar-secondary": "156.154.71.1",

	// SafeDNS
	"safedns-primary":   "195.46.39.39",
	"safedns-secondary": "195.46.39.40",
}

// GetDNSServer 根据名称获取DNS服务器地址
func GetDNSServer(name string) (string, bool) {
	server, ok := PublicDNSServers[name]
	return server, ok
}

// GetAllDNSServers 获取所有DNS服务器列表
func GetAllDNSServers() map[string]string {
	return PublicDNSServers
}

// GetDNSServersByRegion 根据地区获取DNS服务器
func GetDNSServersByRegion(region string) map[string]string {
	servers := make(map[string]string)

	switch region {
	case "global", "international":
		// 国际通用DNS
		servers["google-primary"] = PublicDNSServers["google-primary"]
		servers["google-secondary"] = PublicDNSServers["google-secondary"]
		servers["cloudflare-primary"] = PublicDNSServers["cloudflare-primary"]
		servers["cloudflare-secondary"] = PublicDNSServers["cloudflare-secondary"]
		servers["quad9-primary"] = PublicDNSServers["quad9-primary"]
		servers["opendns-primary"] = PublicDNSServers["opendns-primary"]

	case "china", "cn":
		// 中国DNS
		servers["aliyun-primary"] = PublicDNSServers["aliyun-primary"]
		servers["aliyun-secondary"] = PublicDNSServers["aliyun-secondary"]
		servers["tencent-primary"] = PublicDNSServers["tencent-primary"]
		servers["114-primary"] = PublicDNSServers["114-primary"]
		servers["baidu-primary"] = PublicDNSServers["baidu-primary"]

	case "privacy":
		// 隐私保护DNS
		servers["quad9-primary"] = PublicDNSServers["quad9-primary"]
		servers["cloudflare-primary"] = PublicDNSServers["cloudflare-primary"]
		servers["adguard-default"] = PublicDNSServers["adguard-default"]
		servers["uncensored-primary"] = PublicDNSServers["uncensored-primary"]

	case "security":
		// 安全DNS
		servers["quad9-primary"] = PublicDNSServers["quad9-primary"]
		servers["cleanbrowsing-security"] = PublicDNSServers["cleanbrowsing-security"]
		servers["comodo-primary"] = PublicDNSServers["comodo-primary"]

	case "family":
		// 家庭/儿童保护DNS
		servers["cleanbrowsing-family"] = PublicDNSServers["cleanbrowsing-family"]
		servers["adguard-family"] = PublicDNSServers["adguard-family"]
		servers["yandex-family"] = PublicDNSServers["yandex-family"]

	case "adblock":
		// 广告拦截DNS
		servers["adguard-default"] = PublicDNSServers["adguard-default"]
		servers["adguard-family"] = PublicDNSServers["adguard-family"]
	}

	return servers
}

// DNSStats DNS服务器统计信息
type DNSStats struct {
	TotalServers    int `json:"total_servers"`
	IPv4Servers     int `json:"ipv4_servers"`
	IPv6Servers     int `json:"ipv6_servers"`
	ChinaServers    int `json:"china_servers"`
	GlobalServers   int `json:"global_servers"`
	PrivacyServers  int `json:"privacy_servers"`
	SecurityServers int `json:"security_servers"`
}

// GetDNSStats 获取DNS服务器统计信息
func GetDNSStats() DNSStats {
	stats := DNSStats{
		TotalServers: len(PublicDNSServers),
	}

	// 统计IPv6服务器
	ipv6Count := 0
	for _, server := range PublicDNSServers {
		if len(server) > 4 && server[:4] != "127." && len(server) > 15 {
			ipv6Count++
		}
	}
	stats.IPv6Servers = ipv6Count
	stats.IPv4Servers = stats.TotalServers - stats.IPv6Servers

	// 统计中国DNS
	chinaServers := GetDNSServersByRegion("china")
	stats.ChinaServers = len(chinaServers)

	// 统计国际DNS
	globalServers := GetDNSServersByRegion("global")
	stats.GlobalServers = len(globalServers)

	// 统计隐私保护DNS
	privacyServers := GetDNSServersByRegion("privacy")
	stats.PrivacyServers = len(privacyServers)

	// 统计安全DNS
	securityServers := GetDNSServersByRegion("security")
	stats.SecurityServers = len(securityServers)

	return stats
}
