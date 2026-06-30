package fofa

import (
	"fmt"
	"os"
	"testing"
)

func TestClient_SearchNext(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.SearchNext(&NextRequest{
		Query:  `domain="baidu.com"`,
		Fields: "ip,port,host,protocol",
		Size:   10,
	})
	if err != nil {
		t.Fatalf("连续翻页查询失败: %v", err)
	}

	fmt.Printf("=== 连续翻页首页 ===\n")
	fmt.Printf("查询: %s\n", resp.Query)
	fmt.Printf("匹配总数: %d\n", resp.Size)
	fmt.Printf("本页条数: %d\n", len(resp.Results))
	fmt.Printf("next: %s\n", resp.Next)
	fmt.Printf("HasMore: %v\n", resp.HasMore())

	for i, r := range resp.GetResults() {
		fmt.Printf("  %d. %s:%s %s\n", i+1, r.IP, r.Port, r.Host)
	}

	if resp.HasMore() {
		resp2, err := client.SearchNext(&NextRequest{
			Query:  `domain="baidu.com"`,
			Fields: "ip,port,host,protocol",
			Size:   10,
			Next:   resp.Next,
		})
		if err != nil {
			t.Fatalf("连续翻页第二页失败: %v", err)
		}
		fmt.Printf("第二页条数: %d\n", len(resp2.Results))
	}
}

func TestClient_SearchNextAll(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"), WithProxy("http://127.0.0.1:8080"))

	result := make([]*SearchResult, 0)

	total := 0
	pages := 0
	err := client.SearchNextAll(&NextRequest{
		Query:  `domain="baidu.com" && after="2026-06-01"`,
		Fields: "ip,port",
		Size:   100,
	}, func(resp *NextResponse) error {
		pages++
		total += len(resp.Results)
		result = append(result, resp.GetResults()...)
		return nil
	})
	if err != nil {
		t.Fatalf("SearchNextAll 失败: %v", err)
	}

	fmt.Printf("共 %d 页, %d 条\n", pages, total)
	if pages == 0 || len(result) == 0 {
		t.Fatal("未获取到任何数据")
	}
}
