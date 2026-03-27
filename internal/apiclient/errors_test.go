package apiclient

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      APIError
		expected string
	}{
		{
			name: "formats all fields",
			err: APIError{
				StatusCode: 404,
				Body:       `{"detail":"Not found."}`,
				Endpoint:   "/api/alerts/alert/1/",
				Method:     "GET",
			},
			expected: `MCS API GET /api/alerts/alert/1/ returned status 404: {"detail":"Not found."}`,
		},
		{
			name: "empty body",
			err: APIError{
				StatusCode: 500,
				Body:       "",
				Endpoint:   "/api/tenant/",
				Method:     "POST",
			},
			expected: "MCS API POST /api/tenant/ returned status 500: ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Error()
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "404 APIError",
			err:      &APIError{StatusCode: 404, Method: "GET", Endpoint: "/foo/"},
			expected: true,
		},
		{
			name:     "200 APIError",
			err:      &APIError{StatusCode: 200, Method: "GET", Endpoint: "/foo/"},
			expected: false,
		},
		{
			name:     "403 APIError",
			err:      &APIError{StatusCode: 403, Method: "GET", Endpoint: "/foo/"},
			expected: false,
		},
		{
			name:     "non-APIError",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsNotFound(tc.err)
			if got != tc.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "409 APIError",
			err:      &APIError{StatusCode: 409, Method: "POST", Endpoint: "/bar/"},
			expected: true,
		},
		{
			name:     "404 APIError",
			err:      &APIError{StatusCode: 404, Method: "POST", Endpoint: "/bar/"},
			expected: false,
		},
		{
			name:     "non-APIError",
			err:      errors.New("conflict-ish text"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsConflict(tc.err)
			if got != tc.expected {
				t.Errorf("IsConflict() = %v, want %v", got, tc.expected)
			}
		})
	}
}
