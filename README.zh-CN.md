# Toolbox

Toolbox 是一组 Go 工具库，把常用的网络安全与数据处理工具封装成小而一致的 API。各包相互独立，按需引入即可。

[English](README.md) | 中文

## 安装

```bash
go get github.com/polite-007/toolbox
```

需要 Go 1.24 或更高版本。

## 包列表

| 包 | 说明 |
|---|---|
| [`fofa`](fofa/) | FOFA API 客户端，支持搜索、统计、资产查询 |
| [`httpx`](httpx/) | 基于 `projectdiscovery/httpx` 的 HTTP 探测 |
| [`opendir`](opendir/) | 开放目录采集，支持深度、并发、限额与跳过规则 |
| [`naabu`](naabu/) | 基于 `projectdiscovery/naabu` 的端口扫描封装 |
| [`xlsx`](xlsx/) | Excel 读写，支持多 Sheet 与 JSON 导出 |
| [`ipinfo`](ipinfo/) | 多免费源 IP 信息查询，自动 failover |

## 使用示例

### FOFA 搜索

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polite-007/toolbox/fofa"
)

func main() {
	client := fofa.NewClient("you@example.com", "YOUR_API_KEY")

	resp, err := client.Search(&fofa.SearchRequest{
		Ctx:    context.Background(),
		Query:  `title="Directory listing for /"`,
		Page:   1,
		Size:   10,
		Fields: "ip,port,title,link",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Fields)
	for _, row := range resp.Results {
		fmt.Println(row)
	}
}
```

### HTTP 探测

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
		httpx.WithThreads(50),
		httpx.WithTechDetect(),
	)

	err := client.Run(context.Background(), []string{"example.com:443"}, func(r httpx.Result) {
		if r.Err != nil {
			return
		}
		fmt.Printf("%d %s %s\n", r.StatusCode, r.Title, r.URL)
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

`Result.ResponseBody` 只保留 4 KiB 预览；需要流式读取完整文件内容时使用 `opendir`。

### 开放目录采集

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/polite-007/toolbox/opendir"
)

func main() {
	const dest = "./downloads"
	if err := os.MkdirAll(dest, 0o755); err != nil {
		log.Fatal(err)
	}

	client := opendir.NewOpenDirClient(
		opendir.WithDepth(2),
		opendir.WithThreads(50),
		opendir.WithMaxURLs(1000),
		opendir.WithMaxFileSize(50*1024*1024),
		opendir.WithSkipNames([]string{"private"}),         // 大小写敏感的子串
		opendir.WithSkipSuffixes([]string{".jpg", ".png"}), // 大小写敏感的后缀
		opendir.WithOnFile(func(ctx context.Context, f opendir.File, r io.Reader) error {
			// 回调并发执行，需保证 goroutine-safe；
			// r 仅在回调执行期间有效。
			out, err := os.Create(filepath.Join(dest, filepath.Base(f.Name)))
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, r)
			return err
		}),
	)

	stats, err := client.Run(context.Background(), "http://example.com/files/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("files=%d bytes=%d skipped=%d errors=%d\n",
		stats.Files, stats.Bytes, len(stats.Skipped), len(stats.Errors))
}
```

`depth` 按目录列表页面计数：`1` 表示只采集根目录。`Run` 只返回致命错误，单个 URL 的失败记录在 `Stats.Errors` 中。

可运行的示例见 [`opendir/examples`](opendir/examples/)，默认只采集元数据、不下载文件内容。

需要采集多个目标时使用 `RunTargets`，所有目标共享同一个 worker 池。返回结果按输入顺序排列，单个目标失败不会影响其他目标。使用 `opendir.WithOnTargetDone` 可以在每个目标完成时立即拿到结果（回调并发执行）：

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/polite-007/toolbox/opendir"
)

func main() {
	client := opendir.NewOpenDirClient(
		opendir.WithDepth(2),
		opendir.WithThreads(50),
		opendir.WithOnFile(func(ctx context.Context, f opendir.File, r io.Reader) error {
			return nil // 只取元数据
		}),
	)

	results, err := client.RunTargets(context.Background(), []string{
		"http://host-a/files/",
		"http://host-b/files/",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		if result.Err != nil {
			log.Printf("%s: %v", result.Target, result.Err)
			continue
		}
		fmt.Printf("%s: files=%d\n", result.Target, result.Stats.Files)
	}
}
```

### 端口扫描

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polite-007/toolbox/naabu"
)

func main() {
	client := naabu.NewNaabuClient(naabu.WithRate(100))

	err := client.Run(context.Background(), []string{"example.com"}, func(r naabu.Result) {
		fmt.Printf("%s:%d/%s open\n", r.Host, r.Port, r.Protocol)
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

### IP 信息查询

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/polite-007/toolbox/ipinfo"
)

func main() {
	client, err := ipinfo.NewIpinfoClient(
		ipinfo.WithProxy("socks5://127.0.0.1:1080"), // 可选
		ipinfo.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s %s\n", info.IP, info.Country, info.Org)
}
```

## License

[MIT](LICENSE)
