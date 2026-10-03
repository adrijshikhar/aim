package session

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunBoundedSQLite_Success(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")

	// Create table and sample record
	initQuery := "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT); INSERT INTO items (name) VALUES ('aim-scalability');"
	if err := exec.Command(sqliteBin, dbPath, initQuery).Run(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := RunBoundedSQLite(ctx, dbPath, "SELECT name FROM items WHERE id = 1;")
	if err != nil {
		t.Fatalf("RunBoundedSQLite failed: %v", err)
	}

	if strings.TrimSpace(string(out)) != "aim-scalability" {
		t.Errorf("expected 'aim-scalability', got %q", string(out))
	}
}

func TestRunBoundedSQLite_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := RunBoundedSQLite(ctx, "/path/to/nonexistent.db", "SELECT 1;")
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestRunBoundedSQLite_ConcurrencyLimiter(t *testing.T) {
	sqliteBin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("sqlite3 binary not available in PATH")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_concurrent.sqlite")

	if err := exec.Command(sqliteBin, dbPath, "CREATE TABLE counter (val INT); INSERT INTO counter VALUES (1);").Run(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	const totalWorkers = 25
	var wg sync.WaitGroup
	var activeSubprocesses int64
	var maxActiveObserved int64
	errChan := make(chan error, totalWorkers)

	for i := 0; i < totalWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cur := atomic.AddInt64(&activeSubprocesses, 1)
			for {
				maxObs := atomic.LoadInt64(&maxActiveObserved)
				if cur <= maxObs || atomic.CompareAndSwapInt64(&maxActiveObserved, maxObs, cur) {
					break
				}
			}

			out, err := RunBoundedSQLiteWithBin(ctx, sqliteBin, dbPath, "SELECT val FROM counter;")
			atomic.AddInt64(&activeSubprocesses, -1)

			if err != nil {
				errChan <- err
				return
			}
			if strings.TrimSpace(string(out)) != "1" {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent worker failed: %v", err)
	}
}
