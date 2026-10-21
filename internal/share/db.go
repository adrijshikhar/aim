package share

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type ShareDB struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

// NewShareDB initializes and runs migrations for the SQLite database at <aimDir>/share.db.
func NewShareDB(aimDir string) (*ShareDB, error) {
	if err := os.MkdirAll(aimDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create share directory: %w", err)
	}

	dbPath := filepath.Join(aimDir, "share.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize SQLite for concurrent access
	db.SetMaxOpenConns(1) // Single open connection prevents SQLITE_BUSY in multi-goroutine CLI/web

	sdb := &ShareDB{
		db:     db,
		dbPath: dbPath,
	}

	if err := sdb.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	_ = os.Chmod(dbPath, 0600)

	// Migrate existing JSON audit logs if table is empty
	sdb.migrateLegacyAuditLogs(aimDir)

	return sdb, nil
}

func (s *ShareDB) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := s.db.ExecContext(ctx, p); err != nil {
			return err
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS share_requests (
		id TEXT PRIMARY KEY,
		direction TEXT NOT NULL,
		peer_id TEXT NOT NULL,
		peer_ip TEXT NOT NULL,
		profile_name TEXT NOT NULL,
		agent TEXT NOT NULL,
		requested_hours INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL,
		lease_id TEXT,
		error_message TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_requests_dir_status ON share_requests(direction, status);
	CREATE INDEX IF NOT EXISTS idx_requests_expires ON share_requests(expires_at);

	CREATE TABLE IF NOT EXISTS share_audit_logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		action TEXT NOT NULL,
		actor TEXT NOT NULL,
		profile TEXT,
		agent TEXT,
		details TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON share_audit_logs(timestamp DESC);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *ShareDB) migrateLegacyAuditLogs(aimDir string) {
	legacyFile := filepath.Join(aimDir, "share_audit.json")
	data, err := os.ReadFile(legacyFile)
	if err != nil || len(data) == 0 {
		return
	}

	var legacyLogs []AuditEntry
	if err := json.Unmarshal(data, &legacyLogs); err != nil || len(legacyLogs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM share_audit_logs").Scan(&count); err != nil || count > 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO share_audit_logs (id, timestamp, action, actor, profile, agent, details)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, entry := range legacyLogs {
		if entry.ID == "" {
			entry.ID = fmt.Sprintf("aud_%d", entry.Timestamp.UnixNano())
		}
		_, _ = stmt.ExecContext(ctx, entry.ID, entry.Timestamp.Format(time.RFC3339Nano), entry.Action, entry.Actor, entry.Profile, entry.Agent, entry.Details)
	}

	_ = tx.Commit()
}

// Close closes the database connection.
func (s *ShareDB) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveRequest upserts a lease request (incoming or outgoing).
func (s *ShareDB) SaveRequest(ctx context.Context, req ShareRequestRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.ID == "" {
		return errors.New("request id cannot be empty")
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = req.CreatedAt.Add(30 * time.Minute)
	}
	if req.Status == "" {
		req.Status = "pending"
	}

	query := `
	INSERT INTO share_requests (
		id, direction, peer_id, peer_ip, profile_name, agent,
		requested_hours, status, created_at, expires_at, lease_id, error_message
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		direction = excluded.direction,
		peer_id = excluded.peer_id,
		peer_ip = excluded.peer_ip,
		profile_name = excluded.profile_name,
		agent = excluded.agent,
		requested_hours = excluded.requested_hours,
		status = excluded.status,
		expires_at = excluded.expires_at,
		lease_id = excluded.lease_id,
		error_message = excluded.error_message
	`

	_, err := s.db.ExecContext(ctx, query,
		req.ID, req.Direction, req.PeerID, req.PeerIP, req.ProfileName, req.Agent,
		req.RequestedHours, req.Status,
		req.CreatedAt.Format(time.RFC3339Nano),
		req.ExpiresAt.Format(time.RFC3339Nano),
		req.LeaseID, req.ErrorMessage,
	)
	return err
}

// GetRequest retrieves a single request record by ID.
func (s *ShareDB) GetRequest(ctx context.Context, id string) (*ShareRequestRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, direction, peer_id, peer_ip, profile_name, agent,
	       requested_hours, status, created_at, expires_at, lease_id, error_message
	FROM share_requests
	WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var (
		r        ShareRequestRecord
		created  string
		expires  string
		leaseID  sql.NullString
		errorMsg sql.NullString
	)

	err := row.Scan(
		&r.ID, &r.Direction, &r.PeerID, &r.PeerIP, &r.ProfileName, &r.Agent,
		&r.RequestedHours, &r.Status, &created, &expires, &leaseID, &errorMsg,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
		r.CreatedAt = t
	} else if t, err := time.Parse("2006-01-02 15:04:05", created); err == nil {
		r.CreatedAt = t
	}

	if t, err := time.Parse(time.RFC3339Nano, expires); err == nil {
		r.ExpiresAt = t
	} else if t, err := time.Parse("2006-01-02 15:04:05", expires); err == nil {
		r.ExpiresAt = t
	}

	if leaseID.Valid {
		r.LeaseID = leaseID.String
	}
	if errorMsg.Valid {
		r.ErrorMessage = errorMsg.String
	}

	return &r, nil
}

// GetPendingIncomingRequests returns all active incoming requests awaiting manual host approval.
func (s *ShareDB) GetPendingIncomingRequests(ctx context.Context) ([]ShareRequestRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, direction, peer_id, peer_ip, profile_name, agent,
	       requested_hours, status, created_at, expires_at, lease_id, error_message
	FROM share_requests
	WHERE direction = 'incoming' AND status = 'pending'
	ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ShareRequestRecord
	now := time.Now()

	for rows.Next() {
		var (
			r        ShareRequestRecord
			created  string
			expires  string
			leaseID  sql.NullString
			errorMsg sql.NullString
		)
		if err := rows.Scan(
			&r.ID, &r.Direction, &r.PeerID, &r.PeerIP, &r.ProfileName, &r.Agent,
			&r.RequestedHours, &r.Status, &created, &expires, &leaseID, &errorMsg,
		); err != nil {
			continue
		}

		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			r.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339Nano, expires); err == nil {
			r.ExpiresAt = t
		}

		if leaseID.Valid {
			r.LeaseID = leaseID.String
		}
		if errorMsg.Valid {
			r.ErrorMessage = errorMsg.String
		}

		if !r.ExpiresAt.IsZero() && now.After(r.ExpiresAt) {
			continue // skip expired requests in pending view
		}
		results = append(results, r)
	}

	return results, nil
}

// GetPendingOutgoingRequests returns all active outgoing requests sent to peers.
func (s *ShareDB) GetPendingOutgoingRequests(ctx context.Context) ([]ShareRequestRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, direction, peer_id, peer_ip, profile_name, agent,
	       requested_hours, status, created_at, expires_at, lease_id, error_message
	FROM share_requests
	WHERE direction = 'outgoing' AND status = 'pending'
	ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ShareRequestRecord
	now := time.Now()

	for rows.Next() {
		var (
			r        ShareRequestRecord
			created  string
			expires  string
			leaseID  sql.NullString
			errorMsg sql.NullString
		)
		if err := rows.Scan(
			&r.ID, &r.Direction, &r.PeerID, &r.PeerIP, &r.ProfileName, &r.Agent,
			&r.RequestedHours, &r.Status, &created, &expires, &leaseID, &errorMsg,
		); err != nil {
			continue
		}

		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			r.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339Nano, expires); err == nil {
			r.ExpiresAt = t
		}

		if leaseID.Valid {
			r.LeaseID = leaseID.String
		}
		if errorMsg.Valid {
			r.ErrorMessage = errorMsg.String
		}

		if !r.ExpiresAt.IsZero() && now.After(r.ExpiresAt) {
			continue
		}
		results = append(results, r)
	}

	return results, nil
}

