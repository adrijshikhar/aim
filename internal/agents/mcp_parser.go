package agents

import (
	"strings"
)

// ParseMCPServerMap converts an untyped dictionary representation of an MCP server
// (as read from JSON/TOML configurations across Claude, Antigravity, and Codex)
// into a standardized MCPServerInfo structure.
func ParseMCPServerMap(name string, v map[string]any, origin string) MCPServerInfo {
	status := "enabled"
	if dis, ok := v["disabled"].(bool); ok && dis {
		status = "disabled"
	} else if en, ok := v["enabled"].(bool); ok && !en {
		status = "disabled"
	}

	typ, _ := v["type"].(string)
	serverURL, _ := v["serverUrl"].(string)
	urlStr, _ := v["url"].(string)
	cmdStr, _ := v["command"].(string)

	var argsSlice []string
	if rawArgs, ok := v["args"].([]any); ok {
		for _, arg := range rawArgs {
			if as, ok := arg.(string); ok {
				argsSlice = append(argsSlice, as)
			}
		}
	} else if strArgs, ok := v["args"].([]string); ok {
		argsSlice = append(argsSlice, strArgs...)
	}

	target := serverURL
	if target == "" {
		target = urlStr
	}

	auth := "unsupported"
	if target != "" {
		if typ == "" {
			typ = "http"
		}
		if strings.Contains(strings.ToLower(target), "oauth") || strings.Contains(strings.ToLower(name), "oauth") {
			auth = "OAuth"
		} else if headers, ok := v["headers"].(map[string]any); ok && len(headers) > 0 {
			auth = "connected"
		} else if authStatus, ok := v["auth_status"].(string); ok && authStatus != "" {
			auth = NormalizeAuth(authStatus)
		}
	} else {
		if typ == "" {
			typ = "stdio"
		}
		target = CollapseCommand(cmdStr, argsSlice)
		if env, ok := v["env"].(map[string]any); ok {
			for ek := range env {
				if strings.Contains(ek, "KEY") || strings.Contains(ek, "TOKEN") {
					auth = "connected"
					break
				}
			}
		} else if authStatus, ok := v["auth_status"].(string); ok && authStatus != "" {
			auth = NormalizeAuth(authStatus)
		}
	}

	return MCPServerInfo{
		Name:   name,
		Type:   typ,
		Status: status,
		Auth:   auth,
		Target: target,
		Origin: origin,
	}
}
