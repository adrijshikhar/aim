package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/presenter"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			executeWhoamiWithContextTo(cmd.Context(), cmd.OutOrStdout(), reg, pm)
			return nil
		},
	}
}

func executeWhoami(reg *agents.Registry, pm *profile.ProfileManager) {
	executeWhoamiWithContextTo(context.Background(), os.Stdout, reg, pm)
}

func executeWhoamiWithContextTo(ctx context.Context, out io.Writer, reg *agents.Registry, pm *profile.ProfileManager) {
	if ctx == nil {
		ctx = context.Background()
	}
	baseDir := config.BaseDir()
	profilesDir := filepath.Join(baseDir, "profiles")
	if pm != nil {
		if pm.BaseDir != "" {
			baseDir = pm.BaseDir
		}
		profilesDir = pm.ProfilesRoot()
	}
	profileName, agentName := resolveWhoamiTarget(profilesDir)

	cfg, _ := config.LoadConfig()
	if profileName == "" {
		renderWhoamiNoSession(out, pm, cfg)
		return
	}

	if agentName == "" && cfg != nil {
		if agentList := cfg.GetProfileAgents(profileName); len(agentList) > 0 {
			agentName = agentList[0]
		}
	}
	if agentName == "" {
		agentName = "agy" // default
	}

	renderWhoamiSession(ctx, out, reg, cfg, baseDir, profilesDir, profileName, agentName)
}

func resolveWhoamiTarget(profilesDir string) (profileName, agentName string) {
	profileName = os.Getenv("AIM_PROFILE")
	agentName = os.Getenv("AIM_AGENT")
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
	return profileName, agentName
}

func renderWhoamiNoSession(out io.Writer, pm *profile.ProfileManager, cfg *config.Config) {
	bold := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)
	muted := lipgloss.NewStyle().Foreground(tui.TextMuted)
	accent := lipgloss.NewStyle().Foreground(tui.AccentBlue).Bold(true)
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.TextDim).
		Padding(0, 1)
	if width, ok := presenter.IsTerminal(out); ok && width > 4 {
		cardStyle = cardStyle.MaxWidth(width)
	}

	var b strings.Builder
	b.WriteString(bold.Render("No active AIM session in this shell.") + "\n")
	writeConfiguredProfiles(&b, pm, cfg, accent, muted)
	b.WriteString("\n" + muted.Render("Run 'aim' or 'aim run <agent> <profile>' to start a session."))
	printCard(out, cardStyle.Render(strings.TrimRight(b.String(), "\n")))
}

func writeConfiguredProfiles(b *strings.Builder, pm *profile.ProfileManager, cfg *config.Config, accent, muted lipgloss.Style) {
	if pm == nil {
		return
	}
	profs, _ := pm.ListProfiles()
	if len(profs) == 0 {
		return
	}
	b.WriteString("\n" + muted.Render("Configured profiles:") + "\n")
	for _, p := range profs {
		agentTag := ""
		if cfg != nil {
			if agentList := cfg.GetProfileAgents(p); len(agentList) > 0 {
				agentTag = fmt.Sprintf(" [%s]", strings.Join(agentList, ", "))
			}
		}
		_, _ = fmt.Fprintf(b, "  • %s%s\n", accent.Render(p), muted.Render(agentTag))
	}
}

func renderWhoamiSession(ctx context.Context, out io.Writer, reg *agents.Registry, cfg *config.Config, baseDir, profilesDir, profileName, agentName string) {
	_ = cfg
	profileDir := filepath.Join(profilesDir, profileName)
	convID := findActiveConversationID(profileDir)
	chatTitle := getConversationTitle(ctx, profileDir, convID)
	quotaSummary := lookupWhoamiQuota(baseDir, reg, agentName, profileName, profileDir)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.AccentBlue).
		Padding(0, 1)
	if width, ok := presenter.IsTerminal(out); ok && width > 4 {
		cardStyle = cardStyle.MaxWidth(width)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.AccentBlue)
	labelStyle := lipgloss.NewStyle().Foreground(tui.TextMuted).Width(15)
	valStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.TextBright)

	var b strings.Builder
	b.WriteString(titleStyle.Render("⚡ Active AIM Session") + "\n\n")
	_, _ = fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Profile:"), valStyle.Render(profileName))
	_, _ = fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Agent:"), valStyle.Render(agentName))

	writeWhoamiAccount(&b, labelStyle, profileDir, agentName)
	_, _ = fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Profile Home:"), lipgloss.NewStyle().Foreground(tui.TextDim).Render(profileDir))
	writeWhoamiSessionDetails(&b, labelStyle, valStyle, convID, chatTitle, quotaSummary)

	printCard(out, cardStyle.Render(strings.TrimRight(b.String(), "\n")))
}

