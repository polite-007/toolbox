package fofa

import (
	"fmt"
	"os"
	"testing"
)

func TestClient_Stats2_Protocol(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats2(&Stats2Request{
		Query:    `title="login"`,
		Fields:   "protocol",
		MaxCount: 5,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== protocol 去重值 (最多5个) ===\n")
	for i, v := range resp.Content {
		fmt.Printf("  %d. %s\n", i+1, v)
	}
}

func TestClient_Stats2_Title(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats2(&Stats2Request{
		Query:    `title="login"`,
		Fields:   "title",
		MaxCount: 5,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== title 去重值 (最多5个) ===\n")
	for i, v := range resp.Content {
		fmt.Printf("  %d. %s\n", i+1, v)
	}
}

func TestClient_Stats2_Server(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats2(&Stats2Request{
		Query:    `title="login"`,
		Fields:   "server",
		MaxCount: 5,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== server 去重值 (最多5个) ===\n")
	for i, v := range resp.Content {
		fmt.Printf("  %d. %s\n", i+1, v)
	}
}

func TestClient_Stats2_Country(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	resp, err := client.Stats2(&Stats2Request{
		Query:    `title="login"`,
		Fields:   "country",
		MaxCount: 3,
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== country 去重值 (最多3个) ===\n")
	for i, v := range resp.Content {
		fmt.Printf("  %d. %s\n", i+1, v)
	}
}

func TestClient_Stats2_MaxCountLimit(t *testing.T) {
	if os.Getenv("FOFA_EMAIL") == "" || os.Getenv("FOFA_KEY") == "" {
		t.Skip("跳过测试：请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	client := NewClient(os.Getenv("FOFA_EMAIL"), os.Getenv("FOFA_KEY"))

	// MaxCount 超过 10 应被限制为 10
	resp, err := client.Stats2(&Stats2Request{
		Query:    `title="login"`,
		Fields:   "protocol",
		MaxCount: 100, // 故意传超过 10 的值
	})
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}

	fmt.Printf("=== MaxCount 限制测试 (请求100，实际最多10) ===\n")
	fmt.Printf("  返回数量: %d\n", len(resp.Content))
	if len(resp.Content) > 10 {
		t.Errorf("MaxCount 限制未生效，返回了 %d 个结果", len(resp.Content))
	}
	for i, v := range resp.Content {
		fmt.Printf("  %d. %s\n", i+1, v)
	}
}
