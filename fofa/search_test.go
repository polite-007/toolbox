package fofa

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestClient_SearchAll_MultiPageAndRequestImmutability(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		seenPages   []string
		seenSizes   []string
		seenFields  []string
		seenQueries []string
		callCount   int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		page := r.URL.Query().Get("page")
		size := r.URL.Query().Get("size")
		fields := r.URL.Query().Get("fields")
		seenPages = append(seenPages, page)
		seenSizes = append(seenSizes, size)
		seenFields = append(seenFields, fields)
		seenQueries = append(seenQueries, mustDecodeQuery(t, r.URL.Query().Get("qbase64")))
		mu.Unlock()

		var payload map[string]any
		switch page {
		case "2":
			payload = searchAllResponse(page, 2, "app.example", "443")
		default:
			payload = searchAllResponse("1", 2, "www.example", "80")
		}

		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("tester@example.com", "secret", WithBaseURL(server.URL), WithRetryCount(0))
	original := &SearchRequest{
		Query:  `title="login"`,
		Page:   1,
		Size:   1,
		Fields: "title",
	}
	before := *original

	var got []*SearchResult
	err := client.SearchAll(original, func(resp *SearchResponse) error {
		got = append(got, resp.GetResults()...)
		if len(got) == 1 {
			return nil
		}
		return errors.New("stop iteration")
	})
	if err == nil || err.Error() != "stop iteration" {
		t.Fatalf("SearchAll returned error = %v, want stop iteration", err)
	}

	if !reflect.DeepEqual(*original, before) {
		t.Fatalf("SearchAll mutated request: got %+v want %+v", *original, before)
	}
	if callCount != 2 {
		t.Fatalf("SearchAll request count = %d, want 2", callCount)
	}
	if !reflect.DeepEqual(seenPages, []string{"1", "2"}) {
		t.Fatalf("SearchAll pages = %v, want [1 2]", seenPages)
	}
	if !reflect.DeepEqual(seenSizes, []string{"1", "1"}) {
		t.Fatalf("SearchAll sizes = %v, want [1 1]", seenSizes)
	}
	if !reflect.DeepEqual(seenFields, []string{"title,ip,port", "title,ip,port"}) {
		t.Fatalf("SearchAll fields = %v, want [title,ip,port title,ip,port]", seenFields)
	}
	if !reflect.DeepEqual(seenQueries, []string{`title="login"`, `title="login"`}) {
		t.Fatalf("SearchAll queries = %v, want repeated original query", seenQueries)
	}
	if len(got) != 2 {
		t.Fatalf("SearchAll callback results = %d, want 2", len(got))
	}
	if got[0].Host != "www.example" || got[1].Host != "app.example" {
		t.Fatalf("SearchAll hosts = [%s %s], want [www.example app.example]", got[0].Host, got[1].Host)
	}
}

func TestClient_SearchAll_DefaultPageAndDefaultSize(t *testing.T) {
	t.Parallel()

	var (
		pageValue  string
		sizeValue  string
		callCount  int
		fieldValue string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		pageValue = r.URL.Query().Get("page")
		sizeValue = r.URL.Query().Get("size")
		fieldValue = r.URL.Query().Get("fields")

		if err := json.NewEncoder(w).Encode(searchAllResponse("1", 500, "www.example", "80")); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("tester@example.com", "secret", WithBaseURL(server.URL), WithRetryCount(0))

	err := client.SearchAll(&SearchRequest{Query: `title="dashboard"`}, func(resp *SearchResponse) error {
		return nil
	})
	if err != nil {
		t.Fatalf("SearchAll returned error: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("SearchAll request count = %d, want 1", callCount)
	}
	if pageValue != "1" {
		t.Fatalf("SearchAll default page = %q, want %q", pageValue, "1")
	}
	if sizeValue != "500" {
		t.Fatalf("SearchAll default size = %q, want %q", sizeValue, "500")
	}
	if fieldValue != "ip,port" {
		t.Fatalf("SearchAll default fields = %q, want %q", fieldValue, "ip,port")
	}
}

func TestClient_SearchAll_RetryOnceThenSucceed(t *testing.T) {
	t.Parallel()

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			if err := json.NewEncoder(w).Encode(map[string]any{
				"error":  true,
				"errmsg": "请求太快啦，请稍后再试",
			}); err != nil {
				t.Fatalf("encode rate-limit response: %v", err)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(searchAllResponse("1", 1, "retry.example", "443")); err != nil {
			t.Fatalf("encode success response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(
		"tester@example.com",
		"secret",
		WithBaseURL(server.URL),
		WithRetryCount(1),
		WithRetryInterval(10*time.Millisecond),
	)

	var callbackCalls int
	err := client.SearchAll(&SearchRequest{Query: `title="retry"`, Size: 1}, func(resp *SearchResponse) error {
		callbackCalls++
		if len(resp.GetResults()) != 1 {
			t.Fatalf("SearchAll callback results = %d, want 1", len(resp.GetResults()))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("SearchAll returned error: %v", err)
	}

	if callCount != 2 {
		t.Fatalf("SearchAll request count = %d, want 2", callCount)
	}
	if callbackCalls != 1 {
		t.Fatalf("SearchAll callback calls = %d, want 1", callbackCalls)
	}
}

func TestClient_SearchAll_CallbackCancellation(t *testing.T) {
	t.Parallel()

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewEncoder(w).Encode(searchAllResponse("1", 1, "cancel.example", "443")); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("tester@example.com", "secret", WithBaseURL(server.URL), WithRetryCount(0))

	stopErr := errors.New("stop iteration")
	err := client.SearchAll(&SearchRequest{Query: `title="cancel"`, Size: 1}, func(resp *SearchResponse) error {
		return stopErr
	})
	if !errors.Is(err, stopErr) {
		t.Fatalf("SearchAll returned error = %v, want stop iteration", err)
	}
	if callCount != 1 {
		t.Fatalf("SearchAll request count = %d, want 1", callCount)
	}
}

func TestClient_SearchAll_CallbackErrorPropagation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(searchAllResponse("1", 1, "err.example", "443")); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("tester@example.com", "secret", WithBaseURL(server.URL), WithRetryCount(0))
	wantErr := fmt.Errorf("callback failed")

	err := client.SearchAll(&SearchRequest{Query: `title="error"`, Size: 1}, func(resp *SearchResponse) error {
		return wantErr
	})
	if err == nil {
		t.Fatal("SearchAll error = nil, want callback error")
	}
	if err != wantErr {
		t.Fatalf("SearchAll error = %v, want %v", err, wantErr)
	}
}

func mustDecodeQuery(t *testing.T, encoded string) string {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode qbase64 %q: %v", encoded, err)
	}
	return string(decoded)
}

func searchAllResponse(page string, size int, host, port string) map[string]any {
	return map[string]any{
		"error": false,
		"mode":  "extended",
		"page":  mustAtoi(page),
		"size":  size,
		"query": "mock-query",
		"results": [][]string{{host, port}},
	}
}

func mustAtoi(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return n
}
