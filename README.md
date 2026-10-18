# Toolbox

Toolbox is a collection of Go packages that wrap common network-security and data-processing tools behind small, consistent APIs. Each package is independent, so you can import only what you need.

English | [中文](#中文说明)

## Installation

```bash
go get github.com/polite-007/toolbox
```

Requires Go 1.24 or later.

## Packages

| Package | Description |
|---|---|
| [`fofa`](fofa/) | FOFA API client for search, statistics, and asset lookup |
| [`httpx`](httpx/) | HTTP probing based on `projectdiscovery/httpx`, plus a cancellable streaming `Fetcher` |
| [`opendir`](opendir/) | Open directory (directory-listing) collector with depth, concurrency, and skip rules |
| [`naabu`](naabu/) | Port scanning wrapper for `projectdiscovery/naabu` |
| [`xlsx`](xlsx/) | Excel read/write with multiple sheets and JSON export |
| [`dig`](dig/) | DNS queries with multiple record types and custom resolvers |
| [`ipinfo`](ipinfo/) | IP network information from multiple free sources with automatic failover |

## Usage

### HTTP probing

`Run` sends requests and reports each result through a callback.

```go
client := httpx.NewHttpxClient(
	httpx.WithThreads(50),
	httpx.WithTechDetect(),
)

err := client.Run(ctx, []string{"example.com:443"}, func(r httpx.Result) {
	if r.Err != nil {
		return
	}
	fmt.Printf("%d %s %s\n", r.StatusCode, r.Title, r.URL)
})
```

`Result.ResponseBody` contains only a 4 KiB preview. Use `Fetcher` when you need the full body.

### Streaming requests

```go
fetcher := httpx.NewFetcher(
	httpx.WithFetchTimeout(15*time.Second),      // time to first byte
	httpx.WithFetchStallTimeout(30*time.Second), // max interval without body bytes
	httpx.WithFetchMaxBodySize(50*1024*1024),
)
defer fetcher.Close()

resp, err := fetcher.Get(ctx, "http://example.com/archive.zip")
if err != nil {
	return err
}
defer resp.Body.Close()

_, err = io.Copy(dst, resp.Body)
```

The body is streamed and can be canceled through `ctx`. Slow downloads are allowed as long as bytes keep arriving.

### Open directory collection

```go
client := opendir.NewOpenDirClient(
	opendir.WithDepth(2),
	opendir.WithThreads(50),
	opendir.WithMaxURLs(1000),
	opendir.WithMaxFileSize(50*1024*1024),
	opendir.WithSkipNames([]string{"private"}),          // case-sensitive substring
	opendir.WithSkipSuffixes([]string{".jpg", ".png"}),  // case-sensitive suffix
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

stats, err := client.Run(ctx, "http://example.com/files/")
if err != nil {
	return err
}
fmt.Printf("files=%d bytes=%d skipped=%d errors=%d\n",
	stats.Files, stats.Bytes, len(stats.Skipped), len(stats.Errors))
```

`depth` counts listing pages: `1` collects the root directory only. `Run` returns fatal errors only; per-URL failures are recorded in `Stats.Errors`.

### Port scanning

```go
client := naabu.NewNaabuClient(naabu.WithRate(100))

err := client.Run(ctx, []string{"example.com"}, func(r naabu.Result) {
	fmt.Printf("%s:%d/%s open\n", r.Host, r.Port, r.Protocol)
})
```

### IP information

```go
client, err := ipinfo.NewIpinfoClient(
	ipinfo.WithProxy("socks5://127.0.0.1:1080"), // optional
	ipinfo.WithTimeout(10*time.Second),
)
if err != nil {
	return err
}

info, err := client.Lookup(ctx, "8.8.8.8")
if err != nil {
	return err
}
fmt.Printf("%s %s %s\n", info.IP, info.Country, info.Org)
```

## License

[MIT](LICENSE)

---

## 中文说明

Toolbox 是一组 Go 工具库，把常用的网络安全与数据处理工具封装成小而一致的 API。各包相互独立，按需引入即可。

```bash
go get github.com/polite-007/toolbox
```

需要 Go 1.24 或更高版本。

| 包 | 说明 |
|---|---|
| [`fofa`](fofa/) | FOFA API 客户端，支持搜索、统计、资产查询 |
| [`httpx`](httpx/) | 基于 `projectdiscovery/httpx` 的 HTTP 探测，以及可取消的流式 `Fetcher` |
| [`opendir`](opendir/) | 开放目录采集，支持深度、并发、限额与跳过规则 |
| [`naabu`](naabu/) | 基于 `projectdiscovery/naabu` 的端口扫描封装 |
| [`xlsx`](xlsx/) | Excel 读写，支持多 Sheet 与 JSON 导出 |
| [`dig`](dig/) | DNS 查询，支持多记录类型与自定义 DNS 服务器 |
| [`ipinfo`](ipinfo/) | 多免费源 IP 信息查询，自动 failover |

使用示例见上方英文部分，代码可直接复用。
