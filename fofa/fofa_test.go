package fofa

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	email = os.Getenv("FOFA_EMAIL")
	key   = os.Getenv("FOFA_KEY")
)

func TestClient_Stats(t *testing.T) {
	client := NewClient(email, key)

	// 测试统计请求
	req := &StatsRequest{
		Query:  `region="TW" && status_code="200" && type="subdomain" && after="2025-12-19"`,
		Fields: "fid",
		Size:   1000,
	}

	resp, err := client.Stats(req)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	if resp.Error {
		t.Fatalf("API 返回错误: %s", resp.ErrMsg)
	}

	fmt.Printf("统计成功！\n")
	fmt.Println(resp.Aggs)

	// 打印统计结果
	results := resp.GetResults("fid")
	for _, result := range results {
		fmt.Printf("%s: %d\n", result.Name, result.Count)
	}
}

// TestClient_Search_AllFields 测试获取所有字段，验证字段解析是否正常
func TestClient_Search_AllFields(t *testing.T) {
	if email == "" || key == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(email, key)

	// 构建包含所有字段的 Fields 字符串
	allFields := "ip,port,protocol,country,country_name,region,city,longitude,latitude,asn,org,host,domain,os,server,icp,title,jarm,header,banner,cert,base_protocol,link,cert.issuer.org,cert.issuer.cn,cert.subject.org,cert.subject.cn,tls.ja3s,tls.version,cert.sn,cert.not_before,cert.not_after,cert.domain,header_hash,banner_hash,banner_fid,cname,lastupdatetime,product,product_category,product.version,icon_hash,cert.is_valid,cname_domain,body,cert.is_match,cert.is_equal,icon,fid,structinfo,org2"

	// 测试搜索请求
	req := &SearchRequest{
		Query:  `ip="8.8.8.8"`,
		Size:   5,
		Page:   1,
		Fields: allFields,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	if resp.Error {
		t.Fatalf("API 返回错误: %s", resp.ErrMsg)
	}

	fmt.Printf("\n=== 测试获取所有字段 ===\n")
	fmt.Printf("查询语句: %s\n", resp.Query)
	fmt.Printf("返回记录数: %d\n", resp.Size)
	fmt.Printf("当前页码: %d\n", resp.Page)
	fmt.Printf("结果数量: %d\n", len(resp.Results))
	fmt.Printf("字段数量: %d\n", len(resp.Fields))
	fmt.Printf("字段列表: %v\n\n", resp.Fields)

	// 使用 GetResults() 方法转换结果
	results := resp.GetResults()
	fmt.Printf("转换后的结果数量: %d\n\n", len(results))

	// 验证并打印所有字段的值
	for i, result := range results {
		if i >= 2 { // 只打印前2条结果
			break
		}

		fmt.Printf("--- 结果 %d ---\n", i+1)

		// 使用反射获取所有字段的值
		fields := []string{
			"ip", "port", "protocol", "country", "country_name", "region", "city",
			"longitude", "latitude", "asn", "org", "host", "domain", "os", "server",
			"icp", "title", "jarm", "header", "banner", "cert", "base_protocol",
			"link", "cert.issuer.org", "cert.issuer.cn", "cert.subject.org",
			"cert.subject.cn", "tls.ja3s", "tls.version", "cert.sn",
			"cert.not_before", "cert.not_after", "cert.domain", "header_hash",
			"banner_hash", "banner_fid", "cname", "lastupdatetime", "product",
			"product_category", "product.version", "icon_hash", "cert.is_valid",
			"cname_domain", "body", "cert.is_match", "cert.is_equal", "icon",
			"fid", "structinfo", "org2",
		}

		// 验证每个字段
		missingFields := []string{}
		for _, field := range fields {
			value := result.GetContentByFields(field)
			if len(value) > 0 && value[0] != "" {
				fmt.Printf("  %-25s: %s\n", field, value[0])
			} else {
				// 记录空字段，但不报错（因为某些字段可能确实为空）
				missingFields = append(missingFields, field)
			}
		}

		if len(missingFields) > 0 {
			fmt.Printf("  空字段: %v\n", missingFields)
		}

		// 验证结构体字段是否都能正确访问
		fmt.Printf("\n  直接访问结构体字段测试:\n")
		fmt.Printf("    IP: %s\n", result.IP)
		fmt.Printf("    Port: %s\n", result.Port)
		fmt.Printf("    Protocol: %s\n", result.Protocol)
		fmt.Printf("    Host: %s\n", result.Host)
		fmt.Printf("    Domain: %s\n", result.Domain)
		fmt.Printf("    Title: %s\n", result.Title)
		fmt.Printf("    CertIssuerOrg: %s\n", result.CertIssuerOrg)
		fmt.Printf("    CertSubjectOrg: %s\n", result.CertSubjectOrg)
		fmt.Printf("    Product: %s\n", result.Product)
		fmt.Printf("    Org2: %s\n", result.Org2)
		fmt.Printf("\n")
	}

	// 验证没有发生 panic 或错误
	if len(results) == 0 && len(resp.Results) > 0 {
		t.Error("GetResults() 返回空结果，但原始结果不为空，可能存在字段映射问题")
	}

	// 验证字段数量是否匹配
	if len(resp.Fields) != len(strings.Split(allFields, ",")) {
		t.Logf("警告: 请求的字段数量 (%d) 与返回的字段数量 (%d) 不匹配",
			len(strings.Split(allFields, ",")), len(resp.Fields))
	}

	// 验证每个结果的行数是否与字段数匹配
	for i, row := range resp.Results {
		if len(row) != len(resp.Fields) {
			t.Errorf("结果 %d 的列数 (%d) 与字段数 (%d) 不匹配", i, len(row), len(resp.Fields))
		}
	}

	// 测试 GetContentByFields 方法
	if len(results) > 0 {
		testResult := results[0]

		// 测试单个字段
		ipValue := testResult.GetContentByFields("ip")
		if len(ipValue) != 1 {
			t.Errorf("GetContentByFields 返回的字段数量不正确，期望 1，实际 %d", len(ipValue))
		}

		// 测试多个字段
		multiValues := testResult.GetContentByFields("ip,port,host")
		if len(multiValues) != 3 {
			t.Errorf("GetContentByFields 返回的字段数量不正确，期望 3，实际 %d", len(multiValues))
		}

		// 测试不存在的字段
		nonExistent := testResult.GetContentByFields("non_existent_field")
		if len(nonExistent) != 1 || nonExistent[0] != "" {
			t.Errorf("不存在的字段应该返回空字符串")
		}
	}

	fmt.Printf("=== 测试完成，所有字段解析正常 ===\n\n")
}

func TestClient_Stats2(t *testing.T) {
	client := NewClient(email, key)

	req := &Stats2Request{
		Query:  `region="TW" && status_code="200" && type="subdomain" && after="2025-12-19" && title!=""`,
		Fields: "title",
	}

	resp, err := client.Stats2(req)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("统计成功！\n")
	fmt.Println(resp.Content)
}

func TestClient_Search(t *testing.T) {
	client := NewClient(email, key)

	req := &SearchRequest{
		Query:  `body="©" && (fid="3FbqvmRzTRPLgIobcmSaHQ==") && status_code="200" && type="subdomain" && title!="" && title="冷凍控制器管理系統 - EGOi" && is_ipv6=false && after="2025-12-20"`,
		Size:   5,
		Page:   1,
		Fields: "ip,body",
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("搜索成功！\n")
	fmt.Printf("查询语句: %s\n", resp.Query)
	fmt.Printf("返回记录数: %d\n", resp.Size)
	fmt.Printf("当前页码: %d\n", resp.Page)
	fmt.Printf("结果数量: %d\n", len(resp.Results))

	for i, result := range resp.GetResults() {
		fmt.Printf("--- 结果 %d ---\n", i+1)
		fmt.Printf("  IP: %s\n", result.IP)
		fmt.Printf("  Body: %s\n", result.Body)
	}
}
