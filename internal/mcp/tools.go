package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/oauth"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/aim-cli/aim/internal/usage"
)

func (s *Server) handleToolCall(req JSONRPCRequest) JSONRPCResponse {
	var params ToolCallParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: -32602, Message: fmt.Sprintf("Invalid params: %v", err)},
			}
		}
	}

	switch params.Name {
	case "aim_whoami":
		return s.callWhoami(req.ID)
	case "aim_get_quota":
		return s.callGetQuota(req.ID, params.Arguments)
	case "aim_list_profiles":
		return s.callListProfiles(req.ID)
	case "aim_get_sessions":
		return s.callGetSessions(req.ID, params.Arguments)
	case "aim_get_session_details":
		return s.callGetSessionDetails(req.ID, params.Arguments)
	case "aim_get_active_binding":
		return s.callGetActiveBinding(req.ID, params.Arguments)
	case "aim_diagnostics":
		return s.callDiagnostics(req.ID)
	case "aim_daemon_status":
		return s.callDaemonStatus(req.ID)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: fmt.Sprintf("Tool not found: %s", params.Name)},
		}
	}
}

func (s *Server) resolveActiveContext() (string, string) {
	profileName := os.Getenv("AIM_PROFILE")
	agentName := os.Getenv("AIM_AGENT")

	baseDir := s.baseDir
	if baseDir == "" {
		baseDir = config.BaseDir()
	}
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

	cfg, _ := config.LoadConfig()
	if profileName != "" && agentName == "" && cfg != nil {
		agentList := cfg.GetProfileAgents(profileName)
		if len(agentList) > 0 {
			agentName = agentList[0]
		}
	}
	if agentName == "" {
		agentName = "agy"
	}

	return profileName, agentName
}

