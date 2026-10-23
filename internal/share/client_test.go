package share

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_ResolveBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		override string
		expected string
	}{
		{
			name:     "plain ip without port",
			input:    "10.8.19.251",
			expected: "http://10.8.19.251:4319",
		},
		{
			name:     "ip with duplicate port danger",
			input:    "10.8.19.251:4319",
			expected: "http://10.8.19.251:4319",
		},
		{
			name:     "ip with non-default port",
			input:    "10.8.19.251:8080",
			expected: "http://10.8.19.251:8080",
		},
		{
			name:     "ipv6 address without port bracketed",
			input:    "fd7a:115c:a1e0::1",
			expected: "http://[fd7a:115c:a1e0::1]:4319",
		},
		{
			name:     "ipv6 address with port bracketed",
			input:    "[fd7a:115c:a1e0::1]:4319",
			expected: "http://[fd7a:115c:a1e0::1]:4319",
		},
		{
			name:     "hostname without port",
			input:    "Adrijs-MacBook-Pro-Personal.local",
			expected: "http://Adrijs-MacBook-Pro-Personal.local:4319",
		},
		{
			name:     "hostname with port",
			input:    "Adrijs-MacBook-Pro-Personal.local:4319",
			expected: "http://Adrijs-MacBook-Pro-Personal.local:4319",
		},
		{
			name:     "full http url with port",
			input:    "http://10.8.19.251:4319",
			expected: "http://10.8.19.251:4319",
		},
		{
			name:     "full http url with trailing slash",
			input:    "http://10.8.19.251:4319/",
			expected: "http://10.8.19.251:4319",
		},
		{
			name:     "override takes precedence",
			input:    "10.8.19.251:4319",
			override: "http://127.0.0.1:54321",
			expected: "http://127.0.0.1:54321",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(4319)
			if tc.override != "" {
				client.SetBaseURLOverride(tc.override)
			}
			got := client.resolveBaseURL(tc.input)
			if got != tc.expected {
				t.Errorf("resolveBaseURL(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestClient_NetworkFailureModes(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		ctxTimeout time.Duration
		call       func(ctx context.Context, c *Client, target string) error
		wantErrMsg string
	}{
		{
			name: "request lease 404 not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "profile not found", http.StatusNotFound)
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.RequestLease(ctx, target, "missing", "claude", 2)
				return err
			},
			wantErrMsg: "404",
		},
		{
			name: "request lease 500 internal server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "server boom", http.StatusInternalServerError)
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.RequestLease(ctx, target, "work", "claude", 2)
				return err
			},
			wantErrMsg: "500",
		},
		{
			name: "request lease corrupt json response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{invalid-json-body"))
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.RequestLease(ctx, target, "work", "claude", 2)
				return err
			},
			wantErrMsg: "invalid character",
		},
		{
			name: "request lease context timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			ctxTimeout: 20 * time.Millisecond,
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.RequestLease(ctx, target, "work", "claude", 2)
				return err
			},
			wantErrMsg: "context deadline exceeded",
		},
		{
			name: "poll lease 404 expired",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "request not found or expired", http.StatusNotFound)
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.PollLease(ctx, target, "req_expired")
				return err
			},
			wantErrMsg: "not found or expired",
		},
		{
			name: "poll lease corrupt json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("<html>bad gateway</html>"))
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.PollLease(ctx, target, "req_123")
				return err
			},
			wantErrMsg: "invalid character",
		},
		{
			name: "check lease status 500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "db failure", http.StatusInternalServerError)
			},
			call: func(ctx context.Context, c *Client, target string) error {
				_, err := c.CheckLeaseStatus(ctx, target, "ls_123")
				return err
			},
			wantErrMsg: "500",
		},
		{
			name: "cancel request 404 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			},
			call: func(ctx context.Context, c *Client, target string) error {
				return c.CancelLeaseRequest(ctx, target, "req_123")
			},
			wantErrMsg: "404",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			client := NewClient(4319)
			client.SetBaseURLOverride(ts.URL)

			ctx := context.Background()
			if tc.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			err := tc.call(ctx, client, ts.URL)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErrMsg)
			}
			if !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrMsg)
			}
		})
	}
}
