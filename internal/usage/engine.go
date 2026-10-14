package usage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/oauth"
)

type TargetProfile struct {
	Agent      string
	Profile    string
	ProfileDir string
	GetUsageFn func(ctx context.Context, profileName, profileDir string) (*Report, error)
}

func RefreshAsync(ctx context.Context, targets []TargetProfile, cache *CacheStore, force ...bool) <-chan Report {
	out := make(chan Report, len(targets))

	isForce := len(force) > 0 && force[0]

	var refStore *oauth.RefusalStore
	if cache != nil {
		refStore = cache.RefusalStore()
	}

	var pending []TargetProfile
	for _, target := range targets {
		if refStore != nil {
			if rec, active := refStore.IsRefused(target.Agent, target.Profile); active && rec != nil {
				rep := Report{
					Agent:     target.Agent,
					Profile:   target.Profile,
					Status:    StatusUnknown,
					FetchedAt: time.Now(),
					Error:     "re-login required (" + string(rec.Code) + ")",
					Summary:   "Re-login required: " + string(rec.Code),
				}
				if cache != nil {
					_ = cache.Put(rep)
				}
				out <- rep
				continue
			}
		}
		if cache != nil && !isForce {
			if rep, found := cache.Get(target.Agent, target.Profile); found {
				out <- rep
				continue
			}
		}
		pending = append(pending, target)
	}

	if len(pending) == 0 {
		close(out)
		return out
	}

	go func() {
		defer close(out)

		concurrency := 4
		if len(pending) < concurrency {
			concurrency = len(pending)
		}
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

	WorkLoop:
		for _, target := range pending {
			select {
			case <-ctx.Done():
				break WorkLoop
			case sem <- struct{}{}:
			}

			wg.Add(1)
			go func(t TargetProfile) {
				defer func() {
					<-sem
					wg.Done()
				}()

				var refStore *oauth.RefusalStore
				if cache != nil {
					refStore = cache.RefusalStore()
				}

				if refStore != nil {
					if rec, active := refStore.IsRefused(t.Agent, t.Profile); active && rec != nil {
						rep := &Report{
							Agent:     t.Agent,
							Profile:   t.Profile,
							Status:    StatusUnknown,
							FetchedAt: time.Now(),
							Error:     "re-login required (" + string(rec.Code) + ")",
							Summary:   "Re-login required: " + string(rec.Code),
						}
						if cache != nil && ctx.Err() == nil {
							_ = cache.Put(*rep)
						}
						select {
						case <-ctx.Done():
						case out <- *rep:
						}
						return
					}
				}

				timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				var rep *Report
				var err error
				if t.GetUsageFn != nil {
					rep, err = t.GetUsageFn(timeoutCtx, t.Profile, t.ProfileDir)
				}

				if err != nil || rep == nil {
					rep = &Report{
						Agent:     t.Agent,
						Profile:   t.Profile,
						Status:    StatusUnknown,
						FetchedAt: time.Now(),
						Error:     "unable to fetch usage",
					}
					if err != nil {
						rep.Error = err.Error()
					}
				} else {
					if rep.Agent == "" {
						rep.Agent = t.Agent
					}
					if rep.Profile == "" {
						rep.Profile = t.Profile
					}
				}

				if rep != nil && rep.Status != StatusUnknown && err == nil {
					if refStore != nil {
						if err := refStore.Clear(t.Agent, t.Profile); err != nil {
							logger.Debug("failed to clear refusal for %s:%s: %v", t.Agent, t.Profile, err)
						}
					}
				} else if rep != nil {
					errStr := rep.Error
					if errStr == "" && err != nil {
						errStr = err.Error()
					}
					if errStr == "" {
						errStr = rep.Summary
					}
					if code, desc := parseRefusalFromError(errStr); code != oauth.RefusalNone {
						if refStore != nil {
							if err := refStore.Record(t.Agent, t.Profile, code, desc); err != nil {
								logger.Debug("failed to persist refusal for %s:%s: %v", t.Agent, t.Profile, err)
							}
						}
						rep.Error = "re-login required (" + string(code) + ")"
						rep.Summary = "Re-login required: " + string(code)
					}
				}

				if cache != nil && ctx.Err() == nil {
					// Stale-While-Revalidate: If fetch failed with StatusUnknown, but cache already contains
					// a valid, healthy report with quota windows, preserve the existing valid quota!
					if rep.Status == StatusUnknown {
						errLower := strings.ToLower(rep.Error)
						summaryLower := strings.ToLower(rep.Summary)
						combinedErr := errLower + " " + summaryLower

						isAuthError := strings.Contains(combinedErr, "unauthorized") ||
							strings.Contains(combinedErr, "invalid_grant") ||
							strings.Contains(combinedErr, "refresh_token") ||
							strings.Contains(combinedErr, "credential") ||
							strings.Contains(combinedErr, "re-auth") ||
							strings.Contains(combinedErr, "login")

						isTransient := !isAuthError && (errors.Is(ctx.Err(), context.DeadlineExceeded) ||
							strings.Contains(combinedErr, "timeout") ||
							strings.Contains(combinedErr, "deadline exceeded") ||
							strings.Contains(combinedErr, "connection refused") ||
							strings.Contains(combinedErr, "network is unreachable") ||
							strings.Contains(combinedErr, "no route to host") ||
							strings.Contains(combinedErr, "temporary") ||
							strings.Contains(combinedErr, "503") ||
							strings.Contains(combinedErr, "unavailable") ||
							strings.Contains(combinedErr, "service unavailable") ||
							strings.Contains(combinedErr, "offline") ||
							strings.Contains(combinedErr, "reset by peer"))

						if isTransient {
							if existing, found := cache.GetStale(t.Agent, t.Profile); found && existing.Status != StatusUnknown && len(existing.Windows) > 0 {
								// Retain existing valid report
								select {
								case <-ctx.Done():
								case out <- existing:
								}
								return
							}
						}
					}
					if err := cache.Put(*rep); err != nil {
						logger.Debug("failed to write usage cache for %s:%s: %v", rep.Agent, rep.Profile, err)
					}
				}

				select {
				case <-ctx.Done():
				case out <- *rep:
				}
			}(target)
		}

		wg.Wait()
	}()

	return out
}

