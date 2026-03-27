package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func testClient(handler http.Handler) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := NewClient(srv.URL, "test-token", false)
	c.httpClient = srv.Client()
	return c, srv
}

// ---------------------------------------------------------------------------
// NewClient
// ---------------------------------------------------------------------------

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c := NewClient("https://api.example.com/", "tok", false)
	if strings.HasSuffix(c.baseURL, "/") {
		t.Errorf("baseURL should not end with slash, got %q", c.baseURL)
	}
}

func TestNewClient_InsecureFalse(t *testing.T) {
	c := NewClient("https://api.example.com", "tok", false)
	tr := c.httpClient.Transport.(*http.Transport)
	if tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false")
	}
}

func TestNewClient_InsecureTrue(t *testing.T) {
	c := NewClient("https://api.example.com", "tok", true)
	tr := c.httpClient.Transport.(*http.Transport)
	if !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}

// ---------------------------------------------------------------------------
// Authentication & headers
// ---------------------------------------------------------------------------

func TestClient_SetsAuthHeaders(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization header = %q, want %q", got, "Token test-token")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_ = c.Get(context.Background(), "/test", nil)
}

// ---------------------------------------------------------------------------
// Trailing-slash normalization
// ---------------------------------------------------------------------------

func TestClient_TrailingSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/foo", "/api/foo/"},
		{"/api/foo/", "/api/foo/"},
		{"/api/foo?bar=1", "/api/foo/?bar=1"},
		{"/api/foo/?bar=1", "/api/foo/?bar=1"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := r.URL.RequestURI()
				if got != tc.expected {
					t.Errorf("path = %q, want %q", got, tc.expected)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			_ = c.Get(context.Background(), tc.input, nil)
		})
	}
}

// ---------------------------------------------------------------------------
// GET
// ---------------------------------------------------------------------------

func TestClient_Get_Success(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":1,"name":"test"}`)
	}))
	defer srv.Close()

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := c.Get(context.Background(), "/api/items/1/", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 || result.Name != "test" {
		t.Errorf("result = %+v, want {ID:1 Name:test}", result)
	}
}

func TestClient_Get_NilResult(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := c.Get(context.Background(), "/healthz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// POST
// ---------------------------------------------------------------------------

func TestClient_Post_SendsBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var p payload
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if p.Name != "new-item" {
			t.Errorf("body.Name = %q, want %q", p.Name, "new-item")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":42,"name":"new-item"}`)
	}))
	defer srv.Close()

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := c.Post(context.Background(), "/api/items/", payload{Name: "new-item"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 42 {
		t.Errorf("result.ID = %d, want 42", result.ID)
	}
}

// ---------------------------------------------------------------------------
// PUT
// ---------------------------------------------------------------------------

func TestClient_Put_Success(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":1,"name":"updated"}`)
	}))
	defer srv.Close()

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := c.Put(context.Background(), "/api/items/1/", map[string]string{"name": "updated"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "updated" {
		t.Errorf("result.Name = %q, want %q", result.Name, "updated")
	}
}

// ---------------------------------------------------------------------------
// PATCH
// ---------------------------------------------------------------------------

func TestClient_Patch_Success(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":1}`)
	}))
	defer srv.Close()

	var result struct{ ID int `json:"id"` }
	err := c.Patch(context.Background(), "/api/items/1/", map[string]string{"name": "patched"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DELETE
// ---------------------------------------------------------------------------

func TestClient_Delete_Success(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := c.Delete(context.Background(), "/api/items/1/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 4xx error handling (non-retryable)
// ---------------------------------------------------------------------------

func TestClient_4xx_ReturnsAPIError(t *testing.T) {
	codes := []int{400, 401, 403, 404, 409, 422}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(code)
				fmt.Fprint(w, `{"detail":"error"}`)
			}))
			defer srv.Close()

			err := c.Get(context.Background(), "/fail/", nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != code {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, code)
			}
		})
	}
}

func TestClient_4xx_NotRetried(t *testing.T) {
	var calls atomic.Int32
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	_ = c.Get(context.Background(), "/fail/", nil)
	if got := calls.Load(); got != 1 {
		t.Errorf("4xx should not be retried, got %d calls", got)
	}
}

// ---------------------------------------------------------------------------
// 429/5xx retry logic
// ---------------------------------------------------------------------------

func TestClient_429_IsRetried(t *testing.T) {
	var calls atomic.Int32
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	var result struct{ OK bool `json:"ok"` }
	err := c.Get(context.Background(), "/rate-limited/", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Error("expected result.OK to be true")
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("expected 3 calls (2 retries + success), got %d", got)
	}
}

func TestClient_500_IsRetried(t *testing.T) {
	var calls atomic.Int32
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := c.Get(context.Background(), "/flaky/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestClient_5xx_ExhaustsRetries(t *testing.T) {
	var calls atomic.Int32
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, "bad gateway")
	}))
	defer srv.Close()

	err := c.Get(context.Background(), "/always-fail/", nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "retries") {
		t.Errorf("error should mention retries, got: %v", err)
	}
	if got := calls.Load(); got != int32(maxRetries)+1 {
		t.Errorf("expected %d calls, got %d", maxRetries+1, got)
	}
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestClient_ContextCancelled(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Get(ctx, "/cancelled/", nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// ---------------------------------------------------------------------------
// Response decoding
// ---------------------------------------------------------------------------

func TestClient_InvalidJSON_ReturnsError(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `not-json`)
	}))
	defer srv.Close()

	var result struct{}
	err := c.Get(context.Background(), "/bad-json/", &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("error should mention decoding, got: %v", err)
	}
}

func TestClient_EmptyBody_NoDecodeError(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var result struct{ ID int }
	err := c.Get(context.Background(), "/empty/", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ListAll (pagination)
// ---------------------------------------------------------------------------

func TestClient_ListAll_SinglePage(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"count":2,"next":null,"previous":null,"results":[{"id":1},{"id":2}]}`)
	}))
	defer srv.Close()

	items, err := c.ListAll(context.Background(), "/api/things/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
}

func TestClient_ListAll_MultiplePages(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `{"count":3,"next":null,"previous":null,"results":[{"id":3}]}`)
		} else {
			nextURL := fmt.Sprintf("http://%s/api/things/?page=2", r.Host)
			fmt.Fprintf(w, `{"count":3,"next":"%s","previous":null,"results":[{"id":1},{"id":2}]}`, nextURL)
		}
	}))
	defer srv.Close()

	items, err := c.ListAll(context.Background(), "/api/things/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
	}
}

func TestClient_ListAll_EmptyResults(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"count":0,"next":null,"previous":null,"results":[]}`)
	}))
	defer srv.Close()

	items, err := c.ListAll(context.Background(), "/api/empty/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestClient_ListAll_APIError(t *testing.T) {
	c, srv := testClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"detail":"forbidden"}`)
	}))
	defer srv.Close()

	_, err := c.ListAll(context.Background(), "/api/forbidden/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// truncateBody
// ---------------------------------------------------------------------------

func TestTruncateBody(t *testing.T) {
	short := "short string"
	if got := truncateBody(short); got != short {
		t.Errorf("short string should not be truncated, got %q", got)
	}

	long := strings.Repeat("x", 600)
	got := truncateBody(long)
	if len(got) > 520 {
		t.Errorf("truncated body too long: %d chars", len(got))
	}
	if !strings.HasSuffix(got, "...(truncated)") {
		t.Errorf("truncated body should end with ...(truncated), got suffix %q", got[len(got)-20:])
	}
}
