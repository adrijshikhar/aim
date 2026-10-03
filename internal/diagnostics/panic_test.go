package diagnostics

import (
	"runtime"
	"strings"
	"testing"
)

func TestSanitizeStack_HomeDirectoryRedacted(t *testing.T) {
	home := "/Users/testuser"
	rawStack := `goroutine 1 [running]:
main.someFunc()
	/Users/testuser/Projects/aim/cmd/aim/main.go:42 +0x123
github.com/aim-cli/aim/internal/runner.Run()
	/Users/testuser/go/pkg/mod/github.com/aim-cli/aim@v0.12.0/runner.go:88 +0x456
`
	sanitized := SanitizeStack(rawStack, home)
	if strings.Contains(sanitized, "/Users/testuser") {
		t.Errorf("expected home directory to be redacted, got:\n%s", sanitized)
	}
	if !strings.Contains(sanitized, "~/Projects/aim/cmd/aim/main.go:42") {
		t.Errorf("expected ~/ relative path, got:\n%s", sanitized)
	}
}

func TestSanitizeStack_SecretsRedacted(t *testing.T) {
	rawStack := `panic: unexpected error with token ya29.a0AfH6SMB_secret_access_token_12345
and Bearer ghp_secretGitHubToken987654321
and secret_key=sk-proj-supersecretkey1234567890
`
	sanitized := SanitizeStack(rawStack, "/Users/nobody")
	if strings.Contains(sanitized, "secret_access_token_12345") {
		t.Errorf("expected ya29 token to be redacted, got:\n%s", sanitized)
	}
	if strings.Contains(sanitized, "secretGitHubToken987654321") {
		t.Errorf("expected GitHub token to be redacted, got:\n%s", sanitized)
	}
	if strings.Contains(sanitized, "sk-proj-supersecretkey1234567890") {
		t.Errorf("expected secret_key to be redacted, got:\n%s", sanitized)
	}
	if !strings.Contains(sanitized, "[REDACTED]") {
		t.Errorf("expected [REDACTED] placeholder in sanitized output")
	}
}

func TestBuildGitHubIssueURL(t *testing.T) {
	title := "Crash: runtime error: invalid memory address"
	body := "### Details\nVersion: dev\nOS: " + runtime.GOOS

	url := BuildGitHubIssueURL(title, body)
	if !strings.HasPrefix(url, "https://github.com/adrijshikhar/aim/issues/new?") {
		t.Errorf("expected URL to start with GitHub issues URL, got %q", url)
	}
	if !strings.Contains(url, "title=Crash%3A+runtime+error") && !strings.Contains(url, "title=Crash%3A%20runtime%20error") {
		t.Errorf("expected URL-encoded title in URL, got %q", url)
	}
	if !strings.Contains(url, "body=") {
		t.Errorf("expected URL to have body param, got %q", url)
	}
}

func TestHandlePanic_GeneratesValidReportAndURL(t *testing.T) {
	home := "/Users/developer"
	recovered := "runtime error: index out of range [5] with length 2"
	rawStack := []byte(`goroutine 1 [running]:
main.run()
	/Users/developer/Projects/aim/cmd/aim/root.go:99 +0x12
`)

	issueURL, report := HandlePanic(recovered, rawStack, home, "v0.12.1", "1ba5e5a")
	if issueURL == "" {
		t.Errorf("expected non-empty issue URL")
	}
	if !strings.Contains(report, "runtime error: index out of range") {
		t.Errorf("expected report to contain panic message, got:\n%s", report)
	}
	if strings.Contains(report, "/Users/developer") {
		t.Errorf("expected home dir to be scrubbed from report, got:\n%s", report)
	}
	if !strings.Contains(report, "v0.12.1") || !strings.Contains(report, "1ba5e5a") {
		t.Errorf("expected version and commit in report")
	}
}