func (s *Server) callWhoami(id any) JSONRPCResponse {
	if s.pm == nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "Profile manager is not initialized."},
				},
				IsError: true,
			},
		}
	}

	profileName, agentName := s.resolveActiveContext()

	baseDir := s.baseDir
	if baseDir == "" {
		baseDir = config.BaseDir()
	}
	profilesDir := filepath.Join(baseDir, "profiles")

	if profileName == "" {
		profs, _ := s.pm.ListProfiles()
		res := map[string]any{
			"status":             "unattached",
			"message":            "No active AIM session detected in environment.",
			"available_profiles": profs,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: string(data)},
				},
			},
		}
	}

	profileDir := filepath.Join(profilesDir, profileName)
	accInfo := profile.GetProfileAccountInfoForAgent(profileDir, agentName)

	quotaSummary := ""
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)
	if rep, ok := cacheStore.Get(agentName, profileName); ok && rep.Summary != "" {
		quotaSummary = rep.Summary
	} else if rep, ok := cacheStore.GetStale(agentName, profileName); ok && rep.Summary != "" {
		quotaSummary = rep.Summary
	} else if s.reg != nil {
		if adapter, err := s.reg.Get(agentName); err == nil && adapter.HasCredentials(profileDir) {
			quotaSummary = "[run 'aim usage' to inspect limits]"
		}
	}

	payload := map[string]any{
		"profile":       profileName,
		"agent":         agentName,
		"account_email": accInfo.Email,
		"account_name":  accInfo.Name,
		"auth_method":   accInfo.AuthMethod,
		"project_id":    accInfo.ProjectID,
		"profile_dir":   profileDir,
		"quota_summary": quotaSummary,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Failed to marshal identity: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callGetQuota(id any, argsRaw json.RawMessage) JSONRPCResponse {
	var args struct {
		Agent   string `json:"agent"`
		Profile string `json:"profile"`
	}
	if len(argsRaw) > 0 {
		if err := json.Unmarshal(argsRaw, &args); err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      id,
				Error:   &Error{Code: -32602, Message: fmt.Sprintf("Invalid arguments: %v", err)},
			}
		}
	}

	activeProfile, activeAgent := s.resolveActiveContext()
	targetProfile := strings.TrimSpace(args.Profile)
	if targetProfile == "" {
		targetProfile = activeProfile
	}
	targetAgent := strings.ToLower(strings.TrimSpace(args.Agent))
	if targetAgent == "" {
		targetAgent = strings.ToLower(activeAgent)
	}

	if targetProfile == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "No profile specified and no active AIM profile detected in environment."},
				},
				IsError: true,
			},
		}
	}

	baseDir := s.baseDir
	if baseDir == "" {
		baseDir = config.BaseDir()
	}
	cacheStore := usage.NewCacheStore(baseDir, usage.DefaultTTL)

	rep, ok := cacheStore.Get(targetAgent, targetProfile)
	if !ok {
		rep, ok = cacheStore.GetStale(targetAgent, targetProfile)
	}

	if !ok {
		msg := fmt.Sprintf("No cached quota report available for agent %q in profile %q. Run 'aim usage' to fetch live limits.", targetAgent, targetProfile)
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: msg},
				},
			},
		}
	}

	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Failed to marshal quota report: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callListProfiles(id any) JSONRPCResponse {
	if s.pm == nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "Profile manager is not initialized."},
				},
				IsError: true,
			},
		}
	}

	profs, err := s.pm.ListProfiles()
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to list profiles: %v", err)},
				},
				IsError: true,
			},
		}
	}

	cfg, _ := config.LoadConfig()
	type profileDetail struct {
		Name   string   `json:"name"`
		Agents []string `json:"agents,omitempty"`
	}

	details := make([]profileDetail, 0, len(profs))
	for _, p := range profs {
		var agList []string
		if cfg != nil {
			agList = cfg.GetProfileAgents(p)
		}
		details = append(details, profileDetail{
			Name:   p,
			Agents: agList,
		})
	}

	res := map[string]any{
		"profiles": profs,
		"details":  details,
	}

	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Failed to marshal profiles: %v", err)}},
				IsError: true,
			},
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callGetActiveBinding(id any, args json.RawMessage) JSONRPCResponse {
	targetCwd := ""
	if len(args) > 0 {
		var p struct {
			Cwd string `json:"cwd"`
		}
		_ = json.Unmarshal(args, &p)
		targetCwd = strings.TrimSpace(p.Cwd)
	}
	if targetCwd == "" {
		if wd, err := os.Getwd(); err == nil {
			targetCwd = wd
		}
	}

	res, err := profile.ResolveDirProfile(targetCwd)
	isBound := err == nil && res != nil && res.Profile != ""

	payload := map[string]any{
		"cwd":          targetCwd,
		"is_bound":     isBound,
		"profile":      "",
		"binding_file": "",
	}
	if isBound {
		payload["profile"] = res.Profile
		payload["binding_file"] = res.SourceFile
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callDaemonStatus(id any) JSONRPCResponse {
	homeDir := config.RealHomeDir()
	info, _ := daemon.Status(homeDir, daemon.DefaultServiceLabel)

	payload := map[string]any{
		"installed":        info.Installed,
		"active":           info.Active,
		"service_label":    info.Label,
		"config_path":      info.ConfigPath,
		"binary_path":      info.BinaryPath,
		"last_run":         info.LastRun,
		"last_run_message": info.LastRunMessage,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callDiagnostics(id any) JSONRPCResponse {
	baseDir := s.baseDir
	if baseDir == "" {
		baseDir = config.BaseDir()
	}

	results := make([]agents.DiagnosticResult, 0)
	overallStatus := "OK"

	// Check Refusal Store
	refStore := oauth.NewRefusalStore(baseDir)
	if s.pm != nil {
		profs, _ := s.pm.ListProfiles()
		for _, p := range profs {
			activeAgent := "claude"
			if s.reg != nil {
				for _, adapter := range s.reg.All() {
					if _, refused := refStore.IsRefused(adapter.Name(), p); refused {
						overallStatus = "FAIL"
						results = append(results, agents.DiagnosticResult{
							Category: "Auth",
							Status:   "FAIL",
							Message:  fmt.Sprintf("Profile %q (%s) has an active re-login refusal", p, adapter.Name()),
						})
					}
				}
			} else if _, refused := refStore.IsRefused(activeAgent, p); refused {
				overallStatus = "FAIL"
				results = append(results, agents.DiagnosticResult{
					Category: "Auth",
					Status:   "FAIL",
					Message:  fmt.Sprintf("Profile %q (%s) has an active re-login refusal", p, activeAgent),
				})
			}
		}
	}

	// Check Daemon
	homeDir := config.RealHomeDir()
	dInfo, _ := daemon.Status(homeDir, daemon.DefaultServiceLabel)
	if dInfo.Installed {
		results = append(results, agents.DiagnosticResult{
			Category: "Daemon",
			Status:   "OK",
			Message:  fmt.Sprintf("Background daemon %s is configured", dInfo.Label),
		})
	}

	// Default passing check if empty
	if len(results) == 0 {
		results = append(results, agents.DiagnosticResult{
			Category: "System",
			Status:   "OK",
			Message:  "All profile and authentication checks are normal",
		})
	}

	payload := map[string]any{
		"status":  overallStatus,
		"results": results,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callGetSessions(id any, args json.RawMessage) JSONRPCResponse {
	limit := 10
	targetCwd := ""
	targetProfile := ""

	if len(args) > 0 {
		var p struct {
			Limit   int    `json:"limit"`
			Cwd     string `json:"cwd"`
			Profile string `json:"profile"`
		}
		if err := json.Unmarshal(args, &p); err == nil {
			if p.Limit > 0 {
				limit = p.Limit
			}
			targetCwd = strings.TrimSpace(p.Cwd)
			targetProfile = strings.TrimSpace(p.Profile)
		}
	}

	type sessionItem struct {
		ID        string `json:"id"`
		ShortID   string `json:"short_id"`
		Agent     string `json:"agent"`
		Profile   string `json:"profile"`
		Cwd       string `json:"cwd"`
		Goal      string `json:"goal"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at"`
	}

	items := make([]sessionItem, 0)
	if s.sessionMgr != nil {
		all, err := s.sessionMgr.ListSessions(context.Background(), "", targetProfile, false)
		if err == nil {
			var filtered []session.Session
			normTargetCwd := filepath.ToSlash(strings.TrimRight(targetCwd, "/\\"))

			for _, sess := range all {
				if targetCwd != "" {
					sessCwd := filepath.ToSlash(strings.TrimRight(sess.Cwd, "/\\"))
					if !strings.EqualFold(sessCwd, normTargetCwd) {
						continue
					}
				}
				filtered = append(filtered, sess)
			}

			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].LastActiveAt.After(filtered[j].LastActiveAt)
			})

			if len(filtered) > limit {
				filtered = filtered[:limit]
			}

			for _, sess := range filtered {
				goal := sess.Goal
				if goal == "" {
					goal = sess.Summary
				}
				if goal == "" {
					goal = sess.Title
				}
				items = append(items, sessionItem{
					ID:        sess.ID,
					ShortID:   sess.ShortID,
					Agent:     sess.Agent,
					Profile:   sess.Profile,
					Cwd:       sess.Cwd,
					Goal:      goal,
					Status:    string(sess.Status),
					UpdatedAt: sess.LastActiveAt.Format(time.RFC3339),
				})
			}
		}
	}

	payload := map[string]any{
		"sessions": items,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

func (s *Server) callGetSessionDetails(id any, args json.RawMessage) JSONRPCResponse {
	var p struct {
		SessionID string `json:"session_id"`
	}
	if len(args) == 0 {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: "Missing required argument 'session_id'"}},
				IsError: true,
			},
		}
	}
	if err := json.Unmarshal(args, &p); err != nil || strings.TrimSpace(p.SessionID) == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: "Invalid or empty 'session_id'"}},
				IsError: true,
			},
		}
	}

	if s.sessionMgr == nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: "Session manager is not configured"}},
				IsError: true,
			},
		}
	}

	targetID := strings.TrimSpace(p.SessionID)
	sess, err := s.sessionMgr.ResolveSession(context.Background(), "", targetID)
	if err != nil || sess == nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Session not found: %s", targetID)}},
				IsError: true,
			},
		}
	}

	goal := sess.Goal
	if goal == "" {
		goal = sess.Summary
	}
	if goal == "" {
		goal = sess.Title
	}

	payload := map[string]any{
		"id":            sess.ID,
		"short_id":      sess.ShortID,
		"agent":         sess.Agent,
		"profile":       sess.Profile,
		"cwd":           sess.Cwd,
		"goal":          goal,
		"progress":      sess.Progress,
		"recent":        sess.Recent,
		"message_count": sess.MessageCount,
		"status":        string(sess.Status),
		"storage_path":  sess.StoragePath,
		"resume_cmd":    fmt.Sprintf("aim resume %s", sess.ShortID),
		"created_at":    sess.CreatedAt.Format(time.RFC3339),
		"updated_at":    sess.LastActiveAt.Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		},
	}
}

