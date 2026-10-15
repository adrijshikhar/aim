package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
)

// Server implements a read-only Model Context Protocol (MCP) server over stdio.
type Server struct {
	reg        *agents.Registry
	pm         *profile.ProfileManager
	sessionMgr *session.Manager
	baseDir    string
}

// NewServer creates a new read-only MCP server instance.
func NewServer(reg *agents.Registry, pm *profile.ProfileManager, baseDir string) *Server {
	return &Server{
		reg:     reg,
		pm:      pm,
		baseDir: baseDir,
	}
}

// SetSessionManager configures the optional session manager for session queries.
func (s *Server) SetSessionManager(sm *session.Manager) {
	s.sessionMgr = sm
}

// Serve reads JSON-RPC requests from in and writes responses to out.
func (s *Server) Serve(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	enc := json.NewEncoder(out)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			if encErr := enc.Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &Error{Code: -32700, Message: "Parse error"},
			}); encErr != nil {
				return encErr
			}
			continue
		}

		// JSON-RPC 2.0 notifications have no ID and MUST NOT receive a response
		if req.ID == nil {
			continue
		}

		resp := s.handleRequest(req)
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "aim",
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
						Name:        "aim_whoami",
						Description: "Returns the current active AIM profile and agent identity context (read-only).",
						InputSchema: map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						Name:        "aim_get_quota",
						Description: "Returns the current cached quota windows, limit reset times, and exhaustion percentages.",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"agent": map[string]any{
									"type":        "string",
									"description": "Optional agent name (e.g. 'claude', 'agy', 'codex'). Defaults to active agent.",
								},
								"profile": map[string]any{
									"type":        "string",
									"description": "Optional profile name. Defaults to active profile.",
								},
							},
						},
					},
					{
						Name:        "aim_list_profiles",
						Description: "Returns the list of configured profile names available in AIM.",
						InputSchema: map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						Name:        "aim_get_sessions",
						Description: "Returns recent AIM sessions with goal, working directory, and status.",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"limit": map[string]any{
									"type":        "number",
									"description": "Maximum number of sessions to return (default: 10)",
								},
								"cwd": map[string]any{
									"type":        "string",
									"description": "Optional working directory filter",
								},
								"profile": map[string]any{
									"type":        "string",
									"description": "Optional profile filter",
								},
							},
						},
					},
					{
						Name:        "aim_get_session_details",
						Description: "Returns detailed information, recent turns, and resumption command for a specific session ID.",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"session_id": map[string]any{
									"type":        "string",
									"description": "Session ID or short ID prefix",
								},
							},
							"required": []string{"session_id"},
						},
					},
					{
						Name:        "aim_get_active_binding",
						Description: "Discovers any .aim-profile directory binding for the specified or current working directory.",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"cwd": map[string]any{
									"type":        "string",
									"description": "Target directory to check. Defaults to current working directory.",
								},
							},
						},
					},
					{
						Name:        "aim_diagnostics",
						Description: "Runs in-memory health checks and reports authentication and environment diagnostics.",
						InputSchema: map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						Name:        "aim_daemon_status",
						Description: "Returns background daemon execution status, port, PID, and last sync timestamp.",
						InputSchema: map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
				},
			},
		}

	case "tools/call":
		return s.handleToolCall(req)

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}
