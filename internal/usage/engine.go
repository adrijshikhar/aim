package usage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
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

				if cache != nil && ctx.Err() == nil {
					// Stale-While-Revalidate: If fetch failed with StatusUnknown, but cache already contains
					// a valid, healthy report with quota windows, preserve the existing valid quota!
					if rep.Status == StatusUnknown {
						errLower := strings.ToLower(rep.Error)
						summaryLower := strings.ToLower(rep.Summary)
						combinedErr := errLower + " " + summaryLower

						isAuthError := strings.Contains(combinedErr, "unauthorized") ||
							strings.Contains(combinedErr, "invalid_grant") ||
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
					_ = cache.Put(*rep)
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
