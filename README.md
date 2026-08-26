# Toolbox

Go 工具库集合，提供常用安全与数据处理能力，可直接 `go get` 导入使用。

旨在将很多复杂，好用的网络安全工具，pkg化

## 安装

```bash
go get github.com/polite-007/toolbox
```

## 包列表

| 包 | 导入路径 | 说明 |
|---|---|---|
| **fofa** | `github.com/polite-007/toolbox/fofa` | FOFA API 客户端，支持搜索、统计、资产查询 |
| **httpx** | `github.com/polite-007/toolbox/httpx` | HTTP 探测封装（基于 projectdiscovery/httpx） |
| **naabu** | `github.com/polite-007/toolbox/naabu` | 端口扫描封装（基于 projectdiscovery/naabu） |
| **xlsx** | `github.com/polite-007/toolbox/xlsx` | Excel 读写工具，支持多 Sheet、JSON 导出 |
| **dig** | `github.com/polite-007/toolbox/dig` | DNS 查询工具，支持多记录类型、自定义 DNS 服务器 |

## 使用示例

### naabu：端口扫描

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polite-007/toolbox/naabu"
)

func main() {
	client := naabu.NewNaabuClient(
		naabu.WithRate(100),
	)

	// 方式一：统一扫描一批目标的一组端口
	err := client.Run(context.Background(), []string{"example.com"}, func(r naabu.Result) {
		fmt.Printf("%s:%d/%s open\n", r.Host, r.Port, r.Protocol)
	})

	// 方式二：输入 ip:port 列表，按 host 聚合精确扫描，结果保持输入顺序
	err = client.RunHostPorts(context.Background(), []string{
		"1.1.1.1:80",
		"1.1.1.1:443",
		"u:8.8.8.8:53",
	}, func(r naabu.Result) {
		fmt.Printf("%s:%d/%s open\n", r.Host, r.Port, r.Protocol)
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

### httpx：HTTP 探测

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polite-007/toolbox/httpx"
)

func main() {
	client := httpx.NewHttpxClient(
		httpx.WithTechDetect(),
		httpx.WithFollowRedirects(),
	)

	err := client.Run(context.Background(), []string{"example.com:443"}, func(r httpx.Result) {
		if r.Err != nil {
			fmt.Println(r.Err)
			return
		}
		fmt.Printf("%d %s %s %v\n", r.StatusCode, r.Title, r.URL, r.Technologies)
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## License

[LICENSE](./LICENSE)
