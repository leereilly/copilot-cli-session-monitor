package session

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

// Reader reads Copilot CLI session data from the local filesystem.
type Reader struct {
	copilotDir string
	db         *sql.DB
}

// NewReader creates a Reader that looks for Copilot data in the given directory.
// Pass "" to use the default ~/.copilot/ location.
func NewReader(copilotDir string) (*Reader, error) {
	if copilotDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolving home directory: %w", err)
		}
		copilotDir = filepath.Join(home, ".copilot")
	}

	dbPath := filepath.Join(copilotDir, "session-store.db")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("session database not found at %s: %w", dbPath, err)
	}

	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening session-store.db: %w", err)
	}

	// Verify the connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to session-store.db: %w", err)
	}

	return &Reader{copilotDir: copilotDir, db: db}, nil
}

// Close releases the database connection.
func (r *Reader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// ReadSessions returns all active Copilot sessions sorted by last activity.
func (r *Reader) ReadSessions() ([]CopilotSession, error) {
	// Scan lock files first to find which sessions are active
	activePIDs, err := r.scanLockFiles()
	if err != nil {
		return nil, err
	}

	if len(activePIDs) == 0 {
		return nil, nil
	}

	// Query only the sessions that have active lock files
	sessions, err := r.querySessionsByIDs(activePIDs)
	if err != nil {
		return nil, err
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// querySessionsByIDs fetches session metadata only for the given session IDs.
func (r *Reader) querySessionsByIDs(activePIDs map[string]int) ([]CopilotSession, error) {
	ids := make([]string, 0, len(activePIDs))
	for id := range activePIDs {
		ids = append(ids, id)
	}

	// Build parameterized query with placeholders
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, COALESCE(cwd, ''), COALESCE(repository, ''),
		       COALESCE(summary, ''), COALESCE(updated_at, '')
		FROM sessions
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer rows.Close()

	var sessions []CopilotSession
	for rows.Next() {
		var s CopilotSession
		var updatedAtStr string
		if err := rows.Scan(&s.ID, &s.CWD, &s.Repository, &s.Summary, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		s.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAtStr)
		s.Status = StatusActive
		s.PID = activePIDs[s.ID]
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// scanLockFiles finds active session lock files and returns a map of sessionID → PID
// for sessions whose owning process is still alive.
func (r *Reader) scanLockFiles() (map[string]int, error) {
	pattern := filepath.Join(r.copilotDir, "session-state", "*", "inuse.*.lock")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("globbing lock files: %w", err)
	}

	activePIDs := make(map[string]int, len(matches))
	for _, lockPath := range matches {
		sessionID := filepath.Base(filepath.Dir(lockPath))

		filename := filepath.Base(lockPath)
		parts := strings.Split(filename, ".")
		if len(parts) != 3 {
			continue
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		if isProcessAlive(pid) {
			activePIDs[sessionID] = pid
		}
	}
	return activePIDs, nil
}

// isProcessAlive checks if a process with the given PID exists.
func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
