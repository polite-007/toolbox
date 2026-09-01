package ipinfo

// Info 表示一个 IP 的网络信息，是多个数据源返回结构的统一归一结果。
type Info struct {
	IP          string  // IP 地址
	Hostname    string  // 反向域名
	Country     string  // 国家名，如 "United States"
	CountryCode string  // 国家代码，如 "US"
	Region      string  // 州/省，如 "California"
	City        string  // 城市，如 "Mountain View"
	Latitude    float64 // 纬度
	Longitude   float64 // 经度
	Postal      string  // 邮编
	Timezone    string  // 时区，如 "America/Los_Angeles"
	Org         string  // 组织名，如 "Google LLC"
	ASN         string  // AS 号，如 "AS15169"
	ISP         string  // ISP
	Continent   string  // 大洲
	Source      string  // 数据来自哪个源
}
