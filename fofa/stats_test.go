package fofa

import (
	"fmt"
	"os"
	"testing"
)

func TestClient_Stats_Protocol(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "protocol",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== protocol 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Println(resp)
	for _, r := range resp.GetResults(resp.Field) {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Domain(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "domain",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== domain 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("domain") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Port(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "port",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== port 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("port") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Title(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "title",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== title 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("title") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_OS(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "os",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== os 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("os") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Server(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "server",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== server 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("server") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Country(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "country",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== country 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("country") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
		for _, reg := range r.Regions {
			fmt.Printf("    - %s: %d\n", reg.Name, reg.Count)
		}
	}
}

// TestGetResults_CountryKeyMapping 不依赖网络，验证 country 字段
// 在 aggs 中实际使用 countries 作为 key 的映射逻辑
func TestGetResults_CountryKeyMapping(t *testing.T) {
	resp := &StatsResponse{
		Aggs: map[string]interface{}{
			"countries": []interface{}{
				map[string]interface{}{
					"name":  "美国",
					"count": float64(3182),
					"regions": []interface{}{
						map[string]interface{}{"name": "加利福尼亚", "count": float64(500)},
						map[string]interface{}{"name": "纽约", "count": float64(300)},
					},
				},
				map[string]interface{}{
					"name":  "德国",
					"count": float64(259),
				},
			},
		},
	}

	results := resp.GetResults("country")
	if len(results) != 2 {
		t.Fatalf("期望返回 2 条结果，实际 %d 条", len(results))
	}
	if results[0].Name != "美国" || results[0].Count != 3182 {
		t.Fatalf("第一条结果不匹配: %+v", results[0])
	}
	if len(results[0].Regions) != 2 {
		t.Fatalf("期望 2 个区域，实际 %d 个", len(results[0].Regions))
	}
	if results[0].Regions[0].Name != "加利福尼亚" || results[0].Regions[0].Count != 500 {
		t.Fatalf("区域信息不匹配: %+v", results[0].Regions[0])
	}
	if results[1].Name != "德国" || results[1].Count != 259 {
		t.Fatalf("第二条结果不匹配: %+v", results[1])
	}

	// 不存在的字段应返回 nil
	if got := resp.GetResults("protocol"); got != nil {
		t.Fatalf("期望 protocol 字段返回 nil，实际 %+v", got)
	}
}

func TestClient_Stats_ASN(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "asn",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== asn 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("asn") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Org(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "org",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== org 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("org") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_AssetType(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "asset_type",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== asset_type 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("asset_type") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_Fid(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "fid",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== fid 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("fid") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_ICP(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `title="百度"`,
		Fields: "icp",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== icp 统计 ===\n")
	fmt.Printf("总记录数: %d\n", resp.Size)
	for _, r := range resp.GetResults("icp") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}

func TestClient_Stats_MultiFields(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats(&StatsRequest{
		Query:  `domain=baidu.com`,
		Fields: "protocol,port,country,server",
		Size:   100,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== 多字段统计 ===\n")
	fmt.Printf("统计字段: %v\n", resp.Fields())
	fmt.Printf("总记录数: %d\n", resp.Size)

	// 一次性获取所有字段的统计结果
	allResults := resp.GetAllResults()
	for field, results := range allResults {
		fmt.Printf("\n--- %s 统计 ---\n", field)
		fmt.Printf("distinct: %d\n", resp.GetDistinct(field))
		for _, r := range results {
			fmt.Printf("  %s: %d\n", r.Name, r.Count)
		}
	}

	// 也可单独获取某个字段
	fmt.Printf("\n--- 单独获取 port ---\n")
	for _, r := range resp.GetResults("port") {
		fmt.Printf("  %s: %d\n", r.Name, r.Count)
	}
}
