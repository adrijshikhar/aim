package oauth

import (
	"encoding/json"
	"strings"
)

// RefusalCode defines typed OAuth refusal error categories that necessitate
// user re-login and trigger automated retry backoffs.
type RefusalCode string

const (
	RefusalNone                    RefusalCode = ""
	RefusalInvalidGrant            RefusalCode = "invalid_grant"
	RefusalRefreshTokenReused      RefusalCode = "refresh_token_reused"
	RefusalRefreshTokenExpired     RefusalCode = "refresh_token_expired"
	RefusalRefreshTokenInvalidated RefusalCode = "refresh_token_invalidated"
)

// ParseRefusal inspects HTTP status codes and response bodies from OAuth endpoints
// (token refreshes, token exchanges) and classifies fatal authorization failures
// into a standardized RefusalCode with a human-readable reason.
//
// Fatal refusals require HTTP 401, or HTTP 400/403 with an OAuth error payload
// matching recognized re-login codes. Generic 400 errors with non-OAuth bodies
// return RefusalNone.
func ParseRefusal(statusCode int, body []byte) (RefusalCode, string) {
	if statusCode != 400 && statusCode != 401 && statusCode != 403 {
		return RefusalNone, ""
	}

	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Message          string `json:"message"`
	}

	if len(body) > 0 && json.Unmarshal(body, &payload) == nil {
		reason := payload.ErrorDescription
		if reason == "" {
			reason = payload.Message
		}

		switch strings.ToLower(strings.TrimSpace(payload.Error)) {
		case "invalid_grant":
			return RefusalInvalidGrant, reason
		case "refresh_token_reused":
			return RefusalRefreshTokenReused, reason
		case "refresh_token_expired":
			return RefusalRefreshTokenExpired, reason
		case "refresh_token_invalidated":
			return RefusalRefreshTokenInvalidated, reason
		}

		if statusCode == 401 {
			if reason == "" {
				reason = strings.TrimSpace(string(body))
			}
			return RefusalInvalidGrant, reason
		}

		return RefusalNone, ""
	}

	if statusCode == 401 {
		return RefusalInvalidGrant, strings.TrimSpace(string(body))
	}

	return RefusalNone, ""
}
