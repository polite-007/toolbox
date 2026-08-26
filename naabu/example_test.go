package naabu_test

import (
	"fmt"

	"github.com/polite-007/toolbox/naabu"
)

// 示例：创建 naabu 端口扫描客户端，配置端口、速率和 CONNECT 扫描。
func ExampleNewNaabuClient() {
	_ = naabu.NewNaabuClient(
		naabu.WithPorts("80,443,8080"),
		naabu.WithRate(100),
	)
	fmt.Println("naabu client created")
	// Output:
	// naabu client created
}

// 示例：组合使用被动枚举、服务识别和自定义 DNS resolver。
func ExampleNewNaabuClient_withOptions() {
	_ = naabu.NewNaabuClient(
		naabu.WithPassive(),
		naabu.WithServiceDiscovery(),
		naabu.WithResolvers("8.8.8.8,1.1.1.1"),
		naabu.WithSystemResolver(),
	)
	fmt.Println("naabu client created")
	// Output:
	// naabu client created
}

// 示例：创建支持 host:port 输入的 naabu 端口扫描客户端。
func ExampleNaabuClient_RunHostPorts() {
	_ = naabu.NewNaabuClient(
		naabu.WithRate(100),
	)
	fmt.Println("naabu client created")
	// Output:
	// naabu client created
}