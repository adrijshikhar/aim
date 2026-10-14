package oauth_test

import (
	"testing"

	"github.com/aim-cli/aim/internal/oauth"
)

func TestParseRefusal(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantCode   oauth.RefusalCode
		wantReason string
	}{
		{
			name:       "Auth0 invalid_grant",
			status:     400,
			body:       `{"error":"invalid_grant","error_description":"Invalid refresh token"}`,
			wantCode:   oauth.RefusalInvalidGrant,
			wantReason: "Invalid refresh token",
		},
		{
			name:       "Anthropic refresh_token_reused",
			status:     400,
			body:       `{"error":"refresh_token_reused","message":"Token already consumed"}`,
			wantCode:   oauth.RefusalRefreshTokenReused,
			wantReason: "Token already consumed",
		},
		{
			name:       "Generic 400 Bad Request",
			status:     400,
			body:       `{"error":"invalid_request","error_description":"Missing parameter"}`,
			wantCode:   oauth.RefusalNone,
			wantReason: "",
		},
		{
			name:       "HTTP 401 Unauthorized default",
			status:     401,
			body:       `Unauthorized`,
			wantCode:   oauth.RefusalInvalidGrant,
			wantReason: "Unauthorized",
		},
		{
			name:       "HTTP 401 with JSON invalid_grant",
			status:     401,
			body:       `{"error":"invalid_grant","error_description":"Revoked refresh token"}`,
			wantCode:   oauth.RefusalInvalidGrant,
			wantReason: "Revoked refresh token",
		},
		{
			name:       "HTTP 401 with JSON refresh_token_reused",
			status:     401,
			body:       `{"error":"refresh_token_reused","message":"Reused token detected"}`,
			wantCode:   oauth.RefusalRefreshTokenReused,
			wantReason: "Reused token detected",
		},
		{
			name:       "HTTP 401 with empty body",
			status:     401,
			body:       ``,
			wantCode:   oauth.RefusalInvalidGrant,
			wantReason: "",
		},
		{
			name:       "HTTP 403 refresh_token_expired",
			status:     403,
			body:       `{"error":"refresh_token_expired","error_description":"The refresh token has expired"}`,
			wantCode:   oauth.RefusalRefreshTokenExpired,
			wantReason: "The refresh token has expired",
		},
		{
			name:       "HTTP 400 refresh_token_invalidated",
			status:     400,
			body:       `{"error":"refresh_token_invalidated","error_description":"Token invalidated by user"}`,
			wantCode:   oauth.RefusalRefreshTokenInvalidated,
			wantReason: "Token invalidated by user",
		},
		{
			name:       "HTTP 400 uppercase error code",
			status:     400,
			body:       `{"error":"INVALID_GRANT","error_description":"Token rejected"}`,
			wantCode:   oauth.RefusalInvalidGrant,
			wantReason: "Token rejected",
		},
		{
			name:       "HTTP 400 non-JSON body",
			status:     400,
			body:       `Bad Request: invalid headers`,
			wantCode:   oauth.RefusalNone,
			wantReason: "",
		},
		{
			name:       "HTTP 500 Server Error",
			status:     500,
			body:       `{"error":"invalid_grant","error_description":"Internal error"}`,
			wantCode:   oauth.RefusalNone,
			wantReason: "",
		},
		{
			name:       "HTTP 200 OK",
			status:     200,
			body:       `{"access_token":"abc"}`,
			wantCode:   oauth.RefusalNone,
			wantReason: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, reason := oauth.ParseRefusal(tc.status, []byte(tc.body))
			if code != tc.wantCode {
				t.Errorf("expected code %q, got %q", tc.wantCode, code)
			}
			if tc.wantReason != "" && reason != tc.wantReason {
				t.Errorf("expected reason %q, got %q", tc.wantReason, reason)
			}
		})
	}
}
