package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newWhoamiCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:     "whoami",
		Aliases: []string{"current"},
		Short:   "Show active profile, agent, session, and quota info",
		Run: func(cmd *cobra.Command, args []string) {
			executeWhoami(reg, pm)
		},
	}
}

func executeWhoami(reg *agents.Registry, pm *profile.ProfileManager) {
	profileName := os.Getenv("AIM_PROFILE")
	agentName := os.Getenv("AIM_AGENT")

	baseDir := config.BaseDir()
	profilesDir := filepath.Join(baseDir, "profiles")
	homeDir := os.Getenv("HOME")

	// If AIM_PROFILE is not set, try to infer from HOME directory
	if profileName == "" && strings.HasPrefix(homeDir, profilesDir) {
		rel, err := filepath.Rel(profilesDir, homeDir)
		if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 0 && parts[0] != "" {
				profileName = parts[0]
			}
		}
	}

	if profileName != "" {
		if strings.ContainsAny(profileName, "/\\") || filepath.Base(profileName) != profileName || profileName == "." || profileName == ".." {
			profileName = ""
		}
	}

	logger.Debug("[whoami] Inferred profileName=%q, agentName=%q (HOME=%s, profilesDir=%s)", profileName, agentName, homeDir, profilesDir)

	cfg, _ := config.LoadConfig()

	// If still no active profile, show summary of available profiles
	if profileName == "" {
		bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
		muted := lipgloss.NewStyle().Foreground(tui.TextMuted)
		accent := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tui.TextDim).
			Padding(0, 1)

		var b strings.Builder
		b.WriteString(bold.Render("No active AIM session in this shell.") + "\n")
		if pm != nil {
			profs, _ := pm.ListProfiles()
			if len(profs) > 0 {
				b.WriteString("\n" + muted.Render("Configured profiles:") + "\n")
				for _, p := range profs {
					var agentList []string
					if cfg != nil {
						agentList = cfg.GetProfileAgents(p)
					}
					agentTag := ""
					if len(agentList) > 0 {
						agentTag = fmt.Sprintf(" [%s]", strings.Join(agentList, ", "))
					}
					b.WriteString(fmt.Sprintf("  • %s%s\n", accent.Render(p), muted.Render(agentTag)))
				}
			}
		}
		b.WriteString("\n" + muted.Render("Run 'aim' or 'aim run <agent> <profile>' to start a session."))
		fmt.Println()
		fmt.Println(cardStyle.Render(strings.TrimRight(b.String(), "\n")))
		fmt.Println()
		return
	}

	// Active session detected
	if agentName == "" && cfg != nil {
		agentList := cfg.GetProfileAgents(profileName)
		if len(agentList) > 0 {
			agentName = agentList[0]
		}
	}
	if agentName == "" {
		agentName = "agy" // default
	}

	profileDir := filepath.Join(profilesDir, profileName)
	convID := findActiveConversationID(profileDir)
	chatTitle := getConversationTitle(profileDir, convID)

	quotaSummary := ""
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	if rep, ok := cacheStore.Get(agentName, profileName); ok && rep.Summary != "" {
		quotaSummary = rep.Summary
	} else if reg != nil {
		if adapter, err := reg.Get(agentName); err == nil && adapter.HasCredentials(profileDir) {
			quotaSummary = "[run 'aim usage' to inspect limits]"
		}
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.AccentBlue).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
	labelStyle := lipgloss.NewStyle().Foreground(tui.TextMuted).Width(15)
	valStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	var b strings.Builder
	b.WriteString(titleStyle.Render("⚡ Active AIM Session") + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Profile:"), valStyle.Render(profileName)))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Agent:"), valStyle.Render(agentName)))

	accInfo := profile.GetProfileAccountInfo(profileDir)
	if accInfo.Email != "" {
		accStr := accInfo.Email
		if accInfo.Name != "" {
			accStr += " (" + accInfo.Name + ")"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Account:"), lipgloss.NewStyle().Foreground(tui.AccentCyan).Bold(true).Render(accStr)))
	}
	if accInfo.AuthMethod != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Auth Method:"), lipgloss.NewStyle().Foreground(tui.TextMuted).Render(accInfo.AuthMethod)))
	}
	if accInfo.ProjectID != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Project ID:"), lipgloss.NewStyle().Foreground(tui.TextDim).Render(accInfo.ProjectID)))
	}

	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Profile Home:"), lipgloss.NewStyle().Foreground(tui.TextDim).Render(profileDir)))

	if convID != "" {
		shortID := convID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		b.WriteString(fmt.Sprintf("%s %s %s\n", labelStyle.Render("Session ID:"), valStyle.Render(shortID), lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("(%s)", convID))))
	}

	if chatTitle != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Chat Title:"), lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(chatTitle)))
	}

	if quotaSummary != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Quota:"), lipgloss.NewStyle().Foreground(tui.StatusGreen).Render(quotaSummary)))
	}

	fmt.Println()
	fmt.Println(cardStyle.Render(strings.TrimRight(b.String(), "\n")))
	fmt.Println()
}

func findActiveConversationID(profileDir string) string {
	if meta := os.Getenv("ANTIGRAVITY_SOURCE_METADATA"); meta != "" {
		var m struct {
			Tool struct {
				ConversationID string `json:"conversationId"`
			} `json:"tool"`
		}
		if err := json.Unmarshal([]byte(meta), &m); err == nil && m.Tool.ConversationID != "" {
			return m.Tool.ConversationID
		}
	}

	presenceDir := filepath.Join(profileDir, ".gemini", "antigravity-cli", "presence")
	entries, err := os.ReadDir(presenceDir)
	if err == nil {
		var newestName string
		var newestTime time.Time
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".lock") {
				info, err := e.Info()
				if err == nil && info.ModTime().After(newestTime) {
					newestTime = info.ModTime()
					newestName = strings.TrimSuffix(e.Name(), ".lock")
				}
			}
		}
		if newestName != "" {
			return newestName
		}
	}
	return ""
}

func isValidConversationID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

var (
	sqliteBinOnce   sync.Once
	cachedSqliteBin string
)

func getSqliteBin() string {
	sqliteBinOnce.Do(func() {
		if bin, err := exec.LookPath("sqlite3"); err == nil {
			cachedSqliteBin = bin
		}
	})
	return cachedSqliteBin
}

func getConversationTitle(profileDir, convID string) string {
	if !isValidConversationID(convID) {
		return ""
	}
	dbPaths := [2]string{
		filepath.Join(profileDir, ".gemini", "antigravity-cli", "conversation_summaries.db"),
		filepath.Join(config.RealHomeDir(), ".gemini", "antigravity-cli", "conversation_summaries.db"),
	}

	var activeDB string
	for _, db := range dbPaths {
		if fi, err := os.Stat(db); err == nil && fi.Size() > 0 {
			activeDB = db
			break
		}
	}
	if activeDB == "" {
		return ""
	}

	sqliteBin := getSqliteBin()
	if sqliteBin == "" {
		return ""
	}

	cmd := exec.Command(sqliteBin, activeDB, fmt.Sprintf("SELECT title, preview FROM conversation_summaries WHERE conversation_id = '%s' LIMIT 1;", convID))
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		parts := strings.Split(strings.TrimSpace(string(out)), "|")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			preview := strings.TrimSpace(parts[1])
			lines := strings.Split(preview, "\n")
			return strings.TrimSpace(lines[0])
		}
	}
	return ""
}
