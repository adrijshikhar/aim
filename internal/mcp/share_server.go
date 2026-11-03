package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/share"
)

// ShareMCPServer implements an embedded Model Context Protocol (MCP) server
// exposing lease self-inspection and renewal tools for autonomous agents running inside leased profiles.
type ShareMCPServer struct {
	aimDir      string
	profilesDir string
	profileName string

	client *share.Client
	mu     sync.Mutex
}

// NewShareMCPServer creates a new ShareMCPServer instance.
func NewShareMCPServer(aimDir, profilesDir, profileName string) *ShareMCPServer {
	return &ShareMCPServer{
		aimDir:      aimDir,
		profilesDir: profilesDir,
		profileName: profileName,
	}
}

// SetClient overrides the internal share client (useful for unit testing with httptest).
func (s *ShareMCPServer) SetClient(c *share.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = c
}

func (s *ShareMCPServer) getClient() *share.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		return s.client
	}
	return share.NewClient(4319)
}

func (s *ShareMCPServer) getProfileName() string {
	if s.profileName != "" {
		return s.profileName
	}
	return os.Getenv("AIM_PROFILE")
}

func (s *ShareMCPServer) getProfilesDir() string {
	if s.profilesDir != "" {
		return s.profilesDir
	}
	return filepath.Join(s.getAimDir(), "profiles")
}

func (s *ShareMCPServer) getAimDir() string {
	if s.aimDir != "" {
		return s.aimDir
	}
	return config.BaseDir()
}

// ServeStdio reads newline-delimited JSON-RPC messages from in and writes responses to out.
func (s *ShareMCPServer) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		if err := s.HandleMessage(ctx, line, out); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return ctx.Err()
}

// HandleMessage parses a single JSON-RPC message and dispatches it to the appropriate handler.
func (s *ShareMCPServer) HandleMessage(ctx context.Context, data []byte, out io.Writer) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return s.writeResponse(out, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &Error{Code: -32700, Message: "Parse error"},
		})
	}

	// Notifications in JSON-RPC 2.0 have no ID and MUST NOT receive a response
	if req.ID == nil {
		return nil
	}

	resp := s.dispatch(ctx, req)
	return s.writeResponse(out, resp)
}

func (s *ShareMCPServer) writeResponse(out io.Writer, resp JSONRPCResponse) error {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	respBytes = append(respBytes, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = out.Write(respBytes)
	return err
}

func (s *ShareMCPServer) dispatch(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "aim-share",
					"version": "1.0.0",
				},
				"capabilities": map[string]any{
					"tools": map[string]bool{"listChanged": false},
				},
			},
		}

	case "ping":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}

	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": []Tool{
					{
						Name:        "aim_request_lease",
						Description: "Request an extension or renewal of the current leased profile from the lender teammate.",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"hours": map[string]any{
									"type":        "integer",
									"description": "Requested extension duration in hours (default 2, min 1, max 72).",
									"minimum":     1,
									"maximum":     72,
									"default":     2,
								},
							},
						},
					},
					{
						Name:        "aim_lease_status",
						Description: "Check current lease status, remaining time, and lender details for this profile.",
						InputSchema: map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
				},
			},
		}

	case "tools/call":
		return s.handleToolCall(ctx, req)

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func (s *ShareMCPServer) handleToolCall(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
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
	case "aim_lease_status":
		return s.callLeaseStatus(req.ID)
	case "aim_request_lease":
		return s.callRequestLease(ctx, req.ID, params.Arguments)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: fmt.Sprintf("Tool not found: %s", params.Name)},
		}
	}
}

func (s *ShareMCPServer) callLeaseStatus(id any) JSONRPCResponse {
	profName := s.getProfileName()
	if profName == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "Error: No leased profile specified or detected in environment (set --profile or AIM_PROFILE)."},
				},
				IsError: true,
			},
		}
	}

	profilesDir := s.getProfilesDir()
	leaseFile := filepath.Join(profilesDir, profName, ".aim-lease.json")
	data, err := os.ReadFile(leaseFile)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Profile %q is not a leased profile or lease file is missing.", profName)},
				},
				IsError: true,
			},
		}
	}

	var meta share.LeaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Corrupt lease metadata in %s: %v", leaseFile, err)},
				},
				IsError: true,
			},
		}
	}

	now := time.Now()
	remaining := "Expired"
	if now.Before(meta.ExpiresAt) {
		remaining = meta.ExpiresAt.Sub(now).Round(time.Second).String()
	}

	revokedStr := "No"
	if meta.Revoked {
		revokedStr = "Yes"
	}

	offlineStaleStr := "No"
	if !meta.LastVerifiedAt.IsZero() && now.Sub(meta.LastVerifiedAt) > 5*time.Minute {
		offlineStaleStr = "Yes (>5m disconnected)"
	} else if meta.LastVerifiedAt.IsZero() && !meta.GrantedAt.IsZero() && now.Sub(meta.GrantedAt) > 5*time.Minute {
		offlineStaleStr = "Yes (>5m disconnected)"
	}

	agentStr := meta.Agent
	if len(meta.Agents) > 0 {
		agentStr = strings.Join(meta.Agents, ", ")
	}

	text := fmt.Sprintf("Lease Status for %s:\n"+
		"  Lease ID: %s\n"+
		"  Source Peer: %s\n"+
		"  Source Profile: %s\n"+
		"  Agent: %s\n"+
		"  Expiration: %s\n"+
		"  Remaining Duration: %s\n"+
		"  Revoked: %s\n"+
		"  Offline Stale: %s",
		profName,
		meta.LeaseID,
		meta.SourcePeer,
		meta.SourceProfile,
		agentStr,
		meta.ExpiresAt.Format(time.RFC3339),
		remaining,
		revokedStr,
		offlineStaleStr,
	)

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: text},
			},
		},
	}
}

