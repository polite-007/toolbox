# Toolbox

Toolbox is a collection of Go packages that wrap common network-security and data-processing tools behind small, consistent APIs. Each package is independent, so you can import only what you need.

English | [中文](README.zh-CN.md)

## Installation

```bash
go get github.com/polite-007/toolbox
```

Requires Go 1.24 or later.

## Packages

| Package | Description |
|---|---|
| [`fofa`](fofa/) | FOFA API client for search, statistics, and asset lookup |
| [`httpx`](httpx/) | HTTP probing based on `projectdiscovery/httpx` |
| [`opendir`](opendir/) | Open directory (directory-listing) collector with depth, concurrency, and skip rules |
| [`naabu`](naabu/) | Port scanning wrapper for `projectdiscovery/naabu` |
| [`xlsx`](xlsx/) | Excel read/write with multiple sheets and JSON export |
| [`ipinfo`](ipinfo/) | IP network information from multiple free sources with automatic failover |

## Usage

### FOFA search

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

### HTTP probing

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

`Result.ResponseBody` contains only a 4 KiB preview. Use `opendir` when you need to stream complete file contents.

### Open directory collection

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
		opendir.WithSkipNames([]string{"private"}),         // case-sensitive substring
		opendir.WithSkipSuffixes([]string{".jpg", ".png"}), // case-sensitive suffix
		opendir.WithOnFile(func(ctx context.Context, f opendir.File, r io.Reader) error {
			// The callback runs concurrently; keep it goroutine-safe.
			// r is valid only for the duration of the call.
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

`depth` counts listing pages: `1` collects the root directory only. `Run` returns fatal errors only; per-URL failures are recorded in `Stats.Errors`.

A runnable metadata-only example is available in [`opendir/examples`](opendir/examples/).

### Port scanning

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

### IP information

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
		ipinfo.WithProxy("socks5://127.0.0.1:1080"), // optional
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