func lookupWhoamiQuota(baseDir string, reg *agents.Registry, agentName, profileName, profileDir string) string {
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	if rep, ok := cacheStore.Get(agentName, profileName); ok && rep.Summary != "" {
		return rep.Summary
	}
	if reg != nil {
		if adapter, err := reg.Get(agentName); err == nil && adapter.HasCredentials(profileDir) {
			return "[run 'aim usage' to inspect limits]"
		}
	}
	return ""
}

func writeWhoamiAccount(b *strings.Builder, labelStyle lipgloss.Style, profileDir, agentName string) {
	accInfo := profile.GetProfileAccountInfoForAgent(profileDir, agentName)
	if accInfo.Email == "" && accInfo.AuthMethod == "" {
		accInfo = profile.GetProfileAccountInfo(profileDir)
	}
	if accInfo.Email != "" {
		accStr := accInfo.Email
		if accInfo.Name != "" {
			accStr += " (" + accInfo.Name + ")"
		}
		_, _ = fmt.Fprintf(b, "%s %s\n", labelStyle.Render("Account:"), lipgloss.NewStyle().Foreground(tui.AccentCyan).Bold(true).Render(accStr))
	}
	if accInfo.AuthMethod != "" {
		_, _ = fmt.Fprintf(b, "%s %s\n", labelStyle.Render("Auth Method:"), lipgloss.NewStyle().Foreground(tui.TextMuted).Render(accInfo.AuthMethod))
	}
	if accInfo.ProjectID != "" {
		_, _ = fmt.Fprintf(b, "%s %s\n", labelStyle.Render("Project ID:"), lipgloss.NewStyle().Foreground(tui.TextDim).Render(accInfo.ProjectID))
	}
}

func writeWhoamiSessionDetails(b *strings.Builder, labelStyle, valStyle lipgloss.Style, convID, chatTitle, quotaSummary string) {
	if convID != "" {
		shortID := convID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		_, _ = fmt.Fprintf(b, "%s %s %s\n", labelStyle.Render("Session ID:"), valStyle.Render(shortID), lipgloss.NewStyle().Foreground(tui.TextMuted).Render(fmt.Sprintf("(%s)", convID)))
	}
	if chatTitle != "" {
		_, _ = fmt.Fprintf(b, "%s %s\n", labelStyle.Render("Chat Title:"), lipgloss.NewStyle().Foreground(tui.AccentCyan).Render(chatTitle))
	}
	if quotaSummary != "" {
		_, _ = fmt.Fprintf(b, "%s %s\n", labelStyle.Render("Quota:"), lipgloss.NewStyle().Foreground(tui.StatusGreen).Render(quotaSummary))
	}
}

func printCard(out io.Writer, card string) {
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, card)
	_, _ = fmt.Fprintln(out)
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

func findConversationSummariesDB(profileDir string) string {
	dbPaths := [2]string{
		filepath.Join(profileDir, ".gemini", "antigravity-cli", "conversation_summaries.db"),
		filepath.Join(config.RealHomeDir(), ".gemini", "antigravity-cli", "conversation_summaries.db"),
	}
	for _, db := range dbPaths {
		if fi, err := os.Stat(db); err == nil && fi.Size() > 0 {
			return db
		}
	}
	return ""
}

func getConversationTitle(ctx context.Context, profileDir, convID string) string {
	if !isValidConversationID(convID) {
		return ""
	}
	if ctx == nil {
		ctx = context.Background()
	}
	activeDB := findConversationSummariesDB(profileDir)
	if activeDB == "" {
		return ""
	}

	db, err := sql.Open("sqlite", activeDB)
	if err != nil {
		return ""
	}
	defer func() { _ = db.Close() }()
	_, _ = db.ExecContext(ctx, "PRAGMA query_only = ON;")

	var title, preview sql.NullString
	err = db.QueryRowContext(ctx, "SELECT title, preview FROM conversation_summaries WHERE conversation_id = ? LIMIT 1", convID).Scan(&title, &preview)
	if err != nil {
		return ""
	}
	if s := strings.TrimSpace(title.String); s != "" {
		return s
	}
	if p := strings.TrimSpace(preview.String); p != "" {
		line, _, _ := strings.Cut(p, "\n")
		return strings.TrimSpace(line)
	}
	return ""
}
