package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	seen := make(map[string]bool)

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

	for _, p := range targetProviders {
		// 1. Scan configured profiles
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			profName := entry.Name()
			if filterProfile != "" && filterProfile != "host" && filterProfile != "<host>" && profName != filterProfile {
				continue
			}

			profDir := filepath.Join(profilesDir, profName)
			sessions, err := p.ListSessions(ctx, profDir, false)
			if err != nil {
				logger.Debug("[session/manager] ListSessions error on %s: %v", profDir, err)
				continue
			}

			for _, s := range sessions {
				key := fmt.Sprintf("%s:%s", s.Agent, s.ID)
				if !seen[key] {
					seen[key] = true
					s.Profile = profName
					s.IsHost = false
					if active, ok := activeProcesses[s.ID]; ok {
						s.Status = StatusActive
						s.PID = active.PID
						if active.Profile != "" {
							s.Profile = active.Profile
							s.IsHost = (active.Profile == "<host>")
						}
					} else {
						s.Status = StatusIdle
					}
					if filterProfile != "" && filterProfile != "host" && filterProfile != "<host>" && s.Profile != filterProfile {
						continue
					}
					if (filterProfile == "host" || filterProfile == "<host>") && !s.IsHost {
						continue
					}
					allSessions = append(allSessions, s)
				}
			}
		}

		// 2. Scan host storage (unless specifically filtering for a non-host profile)
		if filterProfile == "" || filterProfile == "host" || filterProfile == "<host>" {
			hostDir := config.RealHomeDir()
			hostSessions, err := p.ListSessions(ctx, hostDir, true)
			if err != nil {
				logger.Debug("[session/manager] ListSessions error on host %s: %v", hostDir, err)
			} else {
				for _, s := range hostSessions {
					key := fmt.Sprintf("%s:%s", s.Agent, s.ID)
					if !seen[key] {
						seen[key] = true
						s.Profile = "<host>"
						s.IsHost = true
						if active, ok := activeProcesses[s.ID]; ok {
							s.Status = StatusActive
							s.PID = active.PID
							if active.Profile != "" {
								s.Profile = active.Profile
								s.IsHost = (active.Profile == "<host>")
							}
						} else {
							s.Status = StatusIdle
						}
						if filterProfile != "" && filterProfile != "host" && filterProfile != "<host>" && s.Profile != filterProfile {
							continue
						}
						if (filterProfile == "host" || filterProfile == "<host>") && !s.IsHost {
							continue
						}
						allSessions = append(allSessions, s)
					}
				}
			}
		}
	}

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

// ResolveSession resolves a full ID or short ID prefix to a concrete Session.
func (m *Manager) ResolveSession(ctx context.Context, agent, idOrPrefix string) (*Session, error) {
	if idOrPrefix == "" {
		return nil, fmt.Errorf("session ID or prefix cannot be empty")
	}

	// Fast-path: query registered provider directly via GetSession if agent is specified
	if agent != "" {
		if p, ok := m.providers[agent]; ok {
			var matches []Session
			profilesDir := filepath.Join(config.BaseDir(), "profiles")
			entries, _ := os.ReadDir(profilesDir)
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				profDir := filepath.Join(profilesDir, entry.Name())
				if s, err := p.GetSession(ctx, idOrPrefix, profDir, false); err == nil && s != nil {
					s.Profile = entry.Name()
					s.IsHost = false
					matches = append(matches, *s)
				}
			}
			hostDir := config.RealHomeDir()
			if s, err := p.GetSession(ctx, idOrPrefix, hostDir, true); err == nil && s != nil {
				s.Profile = "<host>"
				s.IsHost = true
				matches = append(matches, *s)
			}

			if len(matches) > 0 {
				uniqueMap := make(map[string]Session)
				for _, match := range matches {
					uniqueMap[match.ID] = match
				}
				if len(uniqueMap) > 1 {
					var ids []string
					for id := range uniqueMap {
						ids = append(ids, id)
					}
					return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions: %s", idOrPrefix, strings.Join(ids, ", "))
				}
				for _, s := range uniqueMap {
					res := s
					if m.scanner != nil {
						if procs, err := m.scanner.ScanActiveProcesses(ctx); err == nil {
							if active, ok := procs[res.ID]; ok {
								res.Status = StatusActive
								res.PID = active.PID
							}
						}
					}
					return &res, nil
				}
			}
		}
	}

	// Fallback to broad scan via ListSessions
	sessions, err := m.ListSessions(ctx, agent, "", false)
	if err != nil {
		return nil, err
	}

	var matches []Session
	for _, s := range sessions {
		if s.ID == idOrPrefix || strings.HasPrefix(s.ID, idOrPrefix) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, idOrPrefix)
	}

	// Deduplicate matches with the exact same ID
	uniqueMap := make(map[string]Session)
	for _, m := range matches {
		uniqueMap[m.ID] = m
	}

	if len(uniqueMap) > 1 {
		var ids []string
		for id := range uniqueMap {
			ids = append(ids, id)
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches multiple sessions: %s", idOrPrefix, strings.Join(ids, ", "))
	}

	// Return the single matched session
	for _, s := range uniqueMap {
		res := s
		return &res, nil
	}

	return nil, ErrSessionNotFound
}
