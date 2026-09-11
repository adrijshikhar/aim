package usage

import (
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusOK        Status = "OK"
	StatusWarning   Status = "WARN"
	StatusCritical  Status = "CRITICAL"
	StatusExhausted Status = "EXHAUSTED"
	StatusUnknown   Status = "UNKNOWN"
)

type LimitWindow struct {
	Category     string        `json:"category"`
	Name         string        `json:"name"`
	RemainingPct int           `json:"remaining_pct"`
	ResetsAt     time.Time     `json:"resets_at"`
	ResetsIn     time.Duration `json:"resets_in"`
}

type Report struct {
	Agent     string        `json:"agent"`
	Profile   string        `json:"profile"`
	Status    Status        `json:"status"`
	Windows   []LimitWindow `json:"windows"`
	Credits   string        `json:"credits,omitempty"`
	Summary   string        `json:"summary"`
	FetchedAt time.Time     `json:"fetched_at"`
	FromCache bool          `json:"from_cache"`
	Error     string        `json:"error,omitempty"`
}

func (r *Report) PrimaryWindow() *LimitWindow {
	if r == nil {
		return nil
	}
	for i := range r.Windows {
		name := strings.ToLower(r.Windows[i].Name)
		if strings.Contains(name, "five") || strings.Contains(name, "5h") || strings.Contains(name, "5 hour") {
			return &r.Windows[i]
		}
	}
	if len(r.Windows) > 0 {
		return &r.Windows[0]
	}
	return nil
}

func (r *Report) WeeklyWindow() *LimitWindow {
	if r == nil {
		return nil
	}
	for i := range r.Windows {
		name := strings.ToLower(r.Windows[i].Name)
		if strings.Contains(name, "week") || strings.Contains(name, "7d") {
			return &r.Windows[i]
		}
	}
	return nil
}

func CalculateStatus(windows []LimitWindow) Status {
	if len(windows) == 0 {
		return StatusUnknown
	}
	minPct := 100
	for _, w := range windows {
		if w.RemainingPct < minPct {
			minPct = w.RemainingPct
		}
	}
	if minPct <= 0 {
		return StatusExhausted
	}
	if minPct < 15 {
		return StatusCritical
	}
	if minPct <= 50 {
		return StatusWarning
	}
	return StatusOK
}

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, hours)
}

func FormatWindowSummary(w *LimitWindow) string {
	if w == nil {
		return ""
	}
	label := "limit"
	lower := strings.ToLower(w.Name)
	if strings.Contains(lower, "five") || strings.Contains(lower, "5h") || strings.Contains(lower, "5 hour") {
		label = "5h"
	} else if strings.Contains(lower, "week") || strings.Contains(lower, "7d") {
		label = "wk"
	}

	pct := w.RemainingPct
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}

	if pct >= 100 {
		return fmt.Sprintf("%s: 100%%", label)
	}
	if w.ResetsIn > 0 {
		return fmt.Sprintf("%s: %d%% [%s]", label, pct, FormatDuration(w.ResetsIn))
	}
	return fmt.Sprintf("%s: %d%%", label, pct)
}

func RenderBar(remainingPct int, width int) string {
	if width <= 0 {
		return ""
	}
	if remainingPct < 0 {
		remainingPct = 0
	}
	if remainingPct > 100 {
		remainingPct = 100
	}
	filled := (remainingPct * width) / 100
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

type ModelGroup struct {
	Category string        `json:"category"`
	Windows  []LimitWindow `json:"windows"`
}

func CleanModelCategory(cat string) string {
	c := strings.TrimSpace(cat)
	if c == "" {
		return "—"
	}
	lower := strings.ToLower(c)
	if strings.Contains(lower, "claude") || strings.Contains(lower, "gpt") {
		return "Claude & GPT"
	}
	if strings.Contains(lower, "gemini") {
		return "Gemini"
	}
	return c
}

func (r *Report) ModelGroups() []ModelGroup {
	if r == nil || len(r.Windows) == 0 {
		return nil
	}

	hasCategory := false
	for _, w := range r.Windows {
		if strings.TrimSpace(w.Category) != "" {
			hasCategory = true
			break
		}
	}
	if !hasCategory {
		return nil
	}

	var groups []ModelGroup
	catIdx := make(map[string]int)

	for _, w := range r.Windows {
		cleaned := CleanModelCategory(w.Category)
		idx, exists := catIdx[cleaned]
		if !exists {
			catIdx[cleaned] = len(groups)
			groups = append(groups, ModelGroup{
				Category: cleaned,
				Windows:  []LimitWindow{w},
			})
		} else {
			groups[idx].Windows = append(groups[idx].Windows, w)
		}
	}
	return groups
}
