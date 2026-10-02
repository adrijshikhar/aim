package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
)

var ErrSessionNotFound = errors.New("session not found")

// Manager coordinates session discovery, resolution, and lifecycle across profiles and host dotfiles.
type Manager struct {
	providers map[string]SessionProvider
	scanner   ProcessScanner
}

// NewManager creates a Manager with the default process scanner.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]SessionProvider),
		scanner:   NewDefaultProcessScanner(),
	}
}

// NewManagerWithScanner creates a Manager with a custom ProcessScanner.
func NewManagerWithScanner(scanner ProcessScanner) *Manager {
	return &Manager{
		providers: make(map[string]SessionProvider),
		scanner:   scanner,
	}
}

// RegisterProvider registers a SessionProvider for an agent.
func (m *Manager) RegisterProvider(p SessionProvider) {
	m.providers[p.Agent()] = p
}

// Provider retrieves the SessionProvider registered for an agent.
func (m *Manager) Provider(agent string) SessionProvider {
	return m.providers[agent]
}

// ListSessions discovers sessions across profiles and host, optionally filtered by agent or profile.
func (m *Manager) ListSessions(ctx context.Context, filterAgent, filterProfile string, activeOnly bool) ([]Session, error) {
	activeProcesses := make(map[string]ActiveProcessInfo)
	if m.scanner != nil {
		if procs, err := m.scanner.ScanActiveProcesses(ctx); err == nil {
			activeProcesses = procs
		}
	}

	var allSessions []Session
	seen := make(map[string]int)
	hostOnly := filterProfile == "host" || filterProfile == "<host>"
	appendSession := func(s Session, profile string, isHost bool) {
		s.Profile, s.IsHost = profile, isHost
		s.Status = StatusIdle
		if active, ok := activeProcesses[s.ID]; ok {
			s.Status, s.PID = StatusActive, active.PID
			if active.Profile != "" {
				s.Profile = active.Profile
				s.IsHost = active.Profile == "<host>"
			}
		}
		if hostOnly && !s.IsHost {
			return
		}
		if filterProfile != "" && !hostOnly && s.Profile != filterProfile {
			return
		}

		key := fmt.Sprintf("%s:%s", s.Agent, s.ID)
		if idx, exists := seen[key]; exists {
			// When listing across profiles, prioritize the most recently active copy
			if filterProfile == "" {
				existing := allSessions[idx]
				if s.LastActiveAt.After(existing.LastActiveAt) || (s.LastActiveAt.Equal(existing.LastActiveAt) && existing.IsHost && !s.IsHost) {
					allSessions[idx] = s
				}
			}
			return
		}
		seen[key] = len(allSessions)
		allSessions = append(allSessions, s)
	}

	// Collect providers to query
	var targetProviders []SessionProvider
	if filterAgent != "" {
		if p, ok := m.providers[filterAgent]; ok {
			targetProviders = append(targetProviders, p)
		}
	} else {
		for _, p := range m.providers {
			targetProviders = append(targetProviders, p)
		}
	}

	profilesDir := filepath.Join(config.BaseDir(), "profiles")
	entries, _ := os.ReadDir(profilesDir)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, p := range targetProviders {
		// 1. Scan configured profiles
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			profName := entry.Name()
			if filterProfile != "" && !hostOnly && profName != filterProfile {
				continue
			}

			profDir := filepath.Join(profilesDir, profName)
			provider := p
			profile := profName
			wg.Add(1)
			go func() {
				defer wg.Done()
				sessions, err := provider.ListSessions(ctx, profDir, false)
				if err != nil {
					logger.Debug("[session/manager] ListSessions error on %s: %v", profDir, err)
					return
				}

				mu.Lock()
				for _, s := range sessions {
					appendSession(s, profile, false)
				}
				mu.Unlock()
			}()
		}

		// 2. Scan host storage (unless specifically filtering for a non-host profile)
		if filterProfile == "" || hostOnly {
			hostDir := config.RealHomeDir()
			provider := p
			wg.Add(1)
			go func() {
				defer wg.Done()
				hostSessions, err := provider.ListSessions(ctx, hostDir, true)
				if err != nil {
					logger.Debug("[session/manager] ListSessions error on host %s: %v", hostDir, err)
					return
				}

				mu.Lock()
				for _, s := range hostSessions {
					appendSession(s, "<host>", true)
				}
				mu.Unlock()
			}()
		}
	}

	wg.Wait()

	// Filter activeOnly if requested
	if activeOnly {
		var filtered []Session
		for _, s := range allSessions {
			if s.Status == StatusActive {
				filtered = append(filtered, s)
			}
		}
		allSessions = filtered
	}

	// Sort chronologically descending
	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].LastActiveAt.After(allSessions[j].LastActiveAt)
	})

	return allSessions, nil
}

