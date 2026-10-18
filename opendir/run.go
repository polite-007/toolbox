package opendir

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// drainLimit 是回调提前返回时最多排空的字节数。
// 排空少量字节有助于连接复用，同时避免为未读的大文件浪费带宽。
const drainLimit = 32 * 1024

// maxListingBytes 是单个目录列表页面的读取上限，防止异常响应耗尽内存
const maxListingBytes = 10 * 1024 * 1024

type taskKind int

const (
	taskDir taskKind = iota
	taskFile
)

type crawlTask struct {
	kind  taskKind
	url   *url.URL
	name  string
	level int
}

// crawler 持有单次采集的运行时状态
type crawler struct {
	config  *Config
	fetcher *fetcher
	ctx     context.Context

	tasks   chan crawlTask
	pending atomic.Int64
	doneCh  chan struct{}
	done    sync.Once
	workers sync.WaitGroup

	mu      sync.Mutex
	stats   *Stats
	visited map[string]struct{}
}

// Run 采集单个目标的开放目录。
//
// 返回的 error 仅表示目标级致命错误（ctx 取消、target 非法、根页面请求失败、
// 未配置回调）；单个 URL 的失败记录在 Stats.Errors 中。
func (c *OpenDirClient) Run(ctx context.Context, target string) (*Stats, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg := c.config
	if cfg.OnFile == nil {
		return nil, ErrNoHandler
	}
	if cfg.Depth < 1 {
		return nil, fmt.Errorf("opendir: depth must be >= 1")
	}
	if cfg.Threads < 1 {
		return nil, fmt.Errorf("opendir: threads must be >= 1")
	}
	if cfg.MaxURLs < 1 {
		return nil, fmt.Errorf("opendir: max_urls must be >= 1")
	}
	if cfg.MaxErrors < 0 {
		return nil, fmt.Errorf("opendir: max_errors must be >= 0")
	}

	root, err := normalizeTarget(target)
	if err != nil {
		return nil, err
	}

	if cfg.MaxDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.MaxDuration)
		defer cancel()
	}

	stats := &Stats{Target: root.String()}
	start := time.Now()
	defer func() {
		stats.Duration = time.Since(start)
	}()

	cr := &crawler{
		config:  cfg,
		fetcher: c.fetcher,
		ctx:     ctx,
		tasks:   make(chan crawlTask, cfg.Threads),
		doneCh:  make(chan struct{}),
		stats:   stats,
		visited: make(map[string]struct{}),
	}

	// 根页面由主 goroutine 同步抓取，用于校验目标并决定是否继续
	entries, err := cr.fetchRoot(root)
	if err != nil {
		return stats, err
	}

	for i := 0; i < cfg.Threads; i++ {
		cr.workers.Add(1)
		go cr.worker()
	}
	go func() {
		select {
		case <-cr.doneCh:
		case <-ctx.Done():
		}
		close(cr.tasks)
	}()

	// 调度期间保持 pending > 0，避免 worker 提前判定采集结束
	cr.pending.Add(1)
	cr.scheduleEntries(entries, 1)
	cr.pending.Add(-1)
	cr.finishIfIdle()

	select {
	case <-cr.doneCh:
	case <-ctx.Done():
	}
	cr.workers.Wait()

	if err := ctx.Err(); err != nil {
		return stats, err
	}
	return stats, nil
}

// fetchRoot 抓取并解析根目录页面
func (c *crawler) fetchRoot(root *url.URL) ([]listingEntry, error) {
	resp, err := c.fetcher.Get(c.ctx, root.String())
	if err != nil {
		return nil, fmt.Errorf("opendir: fetch root: %w", err)
	}
	defer resp.Body.Close()

	c.mu.Lock()
	c.stats.Requests = 1
	c.mu.Unlock()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opendir: root returned status %d", resp.StatusCode)
	}

	entries, err := parseListing(resp.URL, newLimitReader(resp.Body, maxListingBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("opendir: parse root: %w", err)
	}
	if len(entries) == 0 && !isHTML(resp.Header.Get("Content-Type")) {
		return nil, ErrNotListing
	}

	c.mu.Lock()
	c.stats.DirsVisited = 1
	c.visited[resp.URL.String()] = struct{}{}
	c.mu.Unlock()

	return entries, nil
}

// scheduleEntries 调度一个目录页面中的条目
func (c *crawler) scheduleEntries(entries []listingEntry, level int) {
	for _, entry := range entries {
		if entry.isDir {
			if level+1 > c.config.Depth {
				continue
			}
			task := crawlTask{kind: taskDir, url: entry.url, level: level + 1}
			if c.enqueue(task) {
				c.mu.Lock()
				c.stats.DirsDiscovered++
				c.mu.Unlock()
			}
			continue
		}

		if reason, skip := c.skipReason(entry.name); skip {
			c.recordSkipped(SkippedFile{
				URL:    entry.url.String(),
				Name:   entry.name,
				Reason: reason,
			})
			continue
		}
		c.enqueue(crawlTask{kind: taskFile, url: entry.url, name: entry.name, level: level})
	}
}

// skipReason 判断文件是否命中跳过规则（大小写敏感）
func (c *crawler) skipReason(name string) (SkipReason, bool) {
	for _, keyword := range c.config.SkipNames {
		if keyword != "" && strings.Contains(name, keyword) {
			return SkipName, true
		}
	}
	for _, suffix := range c.config.SkipSuffixes {
		if suffix == "" {
			continue
		}
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		if strings.HasSuffix(name, suffix) {
			return SkipSuffix, true
		}
	}
	return "", false
}

