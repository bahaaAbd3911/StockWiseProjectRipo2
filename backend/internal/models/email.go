package models

import "time"

// Email statuses in the admin view of the email queue (computed, not stored).
const (
	EmailPending = "pending" // waiting to be sent, or waiting for its next retry
	EmailFailed  = "failed"  // gave up after the maximum number of attempts
	EmailSent    = "sent"
)

func IsEmailStatus(s string) bool { return s == EmailPending || s == EmailFailed || s == EmailSent }

// Email is one row of the email queue (email_outbox).
type Email struct {
	ID            int64      `json:"id"`
	Kind          string     `json:"kind"`
	To            string     `json:"to"`
	Subject       string     `json:"subject"`
	Body          string     `json:"body,omitempty"` // only in the detail view
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastError     *string    `json:"last_error"`
	NextAttemptAt *time.Time `json:"next_attempt_at"` // null once sent or failed
	SentAt        *time.Time `json:"sent_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type EmailFilter struct {
	Status string // pending | failed | sent ("" = all)
	Kind   string // order_placed, debt_reminder, ...
	To     string // part of the address, case-insensitive
	Limit  int
	Offset int
}

// EmailSummary is the health of the queue at a glance.
type EmailSummary struct {
	Pending         int        `json:"pending"`
	Failed          int        `json:"failed"`
	SentLast24h     int        `json:"sent_last_24h"`
	OldestPendingAt *time.Time `json:"oldest_pending_at"` // a value far in the past means delivery is stuck
}
