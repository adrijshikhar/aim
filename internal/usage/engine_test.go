package usage

import (
	"testing"

	"github.com/aim-cli/aim/internal/oauth"
)

func TestParseRefusalFromError_HTTP401Boundaries(t *testing.T) {
	tests := []struct {
		name         string
		errStr       string
		expectedCode oauth.RefusalCode
	}{
		{
			name:         "port number with 401 should not match",
			errStr:       "connect to 127.0.0.1:4012: connection refused",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "timestamp with 401 should not match",
			errStr:       "request failed at 14:01:00 UTC",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "byte count with 401 should not match",
			errStr:       "read 4012 bytes from socket",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "request ID with 401 should not match",
			errStr:       "request-id: abc-4012-xyz failed",
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "space separated 401",
			errStr:       "received 401 error from server",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "prefix 401",
			errStr:       "401 unauthorized",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status 401",
			errStr:       "HTTP status 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status: 401",
			errStr:       "status: 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "status code: 401",
			errStr:       "HTTP status code: 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "http 401",
			errStr:       "http 401 returned",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "JSON error with port 4012 should not trigger 401 refusal",
			errStr:       `{"error": "network_failure", "endpoint": "127.0.0.1:4012"}`,
			expectedCode: oauth.RefusalNone,
		},
		{
			name:         "trailing suffix 401",
			errStr:       "request failed with 401",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "colon separated 401",
			errStr:       "received 401: token revoked",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "parenthesized 401",
			errStr:       "HTTP error (401)",
			expectedCode: oauth.RefusalInvalidGrant,
		},
		{
			name:         "bracketed 401",
			errStr:       "[401] Token rejected",
			expectedCode: oauth.RefusalInvalidGrant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := parseRefusalFromError(tt.errStr)
			if code != tt.expectedCode {
				t.Fatalf("expected code %v, got %v for error %q", tt.expectedCode, code, tt.errStr)
			}
		})
	}
}
