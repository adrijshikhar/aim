package agents

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/usage"
)

type mockAdapter struct {
	name        string
	displayName string
	aliases     []string
	binName     string
}

func (m *mockAdapter) Name() string        { return m.name }
func (m *mockAdapter) DisplayName() string { return m.displayName }
func (m *mockAdapter) Aliases() []string   { return m.aliases }
func (m *mockAdapter) BinaryName() string  { return m.binName }
func (m *mockAdapter) HasCredentials(profileDir string) bool {
	return false
}
func (m *mockAdapter) Login(ctx context.Context, profileName, profileDir string) error {
	return nil
}
func (m *mockAdapter) PrepareEnv(profileName, profileDir string) (LaunchEnv, error) {
	return LaunchEnv{}, nil
}
func (m *mockAdapter) Doctor(ctx context.Context, profileName, profileDir string) []DiagnosticResult {
	return nil
}
func (m *mockAdapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{
		Agent:     m.name,
		Profile:   profileName,
		Status:    usage.StatusOK,
		FetchedAt: time.Now(),
	}, nil
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Get("agy"); err == nil {
		t.Fatalf("expected error on empty registry")
	}

	mock := &mockAdapter{name: "agy", aliases: []string{"antigravity"}}
	r.Register(mock)

	got, err := r.Get("agy")
	if err != nil || got.Name() != "agy" {
		t.Errorf("failed to retrieve registered adapter: %v", err)
	}

	gotAlias, err := r.Get("antigravity")
	if err != nil || gotAlias.Name() != "agy" {
		t.Errorf("failed to retrieve adapter via alias: %v", err)
	}

	// Case-insensitivity check
	gotCase, err := r.Get("AGY")
	if err != nil || gotCase.Name() != "agy" {
		t.Errorf("failed to retrieve adapter with case-insensitivity: %v", err)
	}

	gotAliasCase, err := r.Get("AntiGravity")
	if err != nil || gotAliasCase.Name() != "agy" {
		t.Errorf("failed to retrieve adapter alias with case-insensitivity: %v", err)
	}
}

func TestRegistryAll(t *testing.T) {
	r := NewRegistry()
	mock1 := &mockAdapter{name: "agy", aliases: []string{"antigravity"}}
	mock2 := &mockAdapter{name: "gemini", aliases: []string{"gemini-cli"}}
	r.Register(mock1)
	r.Register(mock2)

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 adapters in All(), got %d", len(all))
	}
}

func TestRegistryConcurrent(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			adapter := &mockAdapter{
				name:    fmt.Sprintf("agent-%d", idx),
				aliases: []string{fmt.Sprintf("alias-%d", idx)},
			}
			r.Register(adapter)
		}(i)
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = r.Get(fmt.Sprintf("agent-%d", idx))
			_ = r.All()
		}(i)
	}

	wg.Wait()
}

func TestNewRegistryIsIndependent(t *testing.T) {
	first, second := NewRegistry(), NewRegistry()
	first.Register(&mockAdapter{name: "first"})
	if _, err := second.Get("first"); err == nil {
		t.Fatal("registries must not share registrations")
	}
}