func (s *ShareMCPServer) callRequestLease(ctx context.Context, id any, rawArgs json.RawMessage) JSONRPCResponse {
	hours := 2
	if len(rawArgs) > 0 && string(rawArgs) != "null" {
		var args struct {
			Hours any `json:"hours"`
		}
		if err := json.Unmarshal(rawArgs, &args); err == nil && args.Hours != nil {
			switch v := args.Hours.(type) {
			case float64:
				hours = int(v)
			case int:
				hours = v
			case int64:
				hours = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					hours = parsed
				}
			}
		}
	}

	if hours < 1 || hours > 72 {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid hours %d: requested renewal duration must be between 1 and 72 hours.", hours)},
				},
				IsError: true,
			},
		}
	}

	profName := s.getProfileName()
	if profName == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "Error: No leased profile specified or detected in environment (set --profile or AIM_PROFILE)."},
				},
				IsError: true,
			},
		}
	}

	profilesDir := s.getProfilesDir()
	leaseFile := filepath.Join(profilesDir, profName, ".aim-lease.json")
	data, err := os.ReadFile(leaseFile)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Profile %q is not a leased profile or lease file is missing.", profName)},
				},
				IsError: true,
			},
		}
	}

	var meta share.LeaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Corrupt lease metadata in %s: %v", leaseFile, err)},
				},
				IsError: true,
			},
		}
	}

	if meta.SourcePeer == "" || meta.SourceProfile == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: "Lease metadata is missing SourcePeer or SourceProfile."},
				},
				IsError: true,
			},
		}
	}

	client := s.getClient()
	aimDir := s.getAimDir()

	targetHost := meta.SourcePeer
	if aimDir != "" {
		store := share.NewStore(aimDir)
		for _, p := range store.GetPeers() {
			shortHost := strings.TrimSuffix(strings.ToLower(p.HostName), ".local")
			if strings.EqualFold(p.HostName, meta.SourcePeer) ||
				strings.EqualFold(shortHost, strings.ToLower(meta.SourcePeer)) ||
				p.VPNIP == meta.SourcePeer ||
				(p.ID != "" && p.ID == meta.SourcePeer) ||
				(meta.LenderNodeID != "" && p.ID == meta.LenderNodeID) {
				if p.VPNIP == meta.SourcePeer {
					targetHost = p.VPNIP
				} else if p.HostName != "" {
					targetHost = p.HostName
				} else if p.VPNIP != "" {
					targetHost = p.VPNIP
				}
				break
			}
		}
		_ = store.Close()
	}

	agentName := meta.Agent
	if agentName == "" && len(meta.Agents) > 0 {
		agentName = meta.Agents[0]
	}
	if agentName == "" {
		agentName = "all"
	}

	reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
	resp, err := client.RequestLease(reqCtx, targetHost, meta.SourceProfile, agentName, hours)
	reqCancel()
	if err != nil && !strings.Contains(targetHost, ".") && !strings.Contains(targetHost, ":") && !strings.HasPrefix(targetHost, "http://") && !strings.HasPrefix(targetHost, "https://") {
		// Retry with .local for zero-conf mDNS resolution
		retryCtx, retryCancel := context.WithTimeout(ctx, 10*time.Second)
		resp, err = client.RequestLease(retryCtx, targetHost+".local", meta.SourceProfile, agentName, hours)
		retryCancel()
	}

	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Lease renewal request failed: %v", err)},
				},
				IsError: true,
			},
		}
	}

	if resp.Status == "pending" {
		pollTicker := time.NewTicker(2 * time.Second)
		defer pollTicker.Stop()
		pollCtx, pollCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer pollCancel()

		approved := false
		for !approved {
			select {
			case <-pollCtx.Done():
				return JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      id,
					Result: ToolResult{
						Content: []ToolContent{
							{Type: "text", Text: "Lease renewal request timed out waiting for lender approval after 2 minutes."},
						},
						IsError: true,
					},
				}
			case <-pollTicker.C:
				pResp, pErr := client.PollLease(pollCtx, targetHost, resp.RequestID)
				if pErr == nil && pResp != nil {
					if pResp.Status == "approved" {
						resp = pResp
						approved = true
						break
					} else if pResp.Status == "denied" {
						return JSONRPCResponse{
							JSONRPC: "2.0",
							ID:      id,
							Result: ToolResult{
								Content: []ToolContent{
									{Type: "text", Text: fmt.Sprintf("Teammate denied lease renewal request: %s", pResp.Message)},
								},
								IsError: true,
							},
						}
					}
				}
			}
		}
	}

	if resp.Status != "approved" {
		msg := resp.Message
		if msg == "" {
			msg = resp.Status
		}
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Renewal request was not approved (status: %s): %s", resp.Status, msg)},
				},
				IsError: true,
			},
		}
	}

	mgr := share.NewLeaseManager(profilesDir)
	mgr.SetAimDir(aimDir)
	installedName, err := mgr.InstallLease(meta.SourcePeer, meta.SourceProfile, resp)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: ToolResult{
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to install renewed lease: %v", err)},
				},
				IsError: true,
			},
		}
	}

	text := fmt.Sprintf("Successfully renewed lease for %s (valid until %s).\nPaused sessions have been automatically unfrozen.",
		installedName, resp.ExpiresAt.Format(time.RFC3339))
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: text},
			},
		},
	}
}