// FindAllSessionsByID returns all matching sessions across all profiles and host for the given ID or prefix.
func (m *Manager) FindAllSessionsByID(ctx context.Context, agent, idOrPrefix string) ([]Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if idOrPrefix == "" {
		return nil, fmt.Errorf("session ID or prefix cannot be empty")
	}

	var targetProviders []SessionProvider
	if agent != "" {
		if p, ok := m.providers[agent]; ok {
			targetProviders = append(targetProviders, p)
		}
	} else {
		for _, p := range m.providers {
			targetProviders = append(targetProviders, p)
		}
	}

	if len(targetProviders) > 0 {
		var matches []Session
		profilesDir := filepath.Join(config.BaseDir(), "profiles")
		entries, _ := os.ReadDir(profilesDir)
		hostDir := config.RealHomeDir()

		for _, p := range targetProviders {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				profDir := filepath.Join(profilesDir, entry.Name())
				s, err := p.GetSession(ctx, idOrPrefix, profDir, false)
				if err != nil {
					logger.Debug("[session] error querying session %q in profile %q: %v", idOrPrefix, entry.Name(), err)
					if strings.Contains(err.Error(), "ambiguous") {
						return nil, err
					}
				} else if s != nil {
					s.Profile = entry.Name()
					s.IsHost = false
					matches = append(matches, *s)
				}
			}
			s, err := p.GetSession(ctx, idOrPrefix, hostDir, true)
			if err != nil {
				logger.Debug("[session] error querying session %q in host: %v", idOrPrefix, err)
				if strings.Contains(err.Error(), "ambiguous") {
					return nil, err
				}
			} else if s != nil {
				s.Profile = "<host>"
				s.IsHost = true
				matches = append(matches, *s)
			}
		}
		if len(matches) > 0 {
			return matches, nil
		}
	}

	// Fallback to broad scan via ListSessions
	sessions, err := m.ListSessions(ctx, agent, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var matches []Session
	for _, s := range sessions {
		if s.ID == idOrPrefix || strings.HasPrefix(s.ID, idOrPrefix) {
			matches = append(matches, s)
		}
	}
	if len(matches) == 0 {
		lowerTarget := strings.ToLower(strings.TrimSpace(idOrPrefix))
		if lowerTarget != "" {
			for _, s := range sessions {
				if strings.ToLower(strings.TrimSpace(s.Title)) == lowerTarget {
					matches = append(matches, s)
				}
			}
			if len(matches) > 0 {
				sort.Slice(matches, func(i, j int) bool {
					return matches[i].LastActiveAt.After(matches[j].LastActiveAt)
				})
				return []Session{matches[0]}, nil
			}
			for _, s := range sessions {
				if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.Title)), lowerTarget) {
					matches = append(matches, s)
				}
			}
			if len(matches) > 0 {
				sort.Slice(matches, func(i, j int) bool {
					return matches[i].LastActiveAt.After(matches[j].LastActiveAt)
				})
				return []Session{matches[0]}, nil
			}
		}
	}
	return matches, nil
}

