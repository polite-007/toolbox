package opendir

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// newFixtureServer 启动一个模拟目录服务器的测试服务。
// files 的 key 是解码后的路径，例如 "/目录/文件 1.txt"。
func newFixtureServer(t *testing.T, files map[string]string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/") {
			writeFixtureListing(w, path, files)
			return
		}
		content, ok := files[path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		_, _ = io.WriteString(w, content)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func writeFixtureListing(w http.ResponseWriter, dir string, files map[string]string) {
	children := map[string]bool{}
	for path := range files {
		if !strings.HasPrefix(path, dir) {
			continue
		}
		rest := strings.TrimPrefix(path, dir)
		if rest == "" {
			continue
		}
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			children[rest[:i]+"/"] = true
		} else {
			children[rest] = false
		}
	}

	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>Directory listing for %s</title></head><body><ul>", html.EscapeString(dir))
	for _, name := range names {
		href := (&url.URL{Path: name}).EscapedPath()
		fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, href, html.EscapeString(name))
	}
	fmt.Fprint(w, "</ul></body></html>")
}

type collectedFile struct {
	meta File
	body string
}

type collector struct {
	mu    sync.Mutex
	files []collectedFile
	errOn map[string]error
}

func newCollector() *collector {
	return &collector{errOn: map[string]error{}}
}

func (c *collector) handle(_ context.Context, f File, r io.Reader) error {
	if err := c.errOn[f.Name]; err != nil {
		return err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.files = append(c.files, collectedFile{meta: f, body: string(body)})
	c.mu.Unlock()
	return nil
}

func (c *collector) names() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	names := make([]string, 0, len(c.files))
	for _, f := range c.files {
		names = append(names, f.meta.Name)
	}
	sort.Strings(names)
	return names
}

func (c *collector) byName(name string) (collectedFile, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, f := range c.files {
		if f.meta.Name == name {
			return f, true
		}
	}
	return collectedFile{}, false
}

func TestRunDepth(t *testing.T) {
	files := map[string]string{
		"/root.txt":        "root",
		"/a.txt":           "a",
		"/sub/s1.txt":      "s1",
		"/sub/deep/d1.txt": "d1",
		"/other/o.txt":     "o",
	}
	srv := newFixtureServer(t, files)

	tests := []struct {
		depth     int
		wantFiles []string
		wantDirs  int
	}{
		{1, []string{"a.txt", "root.txt"}, 1},
		{2, []string{"a.txt", "o.txt", "root.txt", "s1.txt"}, 3},
		{3, []string{"a.txt", "d1.txt", "o.txt", "root.txt", "s1.txt"}, 4},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("depth=%d", tt.depth), func(t *testing.T) {
			col := newCollector()
			client := NewOpenDirClient(
				WithDepth(tt.depth),
				WithThreads(4),
				WithOnFile(col.handle),
			)
			stats, err := client.Run(context.Background(), srv.URL)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			got := col.names()
			if strings.Join(got, ",") != strings.Join(tt.wantFiles, ",") {
				t.Errorf("files = %v, want %v", got, tt.wantFiles)
			}
			if stats.DirsVisited != tt.wantDirs {
				t.Errorf("DirsVisited = %d, want %d", stats.DirsVisited, tt.wantDirs)
			}
			if stats.Files != len(tt.wantFiles) {
				t.Errorf("Files = %d, want %d", stats.Files, len(tt.wantFiles))
			}
		})
	}
}

