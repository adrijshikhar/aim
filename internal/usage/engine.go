package usage

import (
	"context"
	"errors"
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

	var pending []TargetProfile
	for _, target := range targets {
		if cache != nil {
			if refStore := cache.RefusalStore(); refStore != nil {
				if rec, active := refStore.IsRefused(target.Agent, target.Profile); active && rec != nil {
					rep := Report{
						Agent:     target.Agent,
						Profile:   target.Profile,
						Status:    StatusUnknown,
						FetchedAt: time.Now(),
						Error:     "re-login required (" + string(rec.Code) + ")",
						Summary:   "Re-login required: " + string(rec.Code),
					}
					_ = cache.Put(rep)
					out <- rep
					continue
				}
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
						_ = refStore.Clear(t.Agent, t.Profile)
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
							_ = refStore.Record(t.Agent, t.Profile, code, desc)
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
		strings.Contains(lower, "status 401") ||
		strings.Contains(lower, "status: 401") ||
		strings.Contains(lower, "status code: 401") ||
		strings.Contains(lower, "status code 401") ||
		strings.Contains(lower, "http 401") ||
		strings.Contains(lower, "401 unauthorized")
}

