package fofa_test

import (
	"context"
	"fmt"
	"time"

	"github.com/polite-007/toolbox/fofa"
)

// 示例：创建 FOFA 客户端并设置超时时间
func ExampleNewClient() {
	// 使用邮箱和 API Key 创建客户端
	client := fofa.NewClient("your@email.com", "your_api_key")
	// 设置请求超时，默认 30 秒
	client.SetTimeout(10 * time.Second)
	fmt.Println("FOFA client created")
	// Output:
	// FOFA client created
}

// 示例：执行 FOFA 搜索查询，返回匹配的资产列表
func ExampleClient_Search() {
	client := fofa.NewClient("your@email.com", "your_api_key")

	// 搜索 title 包含 "login" 的资产，返回 ip、port、host、title 字段
	resp, err := client.Search(&fofa.SearchRequest{
		Query:  `title="login"`,
		Size:   10,                                // 最多返回 10 条
		Fields: "ip,port,host,title",              // 指定返回字段
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// 遍历结果，将每条记录转为结构体
	for _, r := range resp.GetResults() {
		fmt.Println(r.Host, r.IP, r.Port)
	}
	_ = resp
}

// 示例：查询 FOFA 中匹配条件的资产总数
func ExampleClient_Count() {
	client := fofa.NewClient("your@email.com", "your_api_key")

	// 仅获取数量，不返回具体数据
	count, err := client.Count(context.Background(), `title="login"`)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("total:", count)
}
