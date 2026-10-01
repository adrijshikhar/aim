package agents

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/aim-cli/aim/internal/merge"
)

// MCPServerInfo describes a configured MCP server for an agent profile.
type MCPServerInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`             // e.g. "stdio", "http", "sse", "streamable_http"
	Status string `json:"status"`           // "enabled", "disabled"
	Auth   string `json:"auth,omitempty"`   // e.g. "OAuth", "connected", "auth required", "unsupported", "none"
	Target string `json:"target"`           // URL or collapsed command signature
	Origin string `json:"origin,omitempty"` // "host", "profile", "plugin:<name>"
}

// MCPListProvider is optionally implemented by adapters that support listing
// configured MCP servers for a given profile.
type MCPListProvider interface {
	ListMCPServers(ctx context.Context, profileName, profileDir string) ([]MCPServerInfo, error)
}

var (
	reNodeCjs = regexp.MustCompile(`([a-zA-Z0-9_\-\.]+\.(?:cjs|mjs|js|py|sh))`)
	reURLHost = regexp.MustCompile(`https?://([^/\s]+)`)
)

func parseExecCommand(script string) string {
	idx := strings.Index(script, "exec ")
	if idx == -1 {
		return ""
	}
	rem := strings.TrimSpace(script[idx+5:])
	if strings.HasPrefix(rem, "env ") {
		rem = strings.TrimSpace(rem[4:])
		for {
			rem = strings.TrimSpace(rem)
			if rem == "" {
				break
			}
			eqIdx := strings.IndexByte(rem, '=')
			spaceIdx := strings.IndexByte(rem, ' ')
			if eqIdx > 0 && (spaceIdx == -1 || eqIdx < spaceIdx) {
				if eqIdx+1 < len(rem) && (rem[eqIdx+1] == '"' || rem[eqIdx+1] == '\'') {
					q := rem[eqIdx+1]
					closeQ := strings.IndexByte(rem[eqIdx+2:], q)
					if closeQ != -1 {
						rem = rem[eqIdx+2+closeQ+1:]
						continue
					}
				}
				if spaceIdx == -1 {
					rem = ""
					break
				}
				rem = rem[spaceIdx+1:]
			} else {
				break
			}
		}
	}
	// Strip trailing semicolons or exits
	if semi := strings.IndexByte(rem, ';'); semi != -1 {
		rem = rem[:semi]
	}
	return strings.TrimSpace(rem)
}

// NormalizeAuth returns a clean human-readable authentication status string.
func NormalizeAuth(auth string) string {
	a := strings.ToLower(strings.TrimSpace(auth))
	switch a {
	case "o_auth", "oauth", "connected":
		return "OAuth"
	case "not_logged_in", "auth_required", "required":
		return "auth required"
	case "unsupported", "none", "":
		return "unsupported"
	default:
		return auth
	}
}

// CollapseCommand shortens lengthy commands, inline scripts, and deeply nested binary
// paths into a clean, human-readable signature suitable for terminal tables.
func CollapseCommand(cmd string, args []string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" && len(args) == 0 {
		return ""
	}
	base := filepath.Base(cmd)

	// Shell invocation with -c
	if (base == "sh" || base == "bash" || base == "zsh") && len(args) >= 2 && args[0] == "-c" {
		script := strings.TrimSpace(args[1])
		target := parseExecCommand(script)
		if target != "" {
			if strings.HasPrefix(target, "uvx") {
				parts := strings.Fields(target)
				last := parts[len(parts)-1]
				last = strings.Trim(last, `"'`)
				if strings.Contains(target, "--from") && !strings.HasPrefix(last, "-") {
					return "uvx " + last
				}
				if len(target) > 55 {
					return target[:52] + "..."
				}
				return target
			}
			if uh := reURLHost.FindStringSubmatch(target); len(uh) > 1 {
				parts := strings.Fields(target)
				if len(parts) > 0 {
					return fmt.Sprintf("%s (%s)", parts[0], uh[1])
				}
			}
			if len(target) > 55 {
				return target[:52] + "..."
			}
			return target
		}
		firstLine := strings.Split(script, "\n")[0]
		if len(firstLine) > 55 {
			return firstLine[:52] + "..."
		}
		return firstLine
	}

	// Node invocation with -e
	if (base == "node" || base == "bun") && len(args) >= 2 && args[0] == "-e" {
		script := args[1]
		if m := reNodeCjs.FindStringSubmatch(script); len(m) > 1 {
			return base + " .../" + m[1]
		}
		if len(script) > 50 {
			return base + " -e ..."
		}
		return base + " -e " + script
	}

	// Python invocation with -c
	if (base == "python" || base == "python3") && len(args) >= 2 && args[0] == "-c" {
		script := args[1]
		if strings.Contains(script, "hevo_mcp") || strings.Contains(script, "hevo_services") {
			return "python -m hevo_mcp.server"
		}
		if len(script) > 50 {
			return base + " -c ..."
		}
		return base + " -c " + script
	}

	displayCmd := base
	if !strings.Contains(cmd, "/") && !strings.Contains(cmd, `\`) {
		displayCmd = cmd
	}

	var parts []string
	parts = append(parts, displayCmd)
	for _, a := range args {
		if strings.Contains(a, "Authorization:") || strings.Contains(a, "KEY=") || strings.Contains(a, "TOKEN=") {
			parts = append(parts, "***")
		} else if len(a) > 40 {
			parts = append(parts, a[:37]+"...")
		} else {
			parts = append(parts, a)
		}
	}

	res := strings.Join(parts, " ")
	if len(res) > 60 {
		return res[:57] + "..."
	}
	return res
}

// MCPProvider is implemented by adapters whose config holds an MCP server map
// that aim merges per session (spec §4).
type MCPProvider interface {
	MCPCollections(profileDir, realHome string) []merge.Collection
	// IsBackground reports whether the native args start a session that
	// outlives the child aim waits on (spec R2): merge, run, no exit step.
	// profileArgs are profiles.<p>.args, which precede args when launched.
	IsBackground(profileArgs, args []string) bool
	// IsSession is false for an invocation that prints version or help and
	// exits: there is nothing to merge. Each CLI spells these differently.
	IsSession(profileArgs, args []string) bool
}

// ArgsMatch is true when a token before the first "--" of profileArgs then
// args (their launch order) is one of flags, or the very first token of
// either is one of firstWords. Checking each first token keeps profile flags
// such as `--model opus` from hiding the CLI's subcommand, while a profile's
// own `["app-server"]` still counts. A "--" in profileArgs makes all of args
// the agent's.
func ArgsMatch(profileArgs, args, flags, firstWords []string) bool {
	if firstWord(profileArgs, firstWords) {
		return true
	}
	if slices.Contains(profileArgs, "--") {
		return hasFlag(profileArgs, flags)
	}
	return hasFlag(profileArgs, flags) || hasFlag(args, flags) || firstWord(args, firstWords)
}

// hasFlag is true when a token before the first "--" is one of flags.
func hasFlag(args, flags []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if slices.Contains(flags, a) {
			return true
		}
	}
	return false
}

func firstWord(args, words []string) bool {
	return len(args) > 0 && !strings.HasPrefix(args[0], "-") && slices.Contains(words, args[0])
}
