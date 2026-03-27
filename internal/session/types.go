package session

import "time"

// Status represents whether a Copilot session is currently active.
type Status string

const (
	// StatusActive indicates a session with a live owning process.
	StatusActive Status = "active"
)

// CopilotSession holds metadata about a single Copilot CLI session.
type CopilotSession struct {
	ID         string
	Summary    string
	CWD        string
	Repository string
	UpdatedAt  time.Time
	Status     Status
	PID        int
}
