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
| [`httpx`](httpx/) | 基于 `projectdiscovery/httpx` 的 HTTP 探测，以及可取消的流式 `Fetcher` |
| [`opendir`](opendir/) | 开放目录采集，支持深度、并发、限额与跳过规则 |
| [`naabu`](naabu/) | 基于 `projectdiscovery/naabu` 的端口扫描封装 |
| [`xlsx`](xlsx/) | Excel 读写，支持多 Sheet 与 JSON 导出 |
| [`dig`](dig/) | DNS 查询，支持多记录类型与自定义 DNS 服务器 |
| [`ipinfo`](ipinfo/) | 多免费源 IP 信息查询，自动 failover |

## 使用示例

### HTTP 探测

`Run` 发起请求，并通过回调返回每条结果。

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

`Result.ResponseBody` 只保留 4 KiB 预览。需要完整响应体时使用 `Fetcher`。

### 流式请求

```go
fetcher := httpx.NewFetcher(
	httpx.WithFetchTimeout(15*time.Second),      // 首字节超时（TTFB）
	httpx.WithFetchStallTimeout(30*time.Second), // body 连续无字节的最大间隔
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

响应体流式读取，可通过 `ctx` 取消；只要持续有字节到达，慢速下载不会被中断。

### 开放目录采集

```go
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

stats, err := client.Run(ctx, "http://example.com/files/")
if err != nil {
	return err
}
fmt.Printf("files=%d bytes=%d skipped=%d errors=%d\n",
	stats.Files, stats.Bytes, len(stats.Skipped), len(stats.Errors))
```

`depth` 按目录列表页面计数：`1` 表示只采集根目录。`Run` 只返回致命错误，单个 URL 的失败记录在 `Stats.Errors` 中。

### 端口扫描

```go
client := naabu.NewNaabuClient(naabu.WithRate(100))

err := client.Run(ctx, []string{"example.com"}, func(r naabu.Result) {
	fmt.Printf("%s:%d/%s open\n", r.Host, r.Port, r.Protocol)
})
```

### IP 信息查询

```go
client, err := ipinfo.NewIpinfoClient(
	ipinfo.WithProxy("socks5://127.0.0.1:1080"), // 可选
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
