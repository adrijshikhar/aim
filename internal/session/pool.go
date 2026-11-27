package session

import (
	"context"
	"database/sql"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // Register pure-Go sqlite driver for read-only session queries.
)

// OpenReadOnlySQLite opens a SQLite database with query_only mode and a bounded busy timeout.
func OpenReadOnlySQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		absPath = filepath.Clean(dbPath)
	}
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(absPath),
		RawQuery: "mode=ro&_pragma=query_only(ON)&_pragma=busy_timeout(2000)",
	}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

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