// UpdateRequestStatus updates status, lease ID, and error message for a request.
func (s *ShareDB) UpdateRequestStatus(ctx context.Context, id string, status string, leaseID string, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	UPDATE share_requests
	SET status = ?, lease_id = ?, error_message = ?
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, status, leaseID, errMsg, id)
	return err
}

// CancelRequest marks a request as cancelled.
func (s *ShareDB) CancelRequest(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	UPDATE share_requests
	SET status = 'cancelled'
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// AppendAudit records an audit entry into SQLite.
func (s *ShareDB) AppendAudit(ctx context.Context, entry AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("aud_%d", time.Now().UnixNano())
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	query := `
	INSERT INTO share_audit_logs (id, timestamp, action, actor, profile, agent, details)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		entry.ID,
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.Action,
		entry.Actor,
		entry.Profile,
		entry.Agent,
		entry.Details,
	)
	return err
}

// GetAuditLogs returns the most recent audit log entries, up to limit.
func (s *ShareDB) GetAuditLogs(ctx context.Context, limit int) ([]AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT id, timestamp, action, actor, profile, agent, details
	FROM share_audit_logs
	ORDER BY timestamp DESC
	LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditEntry
	for rows.Next() {
		var (
			e       AuditEntry
			ts      string
			profile sql.NullString
			agent   sql.NullString
		)
		if err := rows.Scan(&e.ID, &ts, &e.Action, &e.Actor, &profile, &agent, &e.Details); err != nil {
			continue
		}
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			e.Timestamp = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
			e.Timestamp = t
		}
		if profile.Valid {
			e.Profile = profile.String
		}
		if agent.Valid {
			e.Agent = agent.String
		}
		logs = append(logs, e)
	}

	return logs, nil
}
