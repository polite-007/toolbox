package opendir

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseListingPythonStyle(t *testing.T) {
	const page = `<!DOCTYPE HTML>
<html><head><title>Directory listing for /</title></head><body>
<ul>
<li><a href=".acp-file-refs.json">.acp-file-refs.json</a></li>
<li><a href="conversations/">conversations/</a></li>
<li><a href="20260630%20%E5%AE%89%E5%85%8B.pdf">20260630 安克.pdf</a></li>
<li><a href="a%26b.txt">a&amp;b.txt</a></li>
</ul>
</body></html>`

	base, _ := url.Parse("http://example.com:8765/")
	entries, err := parseListing(base, strings.NewReader(page), "text/html; charset=utf-8")
	if err != nil {
		t.Fatalf("parseListing failed: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("entries = %d, want 4", len(entries))
	}

	want := []struct {
		name  string
		isDir bool
		url   string
	}{
		{".acp-file-refs.json", false, "http://example.com:8765/.acp-file-refs.json"},
		{"conversations", true, "http://example.com:8765/conversations/"},
		{"20260630 安克.pdf", false, "http://example.com:8765/20260630%20%E5%AE%89%E5%85%8B.pdf"},
		{"a&b.txt", false, "http://example.com:8765/a%26b.txt"},
	}
	for i, w := range want {
		if entries[i].name != w.name {
			t.Errorf("entries[%d].name = %q, want %q", i, entries[i].name, w.name)
		}
		if entries[i].isDir != w.isDir {
			t.Errorf("entries[%d].isDir = %v, want %v", i, entries[i].isDir, w.isDir)
		}
		if got := entries[i].url.String(); got != w.url {
			t.Errorf("entries[%d].url = %q, want %q", i, got, w.url)
		}
	}
}

func TestParseListingExclusions(t *testing.T) {
	const page = `<html><body><ul>
<li><a href="../">../</a></li>
<li><a href="./">./</a></li>
<li><a href="#top">top</a></li>
<li><a href="?C=N;O=D">Name</a></li>
<li><a href="mailto:someone@example.com">mail</a></li>
<li><a href="javascript:void(0)">js</a></li>
<li><a href="http://other.example.com/x.txt">x.txt</a></li>
<li><a href="/absolute/file.txt">file.txt</a></li>
<li><a href="file.txt">file.txt</a></li>
<li><a href="dir/">dir/</a></li>
</ul></body></html>`

	base, _ := url.Parse("http://example.com/files/")
	entries, err := parseListing(base, strings.NewReader(page), "text/html")
	if err != nil {
		t.Fatalf("parseListing failed: %v", err)
	}

	got := map[string]bool{}
	for _, e := range entries {
		got[e.name] = e.isDir
	}
	want := map[string]bool{
		"file.txt": false,
		"dir":      true,
	}
	if len(got) != len(want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
	for name, isDir := range want {
		dir, ok := got[name]
		if !ok {
			t.Errorf("missing entry %q", name)
			continue
		}
		if dir != isDir {
			t.Errorf("%q isDir = %v, want %v", name, dir, isDir)
		}
	}
}

func TestParseListingRelativeBase(t *testing.T) {
	const page = `<html><body><a href="child/">child/</a><a href="file.txt">file.txt</a></body></html>`

	// 目录页面缺少尾斜杠时，相对链接以文件的父目录为基准
	base, _ := url.Parse("http://example.com/root/sub")
	entries, err := parseListing(base, strings.NewReader(page), "text/html")
	if err != nil {
		t.Fatalf("parseListing failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if got := entries[0].url.String(); got != "http://example.com/root/child/" {
		t.Errorf("child url = %q, want http://example.com/root/child/", got)
	}
}

func TestFileDir(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"http://example.com/file.txt", "http://example.com/"},
		{"http://example.com/a/b/file.txt", "http://example.com/a/b/"},
		{"http://example.com/%E7%9B%AE%E5%BD%95/file.txt", "http://example.com/%E7%9B%AE%E5%BD%95/"},
	}
	for _, tt := range tests {
		u, _ := url.Parse(tt.file)
		if got := fileDir(u); got != tt.want {
			t.Errorf("fileDir(%q) = %q, want %q", tt.file, got, tt.want)
		}
	}
}

func TestSkipReason(t *testing.T) {
	c := &crawler{config: &Config{
		SkipNames:    []string{"简历", "secret"},
		SkipSuffixes: []string{"pdf", ".tar.gz"},
	}}

	tests := []struct {
		name   string
		reason SkipReason
		skip   bool
	}{
		{"张三 简历.pdf", SkipName, true},
		{"mysecret.txt", SkipName, true},
		{"Secret.txt", "", false},
		{"a.PDF", "", false},
		{"a.pdf", SkipSuffix, true},
		{"archive.tar.gz", SkipSuffix, true},
		{"keep.txt", "", false},
	}
	for _, tt := range tests {
		reason, skip := c.skipReason(tt.name)
		if skip != tt.skip || reason != tt.reason {
			t.Errorf("skipReason(%q) = (%q, %v), want (%q, %v)", tt.name, reason, skip, tt.reason, tt.skip)
		}
	}
}
