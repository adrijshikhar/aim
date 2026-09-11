package usage

import (
	"context"
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
