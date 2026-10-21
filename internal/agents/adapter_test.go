package agents_test

import (
	"context"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

type fullMockAdapter struct {
	agents.BaseAdapter
}

func (m *fullMockAdapter) HasCredentials(profileDir string) bool { return true }
func (m *fullMockAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

func (m *fullMockAdapter) GetCachedUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{Profile: profileName}, nil
}

func (m *fullMockAdapter) GetLiveUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{Profile: profileName}, nil
}

func (m *fullMockAdapter) HarvestCredentials(ctx context.Context, profileName, profileDir string) error {
	return nil
}

func (m *fullMockAdapter) PurgeCredentials(ctx context.Context, profileName, profileDir string) error {
	return nil
}

func (m *fullMockAdapter) ResumeArgs(sessionID string, extraArgs []string) []string {
	return append([]string{"--resume", sessionID}, extraArgs...)
}

type unextendedMockAdapter struct {
	agents.BaseAdapter
}

func (m *unextendedMockAdapter) HasCredentials(profileDir string) bool { return false }
func (m *unextendedMockAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

func TestSegregatedInterfaces_Assertions(t *testing.T) {
	full := &fullMockAdapter{
		BaseAdapter: agents.NewBaseAdapter("full", "Full Mock", "mock", nil),
	}

	// Verify full implements AgentAdapter
	var _ agents.AgentAdapter = full

	// Verify type assertions for newly segregated interfaces
	cachedProvider, ok := any(full).(agents.CachedUsageProvider)
	if !ok {
		t.Fatal("expected full mock to implement CachedUsageProvider")
	}
	report, err := cachedProvider.GetCachedUsage(context.Background(), "test", "/tmp")
	if err != nil || report == nil || report.Profile != "test" {
		t.Fatalf("unexpected GetCachedUsage result: report=%v, err=%v", report, err)
	}

	liveProvider, ok := any(full).(agents.LiveUsageProvider)
	if !ok {
		t.Fatal("expected full mock to implement LiveUsageProvider")
	}
	report, err = liveProvider.GetLiveUsage(context.Background(), "test", "/tmp")
	if err != nil || report == nil || report.Profile != "test" {
		t.Fatalf("unexpected GetLiveUsage result: report=%v, err=%v", report, err)
	}

	harvester, ok := any(full).(agents.CredentialHarvester)
	if !ok {
		t.Fatal("expected full mock to implement CredentialHarvester")
	}
	if err := harvester.HarvestCredentials(context.Background(), "test", "/tmp"); err != nil {
		t.Fatalf("unexpected HarvestCredentials error: %v", err)
	}

	purger, ok := any(full).(agents.CredentialPurger)
	if !ok {
		t.Fatal("expected full mock to implement CredentialPurger")
	}
	if err := purger.PurgeCredentials(context.Background(), "test", "/tmp"); err != nil {
		t.Fatalf("unexpected PurgeCredentials error: %v", err)
	}

	resumer, ok := any(full).(agents.SessionResumer)
	if !ok {
		t.Fatal("expected full mock to implement SessionResumer")
	}
	args := resumer.ResumeArgs("sess-123", []string{"--debug"})
	if len(args) != 3 || args[0] != "--resume" || args[1] != "sess-123" || args[2] != "--debug" {
		t.Fatalf("unexpected ResumeArgs output: %v", args)
	}

	// Verify unextended mock does NOT implement any of them
	unextended := &unextendedMockAdapter{
		BaseAdapter: agents.NewBaseAdapter("unextended", "Unextended Mock", "mock", nil),
	}
	if _, ok := any(unextended).(agents.CachedUsageProvider); ok {
		t.Error("unextended mock should not implement CachedUsageProvider")
	}
	if _, ok := any(unextended).(agents.LiveUsageProvider); ok {
		t.Error("unextended mock should not implement LiveUsageProvider")
	}
	if _, ok := any(unextended).(agents.CredentialHarvester); ok {
		t.Error("unextended mock should not implement CredentialHarvester")
	}
	if _, ok := any(unextended).(agents.CredentialPurger); ok {
		t.Error("unextended mock should not implement CredentialPurger")
	}
	if _, ok := any(unextended).(agents.SessionResumer); ok {
		t.Error("unextended mock should not implement SessionResumer")
	}
}
