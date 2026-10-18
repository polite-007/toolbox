package opendir

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

// listingEntry 是目录列表页面中的一个条目
type listingEntry struct {
	url   *url.URL
	name  string
	isDir bool
}

// parseListing 从目录列表页面中提取条目。
// base 必须是该页面自身的 URL（重定向后的最终 URL），用于解析相对链接；
// contentType 用于识别页面编码。
func parseListing(base *url.URL, r io.Reader, contentType string) ([]listingEntry, error) {
	decoded, err := charset.NewReader(r, contentType)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(decoded)
	if err != nil {
		return nil, err
	}

	var entries []listingEntry
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		if entry, ok := resolveEntry(base, strings.TrimSpace(href)); ok {
			entries = append(entries, entry)
		}
	})
	return entries, nil
}

// resolveEntry 将 href 解析为绝对 URL，并过滤非目录列表条目
func resolveEntry(base *url.URL, href string) (listingEntry, bool) {
	if href == "" || strings.HasPrefix(href, "#") {
		return listingEntry{}, false
	}
	// 父目录 / 当前目录链接（Apache 等会在页面顶部生成）
	switch href {
	case ".", "./", "..", "../":
		return listingEntry{}, false
	}

	ref, err := url.Parse(href)
	if err != nil || ref.Opaque != "" {
		return listingEntry{}, false
	}
	// 只允许 http/https 链接
	if ref.Scheme != "" && ref.Scheme != "http" && ref.Scheme != "https" {
		return listingEntry{}, false
	}
	// 排序链接（如 IIS 的 ?C=N;O=D）不是文件
	if ref.RawQuery != "" {
		return listingEntry{}, false
	}

	resolved := base.ResolveReference(ref)
	resolved.Fragment = ""
	if !sameOrigin(resolved, base) {
		return listingEntry{}, false
	}
	if resolved.RawQuery != "" {
		return listingEntry{}, false
	}
	// 父目录或页面自身
	if strings.HasSuffix(resolved.Path, "/..") || resolved.String() == base.String() {
		return listingEntry{}, false
	}

	isDir := strings.HasSuffix(resolved.Path, "/")
	name := baseName(resolved, isDir)
	if name == "" {
		return listingEntry{}, false
	}
	return listingEntry{url: resolved, name: name, isDir: isDir}, true
}

// sameOrigin 判断两个 URL 是否同 scheme + host:port
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

// baseName 从 URL 中提取并解码文件名；目录会去掉尾部 "/"
func baseName(u *url.URL, isDir bool) string {
	p := u.EscapedPath()
	if isDir {
		p = strings.TrimSuffix(p, "/")
	}
	i := strings.LastIndexByte(p, '/')
	base := p[i+1:]
	if base == "" {
		return ""
	}
	if decoded, err := url.PathUnescape(base); err == nil {
		return decoded
	}
	return base
}

// fileDir 返回文件所在目录的绝对 URL（以 "/" 结尾）
func fileDir(u *url.URL) string {
	dir := *u
	p := dir.EscapedPath()
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		p = p[:i+1]
	} else {
		p = "/"
	}

	decoded, err := url.PathUnescape(p)
	if err != nil {
		return ""
	}
	dir.Path = decoded
	dir.RawPath = p
	dir.RawQuery = ""
	dir.Fragment = ""
	return dir.String()
}