func parseRefusalFromError(errStr string) (oauth.RefusalCode, string) {
	if errStr == "" {
		return oauth.RefusalNone, ""
	}

	trimmed := strings.TrimSpace(errStr)
	lower := strings.ToLower(trimmed)

	// If error string contains a JSON object, attempt OAuth JSON parsing
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		jsonBlob := []byte(trimmed[start : end+1])
		if code, desc := oauth.ParseRefusal(400, jsonBlob); code != oauth.RefusalNone {
			if desc == "" {
				desc = trimmed
			}
			return code, desc
		}
		// Only evaluate 401 fallback if the error explicitly indicates 401 / unauthorized
		if isHTTP401(lower) || strings.Contains(lower, "unauthorized") {
			if code, desc := oauth.ParseRefusal(401, jsonBlob); code != oauth.RefusalNone {
				if desc == "" {
					desc = trimmed
				}
				return code, desc
			}
		}
	}

	switch {
	case strings.Contains(lower, string(oauth.RefusalRefreshTokenReused)):
		return oauth.RefusalRefreshTokenReused, trimmed
	case strings.Contains(lower, string(oauth.RefusalRefreshTokenExpired)):
		return oauth.RefusalRefreshTokenExpired, trimmed
	case strings.Contains(lower, string(oauth.RefusalRefreshTokenInvalidated)):
		return oauth.RefusalRefreshTokenInvalidated, trimmed
	case strings.Contains(lower, string(oauth.RefusalInvalidGrant)):
		return oauth.RefusalInvalidGrant, trimmed
	case isHTTP401(lower) || strings.Contains(lower, "unauthorized"):
		return oauth.RefusalInvalidGrant, trimmed
	}

	return oauth.RefusalNone, ""
}

func isHTTP401(lower string) bool {
	return strings.Contains(lower, " 401 ") ||
		strings.HasPrefix(lower, "401 ") ||
		strings.HasSuffix(lower, " 401") ||
		strings.Contains(lower, "401:") ||
		strings.Contains(lower, "(401)") ||
		strings.Contains(lower, "[401]") ||
		strings.Contains(lower, "status 401") ||
		strings.Contains(lower, "status: 401") ||
		strings.Contains(lower, "status code: 401") ||
		strings.Contains(lower, "status code 401") ||
		strings.Contains(lower, "http 401") ||
		strings.Contains(lower, "401 unauthorized")
}