// enqueue 去重并调度任务；超过 MaxURLs 时停止入队
func (c *crawler) enqueue(task crawlTask) bool {
	if c.ctx.Err() != nil {
		return false
	}

	key := task.url.String()
	c.mu.Lock()
	if c.stats.Requests >= c.config.MaxURLs {
		c.stats.Truncated = true
		c.mu.Unlock()
		return false
	}
	if _, seen := c.visited[key]; seen {
		c.mu.Unlock()
		return false
	}
	c.visited[key] = struct{}{}
	c.stats.Requests++
	c.mu.Unlock()

	c.pending.Add(1)
	select {
	case c.tasks <- task:
		return true
	case <-c.ctx.Done():
		c.pending.Add(-1)
		c.finishIfIdle()
		return false
	}
}

func (c *crawler) worker() {
	defer c.workers.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case task, ok := <-c.tasks:
			if !ok {
				return
			}
			c.process(task)
			c.pending.Add(-1)
			c.finishIfIdle()
		}
	}
}

func (c *crawler) finishIfIdle() {
	if c.pending.Load() == 0 {
		c.done.Do(func() {
			close(c.doneCh)
		})
	}
}

func (c *crawler) process(task crawlTask) {
	switch task.kind {
	case taskDir:
		c.processDir(task)
	case taskFile:
		c.processFile(task)
	}
}

func (c *crawler) processDir(task crawlTask) {
	resp, err := c.fetcher.Get(c.ctx, task.url.String())
	if err != nil {
		c.recordRequestError(task.url.String(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordRequestError(task.url.String(), fmt.Errorf("unexpected status %d", resp.StatusCode))
		return
	}

	entries, err := parseListing(resp.URL, newLimitReader(resp.Body, maxListingBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		c.recordRequestError(task.url.String(), fmt.Errorf("parse listing: %w", err))
		return
	}

	c.mu.Lock()
	c.stats.DirsVisited++
	c.mu.Unlock()

	c.scheduleEntries(entries, task.level)
}

func (c *crawler) processFile(task crawlTask) {
	resp, err := c.fetcher.Get(c.ctx, task.url.String())
	if err != nil {
		c.recordRequestError(task.url.String(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordSkipped(SkippedFile{
			URL:    task.url.String(),
			Name:   task.name,
			Reason: SkipStatus,
		})
		return
	}
	if c.config.MaxFileSize > 0 && resp.ContentLength > c.config.MaxFileSize {
		c.recordSkipped(SkippedFile{
			URL:    task.url.String(),
			Name:   task.name,
			Reason: SkipTooLarge,
		})
		return
	}

	file := File{
		URL:         task.url.String(),
		Dir:         fileDir(task.url),
		Name:        task.name,
		Size:        resp.ContentLength,
		ContentType: resp.Header.Get("Content-Type"),
	}

	var body io.Reader = resp.Body
	if c.config.MaxFileSize > 0 {
		body = newLimitReader(body, c.config.MaxFileSize)
	}

	counter := &countingReader{r: body}
	handlerErr := c.config.OnFile(c.ctx, file, counter)

	// 回调读过内容时，排空少量剩余字节以复用连接；
	// 回调完全没读（只取元数据）时直接关闭，避免无谓下载。
	if counter.n > 0 {
		_, _ = io.CopyN(io.Discard, counter, drainLimit)
	}

	c.mu.Lock()
	c.stats.Bytes += counter.n
	c.mu.Unlock()

	if errors.Is(handlerErr, errBodyTooLarge) {
		c.recordSkipped(SkippedFile{
			URL:    task.url.String(),
			Name:   task.name,
			Reason: SkipTooLarge,
		})
		return
	}
	if handlerErr != nil {
		c.recordRequestError(task.url.String(), handlerErr)
		return
	}
	c.mu.Lock()
	c.stats.Files++
	c.mu.Unlock()
}

// recordRequestError 记录单 URL 错误；ctx 取消导致的错误不记录
func (c *crawler) recordRequestError(url string, err error) {
	if c.ctx.Err() != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.stats.Errors) < c.config.MaxErrors {
		c.stats.Errors = append(c.stats.Errors, FileError{URL: url, Err: err})
		return
	}
	c.stats.ErrorsOmitted++
}

func (c *crawler) recordSkipped(skipped SkippedFile) {
	c.mu.Lock()
	c.stats.Skipped = append(c.stats.Skipped, skipped)
	c.mu.Unlock()
}

// isHTML 判断 Content-Type 是否为 HTML
func isHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml")
}

// countingReader 统计实际读取的字节数
type countingReader struct {
	r io.Reader
	n int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += int64(n)
	return n, err
}

// limitReader 在读取超过上限时返回 errBodyTooLarge
type limitReader struct {
	r   io.Reader
	n   int64
	max int64
}

func newLimitReader(r io.Reader, max int64) *limitReader {
	return &limitReader{r: r, max: max}
}

func (l *limitReader) Read(p []byte) (int, error) {
	remaining := l.max + 1 - l.n
	if remaining <= 0 {
		return 0, errBodyTooLarge
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := l.r.Read(p)
	l.n += int64(n)
	if l.n > l.max {
		return 0, errBodyTooLarge
	}
	return n, err
}
