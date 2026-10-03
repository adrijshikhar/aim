package presenter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

func TestSpinner_Lifecycle(t *testing.T) {
	var buf bytes.Buffer
	spin := NewSpinner(&buf, "Testing spinner...")
	if spin == nil {
		t.Fatal("expected non-nil spinner")
	}

	spin.Start()
	// Calling Start again should be idempotent
	spin.Start()

	time.Sleep(100 * time.Millisecond)
	spin.UpdateText("Updated text")
	time.Sleep(100 * time.Millisecond)

	spin.Stop()
	// Calling Stop again should be safe
	spin.Stop()

	out := buf.String()
	if !strings.Contains(out, "Testing spinner...") && !strings.Contains(out, "Updated text") {
		t.Errorf("spinner output did not contain expected text, got: %q", out)
	}
}

func TestRenderMCPListTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	RenderMCPListTable(&buf, "codex", "default", nil)
	out := buf.String()

	if !strings.Contains(out, "Configured MCP Servers (codex: default)") {
		t.Errorf("expected header banner, got %q", out)
	}
	if !strings.Contains(out, "(no MCP servers configured)") {
		t.Errorf("expected empty message, got %q", out)
	}
}

func TestRenderMCPListTable_Populated(t *testing.T) {
	var buf bytes.Buffer
	servers := []agents.MCPServerInfo{
		{
			Name:   "github",
			Status: "enabled",
			Auth:   "OAuth",
			Type:   "stdio",
			Target: "npx -y @modelcontextprotocol/server-github",
		},
		{
			Name:   "memory",
			Status: "disabled",
			Auth:   "none",
			Type:   "stdio",
			Target: "npx -y @modelcontextprotocol/server-memory",
		},
	}
	RenderMCPListTable(&buf, "claude", "work", servers)
	out := buf.String()

	if !strings.Contains(out, "Configured MCP Servers (claude: work)") {
		t.Errorf("expected header banner, got %q", out)
	}
	if !strings.Contains(out, "github") || !strings.Contains(out, "memory") {
		t.Errorf("expected server names in table, got %q", out)
	}
	if !strings.Contains(out, "Total: 2 server(s) configured (1 enabled, 1 disabled)") {
		t.Errorf("expected summary with enabled/disabled counts, got %q", out)
	}
}

func TestPrintTable(t *testing.T) {
	headers := []string{"COL1", "COL2"}
	rows := [][]string{
		{"val1", "val2"},
		{"val3", "val4"},
	}

	var bufColor bytes.Buffer
	PrintTable(&bufColor, headers, rows, true)
	outColor := bufColor.String()
	if !strings.Contains(outColor, "COL1") || !strings.Contains(outColor, "val1") {
		t.Errorf("expected colored table content, got: %q", outColor)
	}

	var bufPlain bytes.Buffer
	PrintTable(&bufPlain, headers, rows, false)
	outPlain := bufPlain.String()
	if !strings.Contains(outPlain, "COL1") || !strings.Contains(outPlain, "val1") {
		t.Errorf("expected plain table content, got: %q", outPlain)
	}
}

func TestRenderCLIBar(t *testing.T) {
	plain := RenderCLIBar(75, 10, usage.StatusOK, false)
	if !strings.Contains(plain, "75%") {
		t.Errorf("expected plain bar with 75%%, got: %q", plain)
	}

	colored := RenderCLIBar(75, 10, usage.StatusOK, true)
	if !strings.Contains(colored, "75%") {
		t.Errorf("expected colored bar with 75%%, got: %q", colored)
	}

	// Clamp tests
	clampedLow := RenderCLIBar(-10, 10, usage.StatusCritical, true)
	if !strings.Contains(clampedLow, "-10%") {
		t.Errorf("expected -10%% in output, got: %q", clampedLow)
	}

	clampedHigh := RenderCLIBar(150, 10, usage.StatusOK, true)
	if !strings.Contains(clampedHigh, "150%") {
		t.Errorf("expected 150%% in output, got: %q", clampedHigh)
	}

	zeroWidth := RenderCLIBar(50, 0, usage.StatusOK, true)
	if zeroWidth != "50%" {
		t.Errorf("expected '50%%' for zero width, got: %q", zeroWidth)
	}
}

func TestRenderCLIStatus(t *testing.T) {
	plain := RenderCLIStatus(usage.StatusOK, false)
	if plain != "OK" {
		t.Errorf("expected 'OK', got %q", plain)
	}

	colored := RenderCLIStatus(usage.StatusOK, true)
	if !strings.Contains(colored, "OK") {
		t.Errorf("expected colored status to contain 'OK', got %q", colored)
	}
}

func TestRenderDiagnosticResult(t *testing.T) {
	testCases := []struct {
		res      agents.DiagnosticResult
		expected string
	}{
		{
			res:      agents.DiagnosticResult{Status: "OK", Category: "Binary", Message: "found at /bin/sh"},
			expected: "[OK]",
		},
		{
			res:      agents.DiagnosticResult{Status: "WARN", Category: "Memory", Message: "high utilization"},
			expected: "[WARN]",
		},
		{
			res:      agents.DiagnosticResult{Status: "FAIL", Category: "Auth", Message: "token expired"},
			expected: "[FAIL]",
		},
		{
			res:      agents.DiagnosticResult{Status: "INFO", Category: "Info", Message: "general info"},
			expected: "[INFO]",
		},
	}

	for _, tc := range testCases {
		badge, cat, msg := RenderDiagnosticResult(tc.res)
		if !strings.Contains(badge, tc.expected) {
			t.Errorf("expected badge to contain %q, got: %q", tc.expected, badge)
		}
		if !strings.Contains(cat, tc.res.Category+":") {
			t.Errorf("expected category %q, got: %q", tc.res.Category+":", cat)
		}
		if !strings.Contains(msg, tc.res.Message) {
			t.Errorf("expected msg %q, got: %q", tc.res.Message, msg)
		}
	}
}