// InspectProfile performs a strict read-only profile inspection with zero network calls and
// zero file modifications. It respects active refusal backoffs, passes through fresh cache,
// classifies stale cache with prior StatusOK as StatusIdle ("Ready on use"), and falls back
// to inspecting credentials on disk.
func InspectProfile(ctx context.Context, target TargetProfile, cache *CacheStore) Report {
	if ctx != nil && ctx.Err() != nil {
		return Report{
			Agent:     target.Agent,
			Profile:   target.Profile,
			Status:    StatusUnknown,
			FetchedAt: time.Now(),
			Error:     ctx.Err().Error(),
		}
	}

	var refStore *oauth.RefusalStore
	if cache != nil {
		refStore = cache.RefusalStore()
	}

	// 1. Check refusal store: active refusal takes precedence
	if refStore != nil {
		if rec, active := refStore.IsRefused(target.Agent, target.Profile); active && rec != nil {
			return Report{
				Agent:     target.Agent,
				Profile:   target.Profile,
				Status:    StatusUnknown,
				FetchedAt: time.Now(),
				Error:     "re-login required (" + string(rec.Code) + ")",
				Summary:   "Re-login required: " + string(rec.Code),
			}
		}
	}

	// 2. Fresh cache check
	if cache != nil {
		if rep, found := cache.Get(target.Agent, target.Profile); found {
			return rep
		}
	}

	// 3. Stale cache check (NEVER invoke target.GetUsageFn)
	if cache != nil {
		if rep, found := cache.GetStale(target.Agent, target.Profile); found && rep.Status != StatusUnknown && len(rep.Windows) > 0 {
			// If previously exhausted and reset time has not yet passed, retain StatusExhausted
			if rep.Status == StatusExhausted {
				stillExhausted := false
				for _, w := range rep.Windows {
					if w.RemainingPct <= 0 && (w.ResetsAt.IsZero() || time.Now().Before(w.ResetsAt)) {
						stillExhausted = true
						break
					}
				}
				if stillExhausted {
					rep.FromCache = true
					if rep.Agent == "" {
						rep.Agent = target.Agent
					}
					if rep.Profile == "" {
						rep.Profile = target.Profile
					}
					return rep
				}
			}

			rep.Status = StatusIdle
			rep.FromCache = true
			if rep.Summary == "" {
				rep.Summary = "Ready on use"
			}
			if rep.Agent == "" {
				rep.Agent = target.Agent
			}
			if rep.Profile == "" {
				rep.Profile = target.Profile
			}
			return rep
		}
	}

	// 4. Inspect credentials on disk (unrefused)
	if HasCredentialsOnDisk(target.Agent, target.ProfileDir) {
		return Report{
			Agent:     target.Agent,
			Profile:   target.Profile,
			Status:    StatusIdle,
			Summary:   "Ready on use",
			FetchedAt: time.Now(),
		}
	}

	// 5. No credentials on disk
	return Report{
		Agent:     target.Agent,
		Profile:   target.Profile,
		Status:    StatusUnknown,
		Error:     "no credentials",
		FetchedAt: time.Now(),
	}
}

