package fofa

import (
	"fmt"
	"os"
	"testing"
)

func TestHotSearch(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	hotSearchResponse, err := client.HotSearch()
	if err != nil {
		t.Fatalf("获取热门搜索失败: %v", err)
	}

	fmt.Printf("=== 热门搜索 ===\n")
	for _, data := range hotSearchResponse.Data {
		fmt.Printf("  %s: %d\n", data.App, data.Asset_count)
	}
}
