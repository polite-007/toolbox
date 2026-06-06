package dig_test

import (
	"fmt"
	"time"

	"github.com/polite-007/toolbox/dig"
)

// 示例：使用指定 DNS 服务器查询任意记录类型
func ExampleNewDig() {
	// 创建 Dig 实例，指定 DNS 服务器和超时时间
	d := dig.NewDig("8.8.8.8:53", 5*time.Second)

	// 查询 A 记录
	result, err := d.Query("example.com", "A")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	// 输出所有解析到的 IP
	for _, r := range result.Records {
		fmt.Println(r)
	}
}

// 示例：查询域名的 A 记录并格式化输出
func ExampleDig_QueryA() {
	d := dig.NewDig("8.8.8.8:53", 5*time.Second)

	result, err := d.QueryA("a.fofa.info")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	// FormatResult 返回人类可读的格式化结果
	fmt.Println(result.FormatResult())
}

// 示例：检测域名是否使用了 CDN
func ExampleDig_IsCDN() {
	d := dig.NewDig("8.8.8.8:53", 5*time.Second)

	// IsCDN 通过多 DNS 服务器解析对比来判断是否命中 CDN
	cdn, err := d.IsCDN("example.com")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("IsCDN:", cdn.IsCDN)
}
