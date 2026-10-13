package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
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
