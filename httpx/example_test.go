package httpx_test

import (
	"fmt"
	"testing"

	"github.com/polite-007/toolbox/httpx"
	"github.com/projectdiscovery/httpx/runner"
)

// 示例：创建 HTTP 探测客户端，配置线程数、技术检测、跟随重定向
func ExampleNewHttpxClient() {
	_ = httpx.NewHttpxClient(
		httpx.WithThreads(50),
		httpx.WithTechDetect(),
		httpx.WithFollowRedirects(),
	)
	fmt.Println("httpx client created")
	// Output:
	// httpx client created
}

// 示例：使用代理和自定义端口创建客户端
func ExampleNewHttpxClient_withProxy() {
	_ = httpx.NewHttpxClient(
		httpx.WithProxy("http://127.0.0.1:7890"),
		httpx.WithPorts([]string{"80", "443", "8080"}),
		httpx.WithFaviconMMH3(),
	)
	fmt.Println("httpx client created")
	// Output:
	// httpx client created
}

// 示例：组合使用 Headers 和 CustomHost（验证追加不覆盖）
func ExampleNewHttpxClient_headerMerge() {
	_ = httpx.NewHttpxClient(
		httpx.WithHeaders([]string{"X-Token: abc123"}),
		httpx.WithCustomHost("internal.example.com"),
	)
	fmt.Println("httpx client created")
	// Output:
	// httpx client created
}

func TestRun(t *testing.T) {
	client := httpx.NewHttpxClient(
		httpx.WithProxy("http://127.0.0.1:8080"),
	)
	err := client.Run([]string{"a.fofa.info", "www.baidu.com", "www.fofa.info"}, func(r runner.Result) {
		if r.Err != nil {
			fmt.Println(r.Err)
			return
		}
		fmt.Println(r.Title, r.StatusCode)
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}