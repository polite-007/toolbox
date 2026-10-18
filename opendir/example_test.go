package opendir_test

import (
	"context"
	"fmt"
	"io"

	"github.com/polite-007/toolbox/opendir"
)

// 示例：创建一个开放目录采集客户端
func ExampleNewOpenDirClient() {
	client := opendir.NewOpenDirClient(
		opendir.WithDepth(2),
		opendir.WithThreads(50),
		opendir.WithMaxFileSize(50*1024*1024),
		opendir.WithSkipSuffixes([]string{".jpg", ".png"}),
		opendir.WithOnFile(func(ctx context.Context, f opendir.File, r io.Reader) error {
			fmt.Printf("file: %s, size: %d, type: %s\n", f.URL, f.Size, f.ContentType)

			// r 只在回调执行期间有效，需自行保证并发安全
			// _, err := io.Copy(io.Discard, r)
			return nil
		}),
	)
	_ = client

	// 实际使用时：
	// stats, err := client.Run(context.Background(), "http://example.com/files/")
	fmt.Println("opendir client created")
	// Output:
	// opendir client created
}
