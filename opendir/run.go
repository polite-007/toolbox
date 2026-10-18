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
	taskRoot taskKind = iota
	taskDir
	taskFile
)

type crawlTask struct {
	kind  taskKind
	url   *url.URL
	name  string
	level int
	owner *targetState
}

// targetState 保存单个目标的运行时状态
type targetState struct {
	root     *url.URL
	stats    *Stats
	visited  map[string]struct{}
	pending  atomic.Int64
	start    time.Time
	duration time.Duration // 受 crawler.mu 保护
	err      error         // 受 crawler.mu 保护
}

// crawler 持有单次运行（可含多个目标）的共享状态
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
	targets []*targetState
}

// Run 采集单个目标，是 RunTargets 的便捷包装。
func (c *OpenDirClient) Run(ctx context.Context, target string) (*Stats, error) {
	results, err := c.RunTargets(ctx, []string{target})
	if err != nil {
		if len(results) == 1 {
			return results[0].Stats, err
		}
		return nil, err
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("opendir: expected 1 target result, got %d", len(results))
	}
	return results[0].Stats, results[0].Err
}

// RunTargets 并发采集多个目标。
//
// 所有目标共享同一个 worker 池（Config.Threads），因此总并发数不会随目标数增长。
// 返回值按输入顺序排列，重复目标会被去重；单个目标的失败记录在 TargetResult.Err，
// 不会中断其他目标。返回的 error 仅表示全局致命错误（ctx 取消、配置非法、targets 为空）。
func (c *OpenDirClient) RunTargets(ctx context.Context, targets []string) ([]TargetResult, error) {
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
	if len(targets) == 0 {
		return nil, errors.New("opendir: targets are required")
	}

	if cfg.MaxDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.MaxDuration)
		defer cancel()
	}

	cr := &crawler{
		config:  cfg,
		fetcher: c.fetcher,
		ctx:     ctx,
		tasks:   make(chan crawlTask, cfg.Threads),
		doneCh:  make(chan struct{}),
	}

	now := time.Now()
	seen := make(map[string]struct{}, len(targets))
	for _, raw := range targets {
		root, err := normalizeTarget(raw)
		if err != nil {
			cr.targets = append(cr.targets, &targetState{
				stats: &Stats{Target: strings.TrimSpace(raw)},
				start: now,
				err:   err,
			})
			continue
		}
		key := root.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cr.targets = append(cr.targets, &targetState{
			root:    root,
			stats:   &Stats{Target: key},
			visited: make(map[string]struct{}),
			start:   now,
		})
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
	for _, ts := range cr.targets {
		if ts.root == nil {
			continue
		}
		cr.enqueue(crawlTask{kind: taskRoot, url: ts.root, level: 1, owner: ts})
	}
	cr.pending.Add(-1)
	cr.finishIfIdle()

	select {
	case <-cr.doneCh:
	case <-ctx.Done():
	}
	cr.workers.Wait()

	cr.mu.Lock()
	results := make([]TargetResult, 0, len(cr.targets))
	for _, ts := range cr.targets {
		ts.stats.Duration = ts.duration
		results = append(results, TargetResult{
			Target: ts.stats.Target,
			Stats:  ts.stats,
			Err:    ts.err,
		})
	}
	cr.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

// enqueue 在目标内去重并调度任务；超过 MaxURLs 时停止入队
func (c *crawler) enqueue(task crawlTask) bool {
	if c.ctx.Err() != nil {
		return false
	}
	owner := task.owner
	key := task.url.String()

	c.mu.Lock()
	if owner.stats.Requests >= c.config.MaxURLs {
		owner.stats.Truncated = true
		c.mu.Unlock()
		return false
	}
	if _, seen := owner.visited[key]; seen {
		c.mu.Unlock()
		return false
	}
	owner.visited[key] = struct{}{}
	owner.stats.Requests++
	c.mu.Unlock()

	owner.pending.Add(1)
	c.pending.Add(1)
	select {
	case c.tasks <- task:
		return true
	case <-c.ctx.Done():
		owner.pending.Add(-1)
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

			owner := task.owner
			if owner.pending.Add(-1) == 0 {
				c.mu.Lock()
				owner.duration = time.Since(owner.start)
				c.mu.Unlock()
			}

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
	case taskRoot:
		c.processRoot(task)
	case taskDir:
		c.processDir(task)
	case taskFile:
		c.processFile(task)
	}
}

// processRoot 抓取并解析根目录页面；失败时记录目标级错误
func (c *crawler) processRoot(task crawlTask) {
	if c.ctx.Err() != nil {
		return
	}

	resp, err := c.fetcher.Get(c.ctx, task.url.String())
	if err != nil {
		c.setTargetError(task.owner, fmt.Errorf("opendir: fetch root: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.setTargetError(task.owner, fmt.Errorf("opendir: root returned status %d", resp.StatusCode))
		return
	}

	entries, err := parseListing(resp.URL, newLimitReader(resp.Body, maxListingBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		c.setTargetError(task.owner, fmt.Errorf("opendir: parse root: %w", err))
		return
	}
	if len(entries) == 0 && !isHTML(resp.Header.Get("Content-Type")) {
		c.setTargetError(task.owner, ErrNotListing)
		return
	}

	c.mu.Lock()
	task.owner.stats.DirsVisited++
	c.mu.Unlock()

	c.scheduleEntries(task.owner, entries, 1)
}

func (c *crawler) processDir(task crawlTask) {
	if c.ctx.Err() != nil {
		return
	}

	resp, err := c.fetcher.Get(c.ctx, task.url.String())
	if err != nil {
		c.recordRequestError(task.owner, task.url.String(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordRequestError(task.owner, task.url.String(), fmt.Errorf("unexpected status %d", resp.StatusCode))
		return
	}

	entries, err := parseListing(resp.URL, newLimitReader(resp.Body, maxListingBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		c.recordRequestError(task.owner, task.url.String(), fmt.Errorf("parse listing: %w", err))
		return
	}

	c.mu.Lock()
	task.owner.stats.DirsVisited++
	c.mu.Unlock()

	c.scheduleEntries(task.owner, entries, task.level)
}

func (c *crawler) processFile(task crawlTask) {
	if c.ctx.Err() != nil {
		return
	}

	resp, err := c.fetcher.Get(c.ctx, task.url.String())
	if err != nil {
		c.recordRequestError(task.owner, task.url.String(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordSkipped(task.owner, SkippedFile{
			URL:    task.url.String(),
			Name:   task.name,
			Reason: SkipStatus,
		})
		return
	}
	if c.config.MaxFileSize > 0 && resp.ContentLength > c.config.MaxFileSize {
		c.recordSkipped(task.owner, SkippedFile{
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
	task.owner.stats.Bytes += counter.n
	c.mu.Unlock()

	if errors.Is(handlerErr, errBodyTooLarge) {
		c.recordSkipped(task.owner, SkippedFile{
			URL:    task.url.String(),
			Name:   task.name,
			Reason: SkipTooLarge,
		})
		return
	}
	if handlerErr != nil {
		c.recordRequestError(task.owner, task.url.String(), handlerErr)
		return
	}

	c.mu.Lock()
	task.owner.stats.Files++
	c.mu.Unlock()
}

// scheduleEntries 调度一个目录页面中的条目
func (c *crawler) scheduleEntries(owner *targetState, entries []listingEntry, level int) {
	for _, entry := range entries {
		if entry.isDir {
			if level+1 > c.config.Depth {
				continue
			}
			task := crawlTask{kind: taskDir, url: entry.url, level: level + 1, owner: owner}
			if c.enqueue(task) {
				c.mu.Lock()
				owner.stats.DirsDiscovered++
				c.mu.Unlock()
			}
			continue
		}

		if reason, skip := c.skipReason(entry.name); skip {
			c.recordSkipped(owner, SkippedFile{
				URL:    entry.url.String(),
				Name:   entry.name,
				Reason: reason,
			})
			continue
		}
		c.enqueue(crawlTask{kind: taskFile, url: entry.url, name: entry.name, level: level, owner: owner})
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

// setTargetError 记录目标级致命错误；ctx 取消导致的错误不记录
func (c *crawler) setTargetError(owner *targetState, err error) {
	if c.ctx.Err() != nil {
		return
	}
	c.mu.Lock()
	if owner.err == nil {
		owner.err = err
	}
	c.mu.Unlock()
}

// recordRequestError 记录单 URL 错误；ctx 取消导致的错误不记录
func (c *crawler) recordRequestError(owner *targetState, url string, err error) {
	if c.ctx.Err() != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(owner.stats.Errors) < c.config.MaxErrors {
		owner.stats.Errors = append(owner.stats.Errors, FileError{URL: url, Err: err})
		return
	}
	owner.stats.ErrorsOmitted++
}

func (c *crawler) recordSkipped(owner *targetState, skipped SkippedFile) {
	c.mu.Lock()
	owner.stats.Skipped = append(owner.stats.Skipped, skipped)
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
