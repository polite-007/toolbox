package fofa

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestClient_Search_Free 测试免费版字段查询
func TestClient_Search_Free(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	req := &SearchRequest{
		Query:  `title="login" && status_code="200"`,
		Size:   3,
		Page:   1,
		Fields: SearchFieldsFree,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("=== 免费版字段查询 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Printf("返回字段数: %d\n", len(resp.Fields))

	for i, r := range resp.GetResults() {
		fmt.Printf("--- 结果 %d ---\n", i+1)
		fmt.Printf("  IP: %s, Port: %s, Host: %s\n", r.IP, r.Port, r.Host)
		fmt.Printf("  Title: %s, Protocol: %s\n", r.Title, r.Protocol)
	}
}

// TestClient_Search_Personal 测试个人版字段查询
func TestClient_Search_Personal(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	req := &SearchRequest{
		Query:  `title="login" && status_code="200"`,
		Size:   3,
		Page:   1,
		Fields: SearchFieldsPersonal,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("=== 个人版字段查询 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Printf("返回字段数: %d\n", len(resp.Fields))

	jsonResult, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	fmt.Printf("JSON 结果: \n%s\n", string(jsonResult))
}

// TestClient_Search_Professional 测试专业版字段查询
func TestClient_Search_Professional(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	req := &SearchRequest{
		Query:  `title="login" && status_code="200"`,
		Size:   3,
		Page:   1,
		Fields: SearchFieldsProfessional,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("=== 专业版字段查询 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Printf("返回字段数: %d\n", len(resp.Fields))

	for i, r := range resp.GetResults() {
		fmt.Printf("--- 结果 %d ---\n", i+1)
		fmt.Printf("  IP: %s, Port: %s, Host: %s\n", r.IP, r.Port, r.Host)
		fmt.Printf("  Product: %s, Cname: %s\n", r.Product, r.Cname)
		fmt.Printf("  LastUpdateTime: %s\n", r.Lastupdatetime)
	}
}

// TestClient_Search_Business 测试商业版字段查询
func TestClient_Search_Business(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	req := &SearchRequest{
		Query:  `title="login" && status_code="200"`,
		Size:   3,
		Page:   1,
		Fields: SearchFieldsBusiness,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("=== 商业版字段查询 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Printf("返回字段数: %d\n", len(resp.Fields))

	for i, r := range resp.GetResults() {
		fmt.Printf("--- 结果 %d ---\n", i+1)
		fmt.Printf("  IP: %s, Port: %s, Host: %s\n", r.IP, r.Port, r.Host)
		fmt.Printf("  Product: %s, ProductVersion: %s\n", r.Product, r.ProductVersion)
		fmt.Printf("  IconHash: %s, CertIsValid: %s\n", r.IconHash, r.CertIsValid)
	}
}

// TestClient_Search_Enterprise 测试企业版字段查询
func TestClient_Search_Enterprise(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	req := &SearchRequest{
		Query:  `title="login" && status_code="200"`,
		Size:   3,
		Page:   1,
		Fields: SearchFieldsEnterprise,
	}

	resp, err := client.Search(req)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("=== 企业版字段查询 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("总记录数: %d\n", resp.Size)
	fmt.Printf("返回字段数: %d\n", len(resp.Fields))

	for i, r := range resp.GetResults() {
		fmt.Printf("--- 结果 %d ---\n", i+1)
		fmt.Printf("  IP: %s, Port: %s, Host: %s\n", r.IP, r.Port, r.Host)
		fmt.Printf("  Product: %s, ProductVersion: %s\n", r.Product, r.ProductVersion)
		fmt.Printf("  Icon: %s, FID: %s\n", r.Icon, r.FID)
		fmt.Printf("  Structinfo: %s\n", r.Structinfo)
	}
}
