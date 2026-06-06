package httpx_test

import (
	"fmt"

	"github.com/polite-007/toolbox/httpx"
)

// 示例：创建 HTTP 探测客户端，配置线程数、技术检测、跟随重定向
func ExampleNewHttpxClient() {
	c := httpx.NewHttpxClient(
		httpx.WithThreads(50),          // 并发线程数
		httpx.WithTechDetect(),         // 启用 Web 指纹识别
		httpx.WithFollowRedirects(),    // 跟随 HTTP 重定向
	)
	fmt.Println("httpx client created, threads:", c.Options.Threads)
	// Output:
	// httpx client created, threads: 50
}

// 示例：使用代理和自定义端口创建客户端
func ExampleNewHttpxClient_withProxy() {
	c := httpx.NewHttpxClient(
		httpx.WithPorxy("http://127.0.0.1:7890"),  // 设置 HTTP 代理
		httpx.WithPorts([]string{"80", "443", "8080"}), // 指定探测端口
		httpx.WithFaviconMMH3(),                     // 获取 favicon hash
	)
	fmt.Println("httpx client created, proxy:", c.Options.Proxy)
	// Output:
	// httpx client created, proxy: http://127.0.0.1:7890
}
