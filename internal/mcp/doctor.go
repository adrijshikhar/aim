package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
)

// DiagnosticStatus represents the result of a diagnostic check.
type DiagnosticStatus string

const (
	StatusPass DiagnosticStatus = "OK"
	StatusWarn DiagnosticStatus = "WARN"
	StatusFail DiagnosticStatus = "FAIL"
)

// MCPDiagnosticResult represents an individual diagnostic check finding.
type MCPDiagnosticResult struct {
	Category string           `json:"category"`
	Status   DiagnosticStatus `json:"status"`
	Message  string           `json:"message"`
	Remedy   string           `json:"remedy,omitempty"`
}

// DoctorReport contains the aggregated results of all MCP diagnostic checks.
type DoctorReport struct {
	OverallStatus DiagnosticStatus      `json:"overall_status"`
	Results       []MCPDiagnosticResult `json:"results"`
	DurationMs    int64                 `json:"duration_ms"`
}

// RunMCPDoctor runs a complete diagnostic suite checking protocol conformance,
// client configurations, and third-party MCP server binaries.
func RunMCPDoctor(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager, homeDir string) DoctorReport {
	start := time.Now()
	var results []MCPDiagnosticResult

	// 1. Run in-process protocol self-check
	protoResults := runProtocolSelfCheck(ctx, reg, pm)
	results = append(results, protoResults...)

	// 2. Audit detected client configurations and third-party binaries
	clientResults := runClientAndBinaryAudit(homeDir)
	results = append(results, clientResults...)

	// Compute overall status
	overall := StatusPass
	for _, res := range results {
		if res.Status == StatusFail {
			overall = StatusFail
			break
		} else if res.Status == StatusWarn && overall != StatusFail {
			overall = StatusWarn
		}
	}

	return DoctorReport{
		OverallStatus: overall,
		Results:       results,
		DurationMs:    time.Since(start).Milliseconds(),
	}
}

func runProtocolSelfCheck(ctx context.Context, reg *agents.Registry, pm *profile.ProfileManager) []MCPDiagnosticResult {
	var results []MCPDiagnosticResult

	serverInR, clientInW := io.Pipe()
	clientOutR, serverOutW := io.Pipe()

	srv := NewServer(reg, pm, "")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(serverInR, serverOutW)
	}()

	handshakeStart := time.Now()
	clientWriter := bufio.NewWriter(clientInW)
	clientScanner := bufio.NewScanner(clientOutR)

	type checkStep struct {
		name    string
		request string
		verify  func(respLine []byte) error
	}

	steps := []checkStep{
		{
			name:    "initialize",
			request: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}` + "\n",
			verify: func(b []byte) error {
				var resp struct {
					Result struct {
						ProtocolVersion string `json:"protocolVersion"`
						ServerInfo      struct {
							Name string `json:"name"`
						} `json:"serverInfo"`
					} `json:"result"`
				}
				if err := json.Unmarshal(b, &resp); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
				if resp.Result.ProtocolVersion != "2024-11-05" {
					return fmt.Errorf("expected protocolVersion '2024-11-05', got %q", resp.Result.ProtocolVersion)
				}
				if resp.Result.ServerInfo.Name != "aim" {
					return fmt.Errorf("expected server name 'aim', got %q", resp.Result.ServerInfo.Name)
				}
				return nil
			},
		},
		{
			name:    "ping",
			request: `{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n",
			verify: func(b []byte) error {
				var resp struct {
					Result any `json:"result"`
				}
				if err := json.Unmarshal(b, &resp); err != nil {
					return fmt.Errorf("failed to decode ping: %w", err)
				}
				return nil
			},
		},
		{
			name:    "tools/list",
			request: `{"jsonrpc":"2.0","id":3,"method":"tools/list"}` + "\n",
			verify: func(b []byte) error {
				var resp struct {
					Result struct {
						Tools []struct {
							Name string `json:"name"`
						} `json:"tools"`
					} `json:"result"`
				}
				if err := json.Unmarshal(b, &resp); err != nil {
					return fmt.Errorf("failed to decode tools/list: %w", err)
				}
				if len(resp.Result.Tools) < 8 {
					return fmt.Errorf("expected at least 8 tools, got %d", len(resp.Result.Tools))
				}
				return nil
			},
		},
		{
			name:    "tools/call aim_whoami",
			request: `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"aim_whoami","arguments":{}}}` + "\n",
			verify: func(b []byte) error {
				var resp struct {
					Result struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
						IsError bool `json:"isError"`
					} `json:"result"`
				}
				if err := json.Unmarshal(b, &resp); err != nil {
					return fmt.Errorf("failed to decode whoami: %w", err)
				}
				if resp.Result.IsError || len(resp.Result.Content) == 0 {
					return fmt.Errorf("whoami tool call returned error")
				}
				return nil
			},
		},
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = clientInW.CloseWithError(ctx.Err())
			_ = clientOutR.CloseWithError(ctx.Err())
			_ = serverInR.CloseWithError(ctx.Err())
			_ = serverOutW.CloseWithError(ctx.Err())
		case <-done:
		}
	}()

	protoErr := ""
	for _, step := range steps {
		if ctx.Err() != nil {
			protoErr = "self-check timed out"
			break
		}
		if _, err := clientWriter.WriteString(step.request); err != nil {
			protoErr = fmt.Sprintf("failed to write %s: %v", step.name, err)
			break
		}
		if err := clientWriter.Flush(); err != nil {
			protoErr = fmt.Sprintf("failed to flush %s: %v", step.name, err)
			break
		}

		if !clientScanner.Scan() {
			protoErr = fmt.Sprintf("no response received for %s: %v", step.name, clientScanner.Err())
			break
		}
		respLine := clientScanner.Bytes()
		if err := step.verify(respLine); err != nil {
			protoErr = fmt.Sprintf("verification failed for %s: %v", step.name, err)
			break
		}
	}

	// Close client pipes to unblock server and wait for goroutine shutdown
	_ = clientInW.Close()
	_ = serverOutW.Close()
	_ = serverInR.Close()
	_ = clientOutR.Close()
	wg.Wait()

	handshakeMs := time.Since(handshakeStart).Milliseconds()
	if protoErr != "" {
		results = append(results, MCPDiagnosticResult{
			Category: "Protocol",
			Status:   StatusFail,
			Message:  fmt.Sprintf("In-process stdio handshake failed: %s", protoErr),
			Remedy:   "Verify that 'aim' binary is not corrupted and standard library JSON-RPC handlers are intact",
		})
	} else {
		results = append(results, MCPDiagnosticResult{
			Category: "Protocol",
			Status:   StatusPass,
			Message:  fmt.Sprintf("In-process stdio handshake passed (MCP 2024-11-05, 8 tools verified) in %dms", handshakeMs),
		})
	}

	return results
}

