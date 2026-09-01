package ipinfo_test

import (
	"fmt"
	"time"

	"github.com/polite-007/toolbox/ipinfo"
)

// 示例：创建 IP 信息查询客户端，配置超时和代理。
func ExampleNewIpinfoClient() {
	client, err := ipinfo.NewIpinfoClient(
		ipinfo.WithTimeout(10*time.Second),
		ipinfo.WithProxy("socks5://127.0.0.1:1080"),
	)
	if err != nil {
		fmt.Println("create failed:", err)
		return
	}
	fmt.Println("client created:", client != nil)
	// Output:
	// client created: true
}
