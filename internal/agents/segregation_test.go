package agents_test

import (
	"context"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/agents/claude"
	"github.com/aim-cli/aim/internal/agents/codex"
	"github.com/aim-cli/aim/internal/agents/cursor"
	"github.com/aim-cli/aim/internal/agents/gemini"
)

// minimalStubAdapter only implements AgentAdapter, omitting Authenticator, Diagnostician, etc.
type minimalStubAdapter struct {
	agents.BaseAdapter
}

func (m *minimalStubAdapter) HasCredentials(profileDir string) bool { return false }
func (m *minimalStubAdapter) PrepareEnv(ctx context.Context, profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}

func TestInterfaceSegregation_MinimalAdapter(t *testing.T) {
	stub := &minimalStubAdapter{
		BaseAdapter: agents.NewBaseAdapter("stub", "Stub Agent", "stub", nil),
	}

	// Must implement AgentAdapter
	var _ agents.AgentAdapter = stub

	// Must NOT implement optional interfaces
	if _, ok := any(stub).(agents.Authenticator); ok {
		t.Fatal("minimal stub should not implement Authenticator")
	}
	if _, ok := any(stub).(agents.Diagnostician); ok {
		t.Fatal("minimal stub should not implement Diagnostician")
	}
	if _, ok := any(stub).(agents.UsageProvider); ok {
		t.Fatal("minimal stub should not implement UsageProvider")
	}
	if _, ok := any(stub).(agents.AccountInfoProvider); ok {
		t.Fatal("minimal stub should not implement AccountInfoProvider")
	}
	if _, ok := any(stub).(agents.CredentialSeeder); ok {
		t.Fatal("minimal stub should not implement CredentialSeeder")
	}
	if _, ok := any(stub).(agents.CredentialHarvester); ok {
		t.Fatal("minimal stub should not implement CredentialHarvester")
	}
	if _, ok := any(stub).(agents.CredentialPurger); ok {
		t.Fatal("minimal stub should not implement CredentialPurger")
	}
	if _, ok := any(stub).(agents.FullAdapter); ok {
		t.Fatal("minimal stub should not implement FullAdapter")
	}
}

func TestInterfaceSegregation_StandardAdaptersImplementFullAdapter(t *testing.T) {
	tests := []struct {
		name    string
		adapter agents.AgentAdapter
	}{
		{"agy", agy.NewAdapter()},
		{"claude", claude.NewAdapter()},
		{"codex", codex.NewAdapter()},
		{"gemini", gemini.NewAdapter()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify FullAdapter
			fa, ok := tc.adapter.(agents.FullAdapter)
			if !ok {
				t.Fatalf("expected adapter %s to implement FullAdapter", tc.name)
			}

			// Verify each segregated capability
			if _, ok := tc.adapter.(agents.Authenticator); !ok {
				t.Errorf("adapter %s should implement Authenticator", tc.name)
			}
			if _, ok := tc.adapter.(agents.Diagnostician); !ok {
				t.Errorf("adapter %s should implement Diagnostician", tc.name)
			}
			if _, ok := tc.adapter.(agents.UsageProvider); !ok {
				t.Errorf("adapter %s should implement UsageProvider", tc.name)
			}
			if _, ok := tc.adapter.(agents.AccountInfoProvider); !ok {
				t.Errorf("adapter %s should implement AccountInfoProvider", tc.name)
			}

			// Test GetAccountInfo on empty profile dir
			email, authMethod, projectID := fa.GetAccountInfo("")
			if email != "" || authMethod != "" || projectID != "" {
				t.Errorf("expected empty account info for empty dir, got %s, %s, %s", email, authMethod, projectID)
			}
		})
	}
}

func TestInterfaceSegregation_CursorAdapter(t *testing.T) {
	cur := cursor.NewAdapter()

	// Cursor implements core AgentAdapter
	var _ agents.AgentAdapter = cur

	// Cursor implements segregated interfaces
	if _, ok := any(cur).(agents.Diagnostician); !ok {
		t.Error("cursor adapter should implement Diagnostician")
	}
	if _, ok := any(cur).(agents.AccountInfoProvider); !ok {
		t.Error("cursor adapter should implement AccountInfoProvider")
	}
	if _, ok := any(cur).(agents.RunnableFlagsProvider); !ok {
		t.Error("cursor adapter should implement RunnableFlagsProvider")
	}
	if _, ok := any(cur).(agents.MCPConfigManager); !ok {
		t.Error("cursor adapter should implement MCPConfigManager")
	}

	// Cursor is an IDE agent and intentionally does not implement CLI-specific Authenticator or UsageProvider
	if _, ok := any(cur).(agents.Authenticator); ok {
		t.Error("cursor adapter should not implement Authenticator")
	}
	if _, ok := any(cur).(agents.UsageProvider); ok {
		t.Error("cursor adapter should not implement UsageProvider")
	}
}