func runClientAndBinaryAudit(homeDir string) []MCPDiagnosticResult {
	var results []MCPDiagnosticResult
	auditedCommands := make(map[string]bool)

	for _, driver := range AllDrivers() {
		if !driver.Detect(homeDir) {
			continue
		}

		cfgPath, err := driver.ConfigPath(homeDir, "")
		if err != nil {
			continue
		}

		servers, err := driver.ReadServers(cfgPath)
		if err != nil {
			results = append(results, MCPDiagnosticResult{
				Category: "Config",
				Status:   StatusFail,
				Message:  fmt.Sprintf("%s configuration corrupted at %s: %v", driver.Name(), cfgPath, err),
				Remedy:   fmt.Sprintf("Fix JSON syntax errors in %s", cfgPath),
			})
			continue
		}

		if _, hasAim := servers["aim"]; hasAim {
			results = append(results, MCPDiagnosticResult{
				Category: "Client",
				Status:   StatusPass,
				Message:  fmt.Sprintf("%s: registered in %s", driver.Name(), cfgPath),
			})
		} else {
			results = append(results, MCPDiagnosticResult{
				Category: "Client",
				Status:   StatusWarn,
				Message:  fmt.Sprintf("%s: detected at %s, but 'aim' MCP server is not registered", driver.Name(), cfgPath),
				Remedy:   fmt.Sprintf("Run 'aim mcp install %s'", driver.ID()),
			})
		}

		// Audit third-party binaries
		for sName, entry := range servers {
			cmdName := strings.TrimSpace(entry.Command)
			if cmdName == "" || cmdName == "aim" || cmdName == "remote" {
				continue
			}

			// Extract binary name: if whole cmdName exists on disk, use it directly (e.g. "/path with spaces/bin"),
			// otherwise extract first token.
			binName := cmdName
			if _, err := os.Stat(cmdName); err != nil {
				tokens := strings.Fields(cmdName)
				if len(tokens) > 0 {
					binName = tokens[0]
				}
			}

			auditKey := binName + "::" + sName
			if auditedCommands[auditKey] {
				continue
			}
			auditedCommands[auditKey] = true

			resolvedPath, lookErr := exec.LookPath(binName)
			if lookErr != nil {
				results = append(results, MCPDiagnosticResult{
					Category: "Third-Party",
					Status:   StatusFail,
					Message:  fmt.Sprintf("Server %q in %s: executable %q not found in PATH", sName, driver.Name(), binName),
					Remedy:   fmt.Sprintf("Install %q or update command in %s", binName, cfgPath),
				})
			} else {
				results = append(results, MCPDiagnosticResult{
					Category: "Third-Party",
					Status:   StatusPass,
					Message:  fmt.Sprintf("Server %q: binary %q found at %s", sName, binName, resolvedPath),
				})
			}
		}
	}

	return results
}
