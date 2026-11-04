package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "hook",
		Short:  "Internal agent lifecycle hook handlers",
		Hidden: true,
	}

	var agentName string
	guardCmd := &cobra.Command{
		Use:   "guard <profile>",
		Short: "Enforce lease validity on prompt and tool execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookGuard(cmd, args, cmd.InOrStdin())
		},
	}
	guardCmd.Flags().StringVar(&agentName, "agent", "", "Agent engine name")
	cmd.AddCommand(guardCmd)

	return cmd
}

func runHookGuard(cmd *cobra.Command, args []string, stdinReader io.Reader) error {
	if len(args) == 0 {
		return fmt.Errorf("profile name required")
	}
	profName := args[0]
	aimDir := config.BaseDir()
	profDir := filepath.Join(aimDir, "profiles", profName)

	var stdinBytes []byte
	if stdinReader != nil {
		if f, ok := stdinReader.(*os.File); ok {
			fi, err := f.Stat()
			if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
				stdinBytes, _ = io.ReadAll(f)
			}
		} else {
			stdinBytes, _ = io.ReadAll(stdinReader)
		}
	}

	var out io.Writer = os.Stdout
	if cmd != nil {
		out = cmd.OutOrStdout()
	}

	return executeHookGuardForDir(aimDir, profDir, profName, stdinBytes, out)
}

func executeHookGuardForDir(aimDir, profDir, profName string, stdinBytes []byte, out io.Writer) error {
	leaseFile := filepath.Join(profDir, ".aim-lease.json")
	data, err := os.ReadFile(leaseFile)
	if err != nil {
		return fmt.Errorf("lease file missing: %w", err)
	}

	var meta share.LeaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("corrupt lease metadata: %w", err)
	}

	// Tool whitelist inspection
	toolName := ""
	if len(stdinBytes) > 0 {
		var toolPayload struct {
			ToolName string `json:"tool_name"`
			Name     string `json:"name"`
			Tool     string `json:"tool"`
		}
		if err := json.Unmarshal(stdinBytes, &toolPayload); err == nil {
			if toolPayload.ToolName != "" {
				toolName = toolPayload.ToolName
			} else if toolPayload.Name != "" {
				toolName = toolPayload.Name
			} else if toolPayload.Tool != "" {
				toolName = toolPayload.Tool
			}
		}
		if toolName == "" {
			var rawMap map[string]any
			if err := json.Unmarshal(stdinBytes, &rawMap); err == nil {
				for _, k := range []string{"tool_name", "name", "tool", "toolName"} {
					if v, ok := rawMap[k].(string); ok && v != "" {
						toolName = v
						break
					}
				}
				if toolName == "" {
					if tc, ok := rawMap["tool_call"].(map[string]any); ok {
						if n, ok := tc["name"].(string); ok {
							toolName = n
						}
					}
				}
			}
		}
	}

	now := time.Now()
	isRevoked := meta.Revoked
	isExpired := now.After(meta.ExpiresAt)

	// Anti-hoarding 5-minute heartbeat check (also checks negative duration to prevent future clock bypass)
	isOfflineStale := false
	if !meta.LastVerifiedAt.IsZero() {
		diff := now.Sub(meta.LastVerifiedAt)
		isOfflineStale = diff > 5*time.Minute || diff < -1*time.Minute
	} else if !meta.GrantedAt.IsZero() {
		diff := now.Sub(meta.GrantedAt)
		isOfflineStale = diff > 5*time.Minute || diff < -1*time.Minute
	}

	// Cryptographic signature check against pinned lender key in peers.json (fail closed)
	isInvalidSig := false
	var pinnedPubKey ed25519.PublicKey
	if aimDir != "" {
		store := share.NewStore(aimDir)
		for _, p := range store.GetPeers() {
			if (meta.LenderNodeID != "" && p.ID == meta.LenderNodeID) ||
				strings.EqualFold(p.HostName, meta.SourcePeer) ||
				p.ID == meta.SourcePeer ||
				p.VPNIP == meta.SourcePeer {
				if len(p.PublicKey) == ed25519.PublicKeySize {
					pinnedPubKey = p.PublicKey
					break
				}
			}
		}
		_ = store.Close()
	}

	if len(pinnedPubKey) != ed25519.PublicKeySize || meta.Signature == "" {
		isInvalidSig = true
	} else {
		lenderID := meta.LenderNodeID
		if lenderID == "" {
			lenderID = meta.SourcePeer
		}
		expStr := meta.ExpiresAt.UTC().Format(time.RFC3339)

		valid := false
		if meta.BorrowerID != "" {
			valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, lenderID, meta.BorrowerID, meta.SourceProfile, meta.Agent, expStr, meta.Signature)
			if !valid && meta.SourceProfile != profName {
				valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, lenderID, meta.BorrowerID, profName, meta.Agent, expStr, meta.Signature)
			}
			if !valid && meta.SourcePeer != lenderID {
				valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, meta.SourcePeer, meta.BorrowerID, meta.SourceProfile, meta.Agent, expStr, meta.Signature)
			}
		}
		if !valid {
			h, _ := os.Hostname()
			if h != "" {
				valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, lenderID, h, meta.SourceProfile, meta.Agent, expStr, meta.Signature)
				if !valid && meta.SourceProfile != profName {
					valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, lenderID, h, profName, meta.Agent, expStr, meta.Signature)
				}
				if !valid && meta.SourcePeer != lenderID {
					valid = share.VerifyLeaseSignature(pinnedPubKey, meta.LeaseID, meta.SourcePeer, h, meta.SourceProfile, meta.Agent, expStr, meta.Signature)
				}
			}
		}
		if !valid {
			isInvalidSig = true
		}
	}

	if isRevoked || isExpired || isOfflineStale || isInvalidSig {
		if toolName == "aim_request_lease" || toolName == "aim_lease_status" {
			return nil
		}

		reason := "Inactive"
		if isRevoked {
			reason = "Revoked by Lender"
		} else if isExpired {
			reason = "Lease Expired"
		} else if isOfflineStale {
			reason = "Heartbeat Stale (>5m disconnected from VPN)"
		} else if isInvalidSig {
			reason = "Cryptographic Verification Failed"
		}

		banner := fmt.Sprintf("\n%s Lease for '%s' is inactive: %s.\n"+
			"%s Prompt & tool execution paused.\n"+
			"%s Run 'aim share renew %s' to continue.\n",
			lipgloss.NewStyle().Foreground(tui.StatusRed).Render("[aim]"),
			profName, reason,
			lipgloss.NewStyle().Foreground(tui.StatusYellow).Render("[aim]"),
			lipgloss.NewStyle().Foreground(tui.AccentCyan).Render("[aim]"),
			profName,
		)
		fmt.Fprint(out, banner)
		return &ExitError{Code: 1}
	}

	return nil
}