// HasCredentialsOnDisk checks if the specified agent has valid, non-empty credentials on disk in profileDir.
// It is strictly read-only: zero network calls, zero files created, modified, or auto-seeded.
func HasCredentialsOnDisk(agent, profileDir string) bool {
	if profileDir == "" {
		return false
	}
	fi, err := os.Stat(profileDir)
	if err != nil {
		return false
	}
	if !fi.IsDir() {
		return isFileWithCredentials(profileDir)
	}

	agentNorm := strings.ToLower(strings.TrimSpace(agent))
	var candidates []string
	switch agentNorm {
	case "codex", "codex-cli", "openai-codex":
		candidates = []string{
			filepath.Join(profileDir, ".codex", "auth.json"),
			filepath.Join(profileDir, "auth.json"),
		}
	case "claude", "claude-code":
		candidates = []string{
			filepath.Join(profileDir, ".claude", "auth.json"),
			filepath.Join(profileDir, ".claude", ".credentials.json"),
			filepath.Join(profileDir, "auth.json"),
			filepath.Join(profileDir, ".credentials.json"),
		}
	case "gemini", "gemini-cli":
		candidates = []string{
			filepath.Join(profileDir, ".gemini", "gemini-oauth-token"),
			filepath.Join(profileDir, "gemini-oauth-token"),
		}
	case "agy", "antigravity":
		candidates = []string{
			filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
			filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json"),
			filepath.Join(profileDir, "antigravity-oauth-token"),
		}
	case "":
		// When agent is empty, check all known locations
		candidates = []string{
			filepath.Join(profileDir, ".codex", "auth.json"),
			filepath.Join(profileDir, ".claude", "auth.json"),
			filepath.Join(profileDir, ".claude", ".credentials.json"),
			filepath.Join(profileDir, ".gemini", "gemini-oauth-token"),
			filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token"),
			filepath.Join(profileDir, ".config", "gcloud", "application_default_credentials.json"),
			filepath.Join(profileDir, "auth.json"),
			filepath.Join(profileDir, "credentials.json"),
			filepath.Join(profileDir, ".credentials.json"),
			filepath.Join(profileDir, "token.json"),
			filepath.Join(profileDir, ".token.json"),
		}
	default:
		candidates = []string{
			filepath.Join(profileDir, agentNorm+"-oauth-token"),
			filepath.Join(profileDir, "."+agentNorm, "auth.json"),
			filepath.Join(profileDir, "."+agentNorm, "token.json"),
			filepath.Join(profileDir, "auth.json"),
			filepath.Join(profileDir, "credentials.json"),
			filepath.Join(profileDir, "token.json"),
		}
	}

	for _, p := range candidates {
		if isFileWithCredentials(p) {
			return true
		}
	}
	return false
}

func isFileWithCredentials(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() || fi.Size() == 0 || fi.Size() > 1<<20 {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return false
	}

	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 {
		return false
	}

	// If file ends in .json or has JSON syntax
	if strings.HasSuffix(strings.ToLower(path), ".json") || strings.HasPrefix(trimmed, "{") {
		var raw any
		if err := json.Unmarshal(data, &raw); err != nil {
			return false
		}
		return hasJSONCredentials(raw)
	}

	return len(trimmed) > 0
}

func hasJSONCredentials(raw any) bool {
	if raw == nil {
		return false
	}
	m, ok := raw.(map[string]any)
	if !ok || len(m) == 0 {
		return false
	}

	// Check nested "tokens" map (e.g. Codex auth.json)
	if tokens, ok := m["tokens"].(map[string]any); ok && len(tokens) > 0 {
		for _, k := range []string{"access_token", "refresh_token", "id_token", "token"} {
			if s, ok := tokens[k].(string); ok && strings.TrimSpace(s) != "" {
				return true
			}
		}
	}

	// Check nested "token" map (e.g. Antigravity token)
	if tokenObj, ok := m["token"].(map[string]any); ok && len(tokenObj) > 0 {
		for _, k := range []string{"access_token", "refresh_token", "id_token", "token"} {
			if s, ok := tokenObj[k].(string); ok && strings.TrimSpace(s) != "" {
				return true
			}
		}
	}

	// Check Claude oauth
	if oauthObj, ok := m["claudeAiOauth"].(map[string]any); ok && len(oauthObj) > 0 {
		for _, k := range []string{"accessToken", "refreshToken", "access_token", "refresh_token", "token"} {
			if s, ok := oauthObj[k].(string); ok && strings.TrimSpace(s) != "" {
				return true
			}
		}
	}

	// Check direct token / key fields
	for _, k := range []string{
		"access_token", "refresh_token", "id_token", "token",
		"api_key", "apiKey", "openai_api_key",
		"session_key", "sessionKey",
		"client_email", "private_key",
	} {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return true
		}
	}

	// Check Google ADC service_account or authorized_user
	if t, ok := m["type"].(string); ok && (t == "authorized_user" || t == "service_account") {
		return true
	}

	return false
}
