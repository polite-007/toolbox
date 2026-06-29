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
