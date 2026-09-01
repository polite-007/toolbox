package ipinfo

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// source 表示一个 IP 信息数据源。
type source struct {
	name  string
	url   func(ip string) string
	parse func(body []byte) (Info, error)
}

var asnRe = regexp.MustCompile(`^(AS\d+)\s`)

func formatASN(asn interface{}) string {
	switch v := asn.(type) {
	case float64:
		if v > 0 {
			return fmt.Sprintf("AS%d", int64(v))
		}
	case int:
		if v > 0 {
			return fmt.Sprintf("AS%d", v)
		}
	case int64:
		if v > 0 {
			return fmt.Sprintf("AS%d", v)
		}
	case string:
		// 形如 "AS15169" 或 "15169"
		if v == "" {
			return ""
		}
		if strings.HasPrefix(strings.ToUpper(v), "AS") {
			return strings.ToUpper(v)
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return fmt.Sprintf("AS%d", n)
		}
	}
	return ""
}

// parseLoc 解析 "lat,lng" 形式的经纬度字符串。
func parseLoc(loc string) (lat, lng float64) {
	parts := strings.Split(loc, ",")
	if len(parts) != 2 {
		return 0, 0
	}
	lat, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return lat, lng
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// --- ipinfo.io ---

type ipinfoResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

func parseIpinfo(body []byte) (Info, error) {
	var r ipinfoResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{}, err
	}
	lat, lng := parseLoc(r.Loc)
	info := Info{
		IP:          r.IP,
		Hostname:    r.Hostname,
		City:        r.City,
		Region:      r.Region,
		CountryCode: r.Country, // ipinfo 的 country 是代码
		Latitude:    lat,
		Longitude:   lng,
		Org:         r.Org,
		Postal:      r.Postal,
		Timezone:    r.Timezone,
		Source:      "ipinfo.io",
	}
	// org 形如 "AS15169 Google LLC"，提取 ASN
	if m := asnRe.FindStringSubmatch(r.Org); m != nil {
		info.ASN = m[1]
	}
	return info, nil
}

// --- ipwho.is ---

type ipwhoisResponse struct {
	IP          string  `json:"ip"`
	Success     bool    `json:"success"`
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Postal      string  `json:"postal"`
	Connection  struct {
		ASN float64 `json:"asn"`
		Org string  `json:"org"`
		ISP string  `json:"isp"`
	} `json:"connection"`
	Timezone struct {
		ID string `json:"id"`
	} `json:"timezone"`
}

func parseIpwhois(body []byte) (Info, error) {
	var r ipwhoisResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{}, err
	}
	if !r.Success {
		return Info{}, fmt.Errorf("ipwho.is returned success=false")
	}
	return Info{
		IP:          r.IP,
		Continent:   r.Continent,
		Country:     r.Country,
		CountryCode: r.CountryCode,
		Region:      r.Region,
		City:        r.City,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		Postal:      r.Postal,
		Org:         r.Connection.Org,
		ASN:         formatASN(r.Connection.ASN),
		ISP:         r.Connection.ISP,
		Timezone:    r.Timezone.ID,
		Source:      "ipwho.is",
	}, nil
}

// --- ip9.com.cn ---

type ip9Response struct {
	Data struct {
		IP          string `json:"ip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Prov        string `json:"prov"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
		Lng         string `json:"lng"`
		Lat         string `json:"lat"`
		PostCode    string `json:"post_code"`
	} `json:"data"`
}

func parseIp9(body []byte) (Info, error) {
	var r ip9Response
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{}, err
	}
	return Info{
		IP:          r.Data.IP,
		Country:     r.Data.Country,
		CountryCode: strings.ToUpper(r.Data.CountryCode),
		Region:      r.Data.Prov,
		City:        r.Data.City,
		ISP:         r.Data.ISP,
		Longitude:   parseFloat(r.Data.Lng),
		Latitude:    parseFloat(r.Data.Lat),
		Postal:      r.Data.PostCode,
		Source:      "ip9.com.cn",
	}, nil
}

// --- ipdata.info ---

type ipdataResponse struct {
	IP          string  `json:"ip"`
	Success     bool    `json:"success"`
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Postal      string  `json:"postal"`
	ASN         float64 `json:"asn"`
	ASNOrg      string  `json:"asn_org"`
	ISP         string  `json:"isp"`
	Timezone    string  `json:"timezone"`
}

func parseIpdata(body []byte) (Info, error) {
	var r ipdataResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{}, err
	}
	if !r.Success {
		return Info{}, fmt.Errorf("ipdata.info returned success=false")
	}
	return Info{
		IP:          r.IP,
		Continent:   r.Continent,
		Country:     r.Country,
		CountryCode: r.CountryCode,
		Region:      r.Region,
		City:        r.City,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		Postal:      r.Postal,
		Org:         r.ASNOrg,
		ASN:         formatASN(r.ASN),
		ISP:         r.ISP,
		Timezone:    r.Timezone,
		Source:      "ipdata.info",
	}, nil
}

// sourceURLs 定义 failover 顺序，也是默认源顺序。
var sourceList = []source{
	{name: "ipinfo.io", url: func(ip string) string { return "https://ipinfo.io/" + ip + "/json" }, parse: parseIpinfo},
	{name: "ipwho.is", url: func(ip string) string { return "https://ipwho.is/" + ip }, parse: parseIpwhois},
	{name: "ip9.com.cn", url: func(ip string) string { return "https://ip9.com.cn/get?ip=" + ip }, parse: parseIp9},
	{name: "ipdata.info", url: func(ip string) string { return "https://ipdata.info/json/" + ip }, parse: parseIpdata},
}

var sourceByName = func() map[string]source {
	m := make(map[string]source, len(sourceList))
	for _, s := range sourceList {
		m[s.name] = s
	}
	return m
}()

// isIP 校验是否为合法 IP。
func isIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
