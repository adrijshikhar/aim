package tui

import (
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/session"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type filterState struct {
	active bool
	input  textinput.Model
}

func newFilterState() filterState {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 64
	ti.Width = 30
	return filterState{
		active: false,
		input:  ti,
	}
}

// FilterInputValue returns the current text inside the filter input.
func (m Model) FilterInputValue() string {
	return m.filter.input.Value()
}

// FilteredProfiles returns the list of profiles matching the current filter query.
func (m Model) FilteredProfiles() []string {
	return m.filteredProfiles()
}

func (m Model) filteredProfiles() []string {
	query := strings.TrimSpace(m.filter.input.Value())
	query = strings.TrimPrefix(query, "/")
	if query == "" {
		return m.profiles
	}
	query = strings.ToLower(query)

	var res []string
	for _, p := range m.profiles {
		// 1. Check profile name
		if strings.Contains(strings.ToLower(p), query) {
			res = append(res, p)
			continue
		}

		// 2. Check attached agent names
		if m.cfg != nil {
			matchedAgent := false
			for _, ag := range m.cfg.GetProfileAgents(p) {
				if strings.Contains(strings.ToLower(ag), query) {
					matchedAgent = true
					break
				}
			}
			if matchedAgent {
				res = append(res, p)
			}
		}
	}
	return res
}

func (m Model) renderFilterBar() string {
	return fmt.Sprintf("  FILTER: %s  (Enter to apply, Esc to clear)\n", m.filter.input.View())
}

func (m Model) openFilter() (Model, tea.Cmd) {
	if m.filter.input.Prompt == "" && m.filter.input.CharLimit == 0 {
		m.filter = newFilterState()
	}
	m.filter.active = true
	cmd := m.filter.input.Focus()
	return m, cmd
}

func (m Model) updateFilter(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelStream()
			return m, tea.Quit
		case "enter":
			m.filter.active = false
			m.filter.input.Blur()
			filtered := m.filteredProfiles()
			if m.cursor >= len(filtered) {
				if len(filtered) > 0 {
					m.cursor = len(filtered) - 1
				} else {
					m.cursor = 0
				}
			}
			return m, nil
		case "esc":
			prevSelected := ""
			filtered := m.filteredProfiles()
			if m.cursor >= 0 && m.cursor < len(filtered) {
				prevSelected = filtered[m.cursor]
			}
			m.filter.active = false
			m.filter.input.Blur()
			m.filter.input.SetValue("")
			m.cursor = 0
			if prevSelected != "" {
				for idx, p := range m.profiles {
					if p == prevSelected {
						m.cursor = idx
						break
					}
				}
			} else if m.cursor >= len(m.profiles) {
				if len(m.profiles) > 0 {
					m.cursor = len(m.profiles) - 1
				}
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.filter.input, cmd = m.filter.input.Update(msg)
			filtered := m.filteredProfiles()
			if m.cursor >= len(filtered) {
				if len(filtered) > 0 {
					m.cursor = len(filtered) - 1
				} else {
					m.cursor = 0
				}
			}
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.filter.input, cmd = m.filter.input.Update(msg)
		return m, cmd
	}
}

// IndexedSession wraps a session.Session with a pre-computed lowercase search corpus.
type IndexedSession struct {
	session.Session
	searchCorpus string // Pre-computed lowercase corpus
}

// NewIndexedSession builds an IndexedSession with its lowercase search corpus pre-computed.
func NewIndexedSession(s session.Session) IndexedSession {
	return IndexedSession{
		Session: s,
		searchCorpus: strings.ToLower(strings.Join([]string{
			s.ID, s.ShortID, s.Title, s.Summary, s.Goal,
			s.Progress, s.Recent, s.Cwd, s.Profile, s.Agent,
		}, "\x00")),
	}
}

// SessionsIndex manages an in-memory pre-indexed collection of sessions for fast filtering.
type SessionsIndex struct {
	items []IndexedSession
}

// NewSessionsIndex creates a new SessionsIndex from a slice of sessions.
func NewSessionsIndex(sessions []session.Session) *SessionsIndex {
	items := make([]IndexedSession, len(sessions))
	for i, s := range sessions {
		items[i] = NewIndexedSession(s)
	}
	return &SessionsIndex{items: items}
}

// Search filters sessions by query and active status using the pre-computed corpus.
func (idx *SessionsIndex) Search(query string, activeOnly bool) []session.Session {
	return idx.SearchWithProfile(query, activeOnly, "")
}

// SearchWithProfile filters sessions by query, active status, and optional profile filter.
func (idx *SessionsIndex) SearchWithProfile(query string, activeOnly bool, profileFilter string) []session.Session {
	if idx == nil || len(idx.items) == 0 {
		return nil
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" && !activeOnly && profileFilter == "" {
		res := make([]session.Session, len(idx.items))
		for i, item := range idx.items {
			res[i] = item.Session
		}
		return res
	}

	res := make([]session.Session, 0, 32)
	for _, item := range idx.items {
		if activeOnly && item.Status != session.StatusActive {
			continue
		}
		if profileFilter != "" && !strings.EqualFold(item.Profile, profileFilter) {
			continue
		}
		if query == "" || strings.Contains(item.searchCorpus, query) {
			res = append(res, item.Session)
		}
	}
	return res
}

