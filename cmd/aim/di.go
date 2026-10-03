package main

import (
	"context"
	"io"
	"os"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/session"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

type (
	sessionManagerCtxKey      struct{}
	interactiveCheckerCtxKey  struct{}
	tuiRunnerCtxKey           struct{}
	tuiSessionsRunnerCtxKey   struct{}
	interactiveTerminalCtxKey struct{}
	sessionPickerRunnerCtxKey struct{}
)

// WithSessionManager injects a session.Manager into the context.
func WithSessionManager(ctx context.Context, mgr *session.Manager) context.Context {
	return context.WithValue(ctx, sessionManagerCtxKey{}, mgr)
}

// SessionManagerFromContext retrieves an injected session.Manager from context, if any.
func SessionManagerFromContext(ctx context.Context) *session.Manager {
	if ctx == nil {
		return nil
	}
	if mgr, ok := ctx.Value(sessionManagerCtxKey{}).(*session.Manager); ok && mgr != nil {
		return mgr
	}
	return nil
}

// getSessionManager returns the session.Manager from context if present, or falls back to NewDefaultSessionManager().
func getSessionManager(ctx context.Context) *session.Manager {
	if mgr := SessionManagerFromContext(ctx); mgr != nil {
		return mgr
	}
	return NewDefaultSessionManager()
}

// WithInteractiveCheck injects an interactivity predicate into the context.
func WithInteractiveCheck(ctx context.Context, fn func(io.Reader) bool) context.Context {
	return context.WithValue(ctx, interactiveCheckerCtxKey{}, fn)
}

// InteractiveCheckFromContext retrieves an injected interactivity predicate from context, if any.
func InteractiveCheckFromContext(ctx context.Context) func(io.Reader) bool {
	if ctx == nil {
		return nil
	}
	if fn, ok := ctx.Value(interactiveCheckerCtxKey{}).(func(io.Reader) bool); ok && fn != nil {
		return fn
	}
	return nil
}

// checkInteractive returns whether r is interactive, consulting context, then defaultIsInteractive.
func checkInteractive(cmd *cobra.Command, r io.Reader) bool {
	if cmd != nil {
		if fn := InteractiveCheckFromContext(cmd.Context()); fn != nil {
			return fn(r)
		}
	}
	return defaultIsInteractive(r)
}

// WithTUIRunner injects a custom TUI runner into the context.
func WithTUIRunner(ctx context.Context, fn func(reg *agents.Registry, pm *profile.ProfileManager) int) context.Context {
	return context.WithValue(ctx, tuiRunnerCtxKey{}, fn)
}

// TUIRunnerFromContext retrieves an injected TUI runner from context, if any.
func TUIRunnerFromContext(ctx context.Context) func(reg *agents.Registry, pm *profile.ProfileManager) int {
	if ctx == nil {
		return nil
	}
	if fn, ok := ctx.Value(tuiRunnerCtxKey{}).(func(reg *agents.Registry, pm *profile.ProfileManager) int); ok && fn != nil {
		return fn
	}
	return nil
}

// WithTUISessionsRunner injects a custom sessions TUI runner into the context.
func WithTUISessionsRunner(ctx context.Context, fn func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int) context.Context {
	return context.WithValue(ctx, tuiSessionsRunnerCtxKey{}, fn)
}

// TUISessionsRunnerFromContext retrieves an injected sessions TUI runner from context, if any.
func TUISessionsRunnerFromContext(ctx context.Context) func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int {
	if ctx == nil {
		return nil
	}
	if fn, ok := ctx.Value(tuiSessionsRunnerCtxKey{}).(func(reg *agents.Registry, pm *profile.ProfileManager, initialAgent, profileFilter string, activeOnly bool) int); ok && fn != nil {
		return fn
	}
	return nil
}

// WithInteractiveTerminal injects an interactive terminal check into the context.
func WithInteractiveTerminal(ctx context.Context, fn func(cmd *cobra.Command) bool) context.Context {
	return context.WithValue(ctx, interactiveTerminalCtxKey{}, fn)
}

// InteractiveTerminalFromContext retrieves an injected interactive terminal check from context, if any.
func InteractiveTerminalFromContext(ctx context.Context) func(cmd *cobra.Command) bool {
	if ctx == nil {
		return nil
	}
	if fn, ok := ctx.Value(interactiveTerminalCtxKey{}).(func(cmd *cobra.Command) bool); ok && fn != nil {
		return fn
	}
	return nil
}

// WithSessionPickerRunner injects a custom session picker runner into the context.
func WithSessionPickerRunner(ctx context.Context, fn SessionPickerRunner) context.Context {
	return context.WithValue(ctx, sessionPickerRunnerCtxKey{}, fn)
}

// SessionPickerRunnerFromContext retrieves an injected session picker runner from context, if any.
func SessionPickerRunnerFromContext(ctx context.Context) SessionPickerRunner {
	if ctx == nil {
		return nil
	}
	if fn, ok := ctx.Value(sessionPickerRunnerCtxKey{}).(SessionPickerRunner); ok && fn != nil {
		return fn
	}
	return nil
}

// defaultIsInteractiveTerminal checks if cmd's stdout and stdin are both interactive terminals.
func defaultIsInteractiveTerminal(cmd *cobra.Command) bool {
	out := cmd.OutOrStdout()
	f, ok := out.(*os.File)
	if !ok || f != os.Stdout {
		return false
	}
	return (isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) &&
		(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
