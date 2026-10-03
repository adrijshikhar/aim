package session

import (
	"context"
	"os/exec"
	"strings"
)

var sqliteLimiter = make(chan struct{}, 6) // Max 6 concurrent sqlite subprocesses

// RunBoundedSQLite executes an SQLite query via CLI with bounded concurrency (max 6 subprocesses).
func RunBoundedSQLite(ctx context.Context, dbPath string, query string) ([]byte, error) {
	return RunBoundedSQLiteWithBin(ctx, "sqlite3", dbPath, query)
}

// RunBoundedSQLiteWithBin executes an SQLite query with a custom sqlite binary and optional extra CLI flags.
func RunBoundedSQLiteWithBin(ctx context.Context, sqliteBin, dbPath, query string, extraFlags ...string) ([]byte, error) {
	if sqliteBin == "" {
		sqliteBin = "sqlite3"
	}

	select {
	case sqliteLimiter <- struct{}{}:
		defer func() { <-sqliteLimiter }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, extraFlags...)
	cmdArgs = append(cmdArgs, dbPath)

	cmd := exec.CommandContext(ctx, sqliteBin, cmdArgs...)
	if query != "" {
		cmd.Stdin = strings.NewReader(query)
	}
	return cmd.Output()
}
