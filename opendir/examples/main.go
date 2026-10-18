package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/polite-007/toolbox/opendir"
)

// 用法：
//
//	go run ./opendir/examples [target]
//
// 示例默认只采集文件元数据（URL、大小、类型），不读取响应体，
// 因此不会下载文件内容。需要下载时，在回调中读取 r 并写入文件即可。
func main() {
	target := "http://example.com/files/"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	client := opendir.NewOpenDirClient(
		opendir.WithDepth(3),
		opendir.WithThreads(50),
		opendir.WithMaxURLs(1000),
		opendir.WithMaxFileSize(50*1024*1024),
		opendir.WithFollowRedirects(true),
		// opendir.WithProxy("socks5://127.0.0.1:7895"), // 按需启用
		opendir.WithSkipSuffixes([]string{".jpg", ".pdf"}),
		opendir.WithOnFile(func(_ context.Context, f opendir.File, _ io.Reader) error {
			// 不读取 r，只输出元数据，不会下载文件内容
			fmt.Printf("%10d  %-30s  %s\n", f.Size, f.ContentType, f.URL)
			return nil
		}),
	)

	stats, err := client.Run(context.Background(), target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nfiles=%d dirs=%d bytes=%d skipped=%d errors=%d truncated=%v duration=%s\n",
		stats.Files, stats.DirsVisited, stats.Bytes,
		len(stats.Skipped), len(stats.Errors), stats.Truncated, stats.Duration)
}