func TestRunSkipRules(t *testing.T) {
	files := map[string]string{
		"/secret.txt": "s",
		"/Keep.txt":   "k",
		"/doc.pdf":    "p",
		"/pic.JPG":    "j",
		"/keep.md":    "m",
	}
	srv := newFixtureServer(t, files)

	col := newCollector()
	client := NewOpenDirClient(
		WithSkipNames([]string{"secret"}),
		WithSkipSuffixes([]string{"pdf", "JPG"}),
		WithOnFile(col.handle),
	)
	stats, err := client.Run(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	want := []string{"Keep.txt", "keep.md"}
	if got := col.names(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("files = %v, want %v", got, want)
	}
	if len(stats.Skipped) != 3 {
		t.Fatalf("Skipped = %d, want 3", len(stats.Skipped))
	}

	reasons := map[string]SkipReason{}
	for _, s := range stats.Skipped {
		reasons[s.Name] = s.Reason
	}
	if reasons["secret.txt"] != SkipName {
		t.Errorf("secret.txt reason = %q, want %q", reasons["secret.txt"], SkipName)
	}
	if reasons["doc.pdf"] != SkipSuffix {
		t.Errorf("doc.pdf reason = %q, want %q", reasons["doc.pdf"], SkipSuffix)
	}
	if reasons["pic.JPG"] != SkipSuffix {
		t.Errorf("pic.JPG reason = %q, want %q", reasons["pic.JPG"], SkipSuffix)
	}
}

func TestRunMaxURLs(t *testing.T) {
	files := map[string]string{
		"/a.txt":      "a",
		"/b.txt":      "b",
		"/sub/s1.txt": "s1",
	}
	srv := newFixtureServer(t, files)

	col := newCollector()
	client := NewOpenDirClient(
		WithDepth(2),
		WithMaxURLs(2),
		WithOnFile(col.handle),
	)
	stats, err := client.Run(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if stats.Requests != 2 {
		t.Errorf("Requests = %d, want 2", stats.Requests)
	}
	if !stats.Truncated {
		t.Error("Truncated = false, want true")
	}
	if len(col.names()) != 1 {
		t.Errorf("files = %v, want 1 file", col.names())
	}
}

func TestRunHandlerErrorContinues(t *testing.T) {
	files := map[string]string{
		"/a.txt": "a",
		"/b.txt": "b",
	}
	srv := newFixtureServer(t, files)

	col := newCollector()
	col.errOn["a.txt"] = errors.New("boom")

	client := NewOpenDirClient(WithOnFile(col.handle))
	stats, err := client.Run(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if got := col.names(); len(got) != 1 || got[0] != "b.txt" {
		t.Errorf("files = %v, want [b.txt]", got)
	}
	if stats.Files != 1 {
		t.Errorf("Files = %d, want 1", stats.Files)
	}
	if len(stats.Errors) != 1 {
		t.Fatalf("Errors = %v, want 1 entry", stats.Errors)
	}
	if stats.Errors[0].URL != srv.URL+"/a.txt" {
		t.Errorf("error URL = %q, want %q", stats.Errors[0].URL, srv.URL+"/a.txt")
	}
}

func TestRunNotListing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "not a listing")
	}))
	defer srv.Close()

	client := NewOpenDirClient(WithOnFile(newCollector().handle))
	_, err := client.Run(context.Background(), srv.URL)
	if !errors.Is(err, ErrNotListing) {
		t.Fatalf("err = %v, want ErrNotListing", err)
	}
}

func TestRunFileMetadata(t *testing.T) {
	files := map[string]string{
		"/目录/文件 1.txt": "content",
	}
	srv := newFixtureServer(t, files)

	col := newCollector()
	client := NewOpenDirClient(WithDepth(2), WithOnFile(col.handle))
	stats, err := client.Run(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if stats.Files != 1 {
		t.Fatalf("Files = %d, want 1", stats.Files)
	}

	got, ok := col.byName("文件 1.txt")
	if !ok {
		t.Fatalf("file not collected: %v", col.names())
	}
	if got.body != "content" {
		t.Errorf("body = %q, want content", got.body)
	}
	if got.meta.Dir != srv.URL+"/%E7%9B%AE%E5%BD%95/" {
		t.Errorf("Dir = %q, want %q", got.meta.Dir, srv.URL+"/%E7%9B%AE%E5%BD%95/")
	}
	wantURL, _ := url.Parse(srv.URL + "/目录/文件 1.txt")
	if got.meta.URL != wantURL.String() {
		t.Errorf("URL = %q, want %q", got.meta.URL, wantURL.String())
	}
	if got.meta.Size != int64(len("content")) {
		t.Errorf("Size = %d, want %d", got.meta.Size, len("content"))
	}
}

func TestRunTooLargeSkipped(t *testing.T) {
	files := map[string]string{
		"/big.bin": "0123456789",
	}
	srv := newFixtureServer(t, files)

	col := newCollector()
	client := NewOpenDirClient(
		WithMaxFileSize(4),
		WithOnFile(col.handle),
	)
	stats, err := client.Run(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(col.names()) != 0 {
		t.Errorf("files = %v, want none", col.names())
	}
	if len(stats.Skipped) != 1 || stats.Skipped[0].Reason != SkipTooLarge {
		t.Fatalf("Skipped = %+v, want one too_large", stats.Skipped)
	}
}

func TestRunContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client := NewOpenDirClient(
		WithTimeout(5),
		WithOnFile(newCollector().handle),
	)
	if _, err := client.Run(ctx, srv.URL); err == nil {
		t.Fatal("Run succeeded, want context error")
	}
}

func TestRunRequiresHandler(t *testing.T) {
	client := NewOpenDirClient()
	if _, err := client.Run(context.Background(), "http://example.com/"); !errors.Is(err, ErrNoHandler) {
		t.Fatalf("err = %v, want ErrNoHandler", err)
	}
}

func TestRunInvalidTarget(t *testing.T) {
	client := NewOpenDirClient(WithOnFile(newCollector().handle))
	for _, target := range []string{"", "ftp://example.com", "http://"} {
		if _, err := client.Run(context.Background(), target); !errors.Is(err, ErrInvalidTarget) {
			t.Errorf("target %q: err = %v, want ErrInvalidTarget", target, err)
		}
	}
}