// ResolveSession resolves a full ID or short ID prefix to a concrete Session.
func (m *Manager) ResolveSession(ctx context.Context, agent, idOrPrefix string) (*Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	res, err := m.LatestSession(ctx, agent, idOrPrefix)
	if err != nil {
		return nil, err
	}
	if m.scanner != nil {
		if procs, err := m.scanner.ScanActiveProcesses(ctx); err == nil {
			if active, ok := procs[res.ID]; ok {
				res.Status = StatusActive
				res.PID = active.PID
			}
		}
	}
	return res, nil
}

// LatestSession resolves a full ID or prefix without additional active-process
// enrichment. It is appropriate when callers need the most recently active stored snapshot.
func (m *Manager) LatestSession(ctx context.Context, agent, idOrPrefix string) (*Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if idOrPrefix == "" {
		return nil, fmt.Errorf("session ID or prefix cannot be empty")
	}

	matches, err := m.FindAllSessionsByID(ctx, agent, idOrPrefix)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, idOrPrefix)
	}

	uniqueMatches := DeduplicateMatches(matches)

	if len(uniqueMatches) > 1 {
		var ids []string
		for _, s := range uniqueMatches {
			ids = append(ids, s.ShortID)
		}
		sort.Strings(ids)
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions: %s", idOrPrefix, strings.Join(ids, ", "))
	}

	res := uniqueMatches[0]
	return &res, nil
}

// DeduplicateMatches collapses multiple instances of the same session ID across profiles,
// keeping the most recently active snapshot. If timestamps are equal, isolated profiles take
// precedence over host.
func DeduplicateMatches(matches []Session) []Session {
	idToSession := make(map[string]Session)
	for _, match := range matches {
		existing, exists := idToSession[match.ID]
		if !exists {
			idToSession[match.ID] = match
			continue
		}
		// Check storage sizes to avoid picking a truncated session merely because it was touched recently
		var matchSize, existSize int64
		if match.StoragePath != "" {
			if fi, err := os.Stat(match.StoragePath); err == nil {
				matchSize = fi.Size()
			}
		}
		if existing.StoragePath != "" {
			if fi, err := os.Stat(existing.StoragePath); err == nil {
				existSize = fi.Size()
			}
		}
		if matchSize > 0 && existSize > 0 {
			if matchSize > existSize+256 {
				idToSession[match.ID] = match
				continue
			} else if existSize > matchSize+256 {
				continue
			}
		}

		// If match is more recently active, take it
		if match.LastActiveAt.After(existing.LastActiveAt) {
			idToSession[match.ID] = match
			continue
		}
		// If timestamps are equal, prioritize isolated profile over host
		if match.LastActiveAt.Equal(existing.LastActiveAt) {
			if existing.IsHost && !match.IsHost {
				idToSession[match.ID] = match
			}
		}
	}

	var result []Session
	for _, s := range idToSession {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LastActiveAt.Equal(result[j].LastActiveAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].LastActiveAt.After(result[j].LastActiveAt)
	})
	return result
}

// ResolveCwd resolves the working directory for a session by delegating to its provider.
func (m *Manager) ResolveCwd(ctx context.Context, s *Session) (string, error) {
	if s == nil {
		return "", nil
	}
	if p, ok := m.providers[s.Agent]; ok {
		return p.ResolveCwd(ctx, s)
	}
	return s.Cwd, nil
}

// ResolveSummary resolves the structured summary for a session by delegating to its provider.
func (m *Manager) ResolveSummary(ctx context.Context, s *Session) (SessionSummary, error) {
	if s == nil {
		return SessionSummary{}, nil
	}
	if p, ok := m.providers[s.Agent]; ok {
		return p.ResolveSummary(ctx, s)
	}
	return SessionSummary{Goal: s.Summary, Raw: s.Summary}, nil
}

