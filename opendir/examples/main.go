package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/polite-007/toolbox/opendir"
)

// 用法：
//
//	go run ./opendir/examples [target]
//
// 示例把采集到的文件下载到用户目录下的 ~/open-dir/<host>/，
// 并按远程目录结构存放。只需要元数据时，把 WithOnFile 换成注释里的版本，
// 不读取 r 即可。
func main() {
	target := "http://example.com/files/"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	u, err := url.Parse(target)
	if err != nil || u.Host == "" {
		log.Fatalf("invalid target %q", target)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	// 用目标主机名区分不同目标的输出目录；":" 不能出现在文件名中，替换掉
	dest := filepath.Join(home, "open-dir", strings.ReplaceAll(u.Host, ":", "_"))
	if err := os.MkdirAll(dest, 0o755); err != nil {
		log.Fatal(err)
	}
	log.Printf("start: target=%s dest=%s", target, dest)

	client := opendir.NewOpenDirClient(
		opendir.WithDepth(3),
		opendir.WithThreads(50),
		opendir.WithMaxURLs(1000),
		opendir.WithMaxFileSize(50*1024*1024), // 50MB
		opendir.WithFollowRedirects(true),
		// opendir.WithProxy("socks5://127.0.0.1:7895"), // 按需启用
		// opendir.WithSkipSuffixes([]string{".jpg", ".pdf"}),
		// opendir.WithSkipNames([]string{"简历", "secret"}),
		opendir.WithSkipDirNames([]string{"pylib",".pnpm",".npm","go",".git", "node_modules", "vendor", ".venv"}),

		// 只采集元数据、不下载内容时使用：
		// opendir.WithOnFile(func(_ context.Context, f opendir.File, _ io.Reader) error {
		// 	fmt.Printf("%10d  %-30s  %s\n", f.Size, f.ContentType, f.URL)
		// 	return nil
		// }),

		opendir.WithOnFile(func(_ context.Context, f opendir.File, r io.Reader) error {
			dirURL, err := url.Parse(f.Dir)
			if err != nil {
				return err
			}
			// 用远程目录路径镜像本地目录结构：dest/<远程目录>/<文件名>
			// path.Clean 会收敛 "." 和 ".."，再去掉前导 "/"，不会逃出 dest
			rel := strings.TrimPrefix(path.Clean(dirURL.Path), "/")
			dir := filepath.Join(dest, filepath.FromSlash(rel))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}

			name := filepath.Base(f.Name)
			if name == "" || name == "." || name == ".." {
				return fmt.Errorf("invalid file name %q", f.Name)
			}
			local := filepath.Join(dir, name)
			out, err := os.Create(local)
			if err != nil {
				return err
			}
			defer out.Close()

			n, err := io.Copy(out, r) // 边下载边写盘
			if err != nil {
				return err
			}
			log.Printf("saved %s (%d bytes) -> %s", f.URL, n, local)
			return nil
		}),
	)

	stats, err := client.Run(context.Background(), target)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("done: files=%d dirs=%d bytes=%d skipped=%d errors=%d truncated=%v duration=%s",
		stats.Files, stats.DirsVisited, stats.Bytes,
		len(stats.Skipped), len(stats.Errors), stats.Truncated, stats.Duration)
}
