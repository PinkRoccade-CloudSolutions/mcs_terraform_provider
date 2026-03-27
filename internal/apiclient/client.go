package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	maxRetries     = 3
	baseRetryDelay = 1 * time.Second
	maxRetryDelay  = 30 * time.Second
	requestTimeout = 60 * time.Second
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, insecure bool) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure, //nolint:gosec // user-configurable for self-signed certs
		},
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   requestTimeout,
		},
	}
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

func (c *Client) Patch(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// PaginatedResponse represents the envelope returned by paginated list endpoints.
type PaginatedResponse struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  json.RawMessage `json:"results"`
}

// ListAll fetches every page of a paginated endpoint and collects the raw JSON arrays.
func (c *Client) ListAll(ctx context.Context, path string) ([]json.RawMessage, error) {
	var all []json.RawMessage
	currentPath := path

	for {
		var page PaginatedResponse
		if err := c.Get(ctx, currentPath, &page); err != nil {
			return nil, err
		}

		var items []json.RawMessage
		if err := json.Unmarshal(page.Results, &items); err != nil {
			return nil, fmt.Errorf("parsing paginated results: %w", err)
		}
		all = append(all, items...)

		if page.Next == nil || *page.Next == "" {
			break
		}

		parsed, err := url.Parse(*page.Next)
		if err != nil {
			return nil, fmt.Errorf("parsing next page URL: %w", err)
		}
		currentPath = parsed.RequestURI()
	}

	return all, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	if idx := strings.IndexByte(path, '?'); idx != -1 {
		base := path[:idx]
		query := path[idx:]
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		path = base + query
	} else if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	fullURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseRetryDelay
			if delay > maxRetryDelay {
				delay = maxRetryDelay
			}
			tflog.Debug(ctx, "retrying request", map[string]interface{}{
				"attempt": attempt,
				"delay":   delay.String(),
			})

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			if body != nil {
				jsonBody, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonBody)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Token "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		tflog.Debug(ctx, "MCS API request", map[string]interface{}{
			"method": method,
			"url":    fullURL,
		})

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		tflog.Debug(ctx, "MCS API response", map[string]interface{}{
			"status": resp.StatusCode,
			"method": method,
			"url":    fullURL,
		})

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &APIError{
				StatusCode: resp.StatusCode,
				Body:       truncateBody(string(respBody)),
				Endpoint:   path,
				Method:     method,
			}
			continue
		}

		if resp.StatusCode >= 400 {
			return &APIError{
				StatusCode: resp.StatusCode,
				Body:       truncateBody(string(respBody)),
				Endpoint:   path,
				Method:     method,
			}
		}

		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

func truncateBody(body string) string {
	const maxLen = 500
	if len(body) > maxLen {
		return body[:maxLen] + "...(truncated)"
	}
	return body
}
