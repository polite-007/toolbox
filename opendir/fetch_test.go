package opendir

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetcherGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Errorf("User-Agent = %q, want test-agent", got)
		}
		if got := r.Header.Get("X-Test"); got != "1" {
			t.Errorf("X-Test = %q, want 1", got)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "hello world")
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{
		UserAgent: "test-agent",
		Headers:   []string{"X-Test: 1"},
	})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if string(body) != "hello world" {
		t.Errorf("body = %q, want %q", body, "hello world")
	}
	if resp.ContentLength != int64(len("hello world")) {
		t.Errorf("ContentLength = %d, want %d", resp.ContentLength, len("hello world"))
	}
}

func TestFetcherAddsDefaultScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestFetcherBodyTooLarge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", 100))
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{MaxBodySize: 10})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("ReadAll error = %v, want errBodyTooLarge", err)
	}
}

func TestFetcherTTFBTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{Timeout: 50 * time.Millisecond})
	defer f.Close()

	if _, err := f.Get(context.Background(), srv.URL); err == nil {
		t.Fatal("Get succeeded, want timeout error")
	}
}

func TestFetcherStallTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "a")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{StallTimeout: 100 * time.Millisecond})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if !errors.Is(err, errStalled) {
		t.Fatalf("ReadAll error = %v, want errStalled", err)
	}
}

func TestFetcherFollowsSameHostRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, "final")
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{FollowRedirects: true, MaxRedirects: defaultMaxRedirects})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "final" {
		t.Errorf("body = %q, want final", body)
	}
	if resp.URL.Path != "/final" {
		t.Errorf("final path = %q, want /final", resp.URL.Path)
	}
}

func TestFetcherRejectsCrossHostRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://localhost:1/elsewhere", http.StatusFound)
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{FollowRedirects: true, MaxRedirects: defaultMaxRedirects})
	defer f.Close()

	_, err := f.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("Get succeeded, want cross-host redirect error")
	}
	if !strings.Contains(err.Error(), "cross-host redirect") {
		t.Fatalf("error = %v, want cross-host redirect", err)
	}
}

func TestFetcherNoRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	}))
	defer srv.Close()

	f := newFetcher(fetchConfig{FollowRedirects: false})
	defer f.Close()

	resp, err := f.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want 302", resp.StatusCode)
	}
}

func TestFetcherContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	f := newFetcher(fetchConfig{Timeout: 5 * time.Second})
	defer f.Close()

	_, err := f.Get(ctx, srv.URL)
	if err == nil {
		t.Fatal("Get succeeded, want context error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want DeadlineExceeded", err)
	}
}
