package opendir

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	// ErrInvalidTarget 表示 target 不是合法的 HTTP/HTTPS URL
	ErrInvalidTarget = errors.New("opendir: invalid target")
	// ErrNoHandler 表示未配置 OnFile 回调
	ErrNoHandler = errors.New("opendir: OnFile handler is required")
	// ErrNotListing 表示目标不是目录列表页面
	ErrNotListing = errors.New("opendir: target does not serve a directory listing")
)

// File 描述一个被采集到的文件
type File struct {
	URL         string // 完整、编码后的绝对 URL
	Dir         string // 文件所在目录的绝对 URL，以 / 结尾
	Name        string // 解码后的文件名
	Size        int64  // 响应 Content-Length，未知为 -1
	ContentType string // 响应 Content-Type
}

// FileHandler 处理单个文件的内容。
//
// r 仅在回调执行期间有效，且由框架限制在 MaxFileSize 以内；
// 回调在多个 worker 中并发执行，必须保证 goroutine-safe。
// 返回 error 时仅记录到 Stats.Errors，采集继续。
type FileHandler func(ctx context.Context, f File, r io.Reader) error

// SkipReason 表示文件被跳过的原因
type SkipReason string

const (
	// SkipName 命中 SkipNames（大小写敏感的子串）
	SkipName SkipReason = "skip_name"
	// SkipSuffix 命中 SkipSuffixes（大小写敏感的后缀）
	SkipSuffix SkipReason = "skip_suffix"
	// SkipStatus 文件请求返回非 2xx
	SkipStatus SkipReason = "status"
	// SkipTooLarge 文件超过 MaxFileSize
	SkipTooLarge SkipReason = "too_large"
)

// SkippedFile 描述一个被跳过的文件
type SkippedFile struct {
	URL    string
	Name   string
	Reason SkipReason
}

// FileError 描述一个文件级错误
type FileError struct {
	URL string
	Err error
}

// Stats 是一次采集的统计结果
type Stats struct {
	Target         string        // 归一化后的目标
	Requests       int           // 实际发出的请求数（含根页面）
	Files          int           // 成功回调的文件数
	Bytes          int64         // 读取的文件字节数
	DirsVisited    int           // 访问过的目录页数
	DirsDiscovered int           // 发现并调度的子目录数
	Skipped        []SkippedFile // 被跳过的文件
	Errors         []FileError   // 单 URL 错误（最多保留 MaxErrors 条）
	ErrorsOmitted  int           // 超出保留上限的错误数
	Duration       time.Duration
	Truncated      bool // 是否触及 MaxURLs 上限
}
