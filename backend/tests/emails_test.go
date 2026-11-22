package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"ecommerce-backend/internal/notify"
)

// queueRow inserts an email directly with the given state (so the test does
// not have to wait through ten real retries).
func queueRow(t *testing.T, to, state string) string {
	t.Helper()
	var query string
	switch state {
	case "failed": // gave up after 10 attempts
		query = `INSERT INTO email_outbox (kind, to_address, subject, body, attempts, last_error)
			VALUES ('order_status', $1, 'Order #1: shipped', 'full text', 10, '550 mailbox does not exist') RETURNING id`
	case "later": // pending, next retry in an hour
		query = `INSERT INTO email_outbox (kind, to_address, subject, body, attempts, last_error, next_attempt_at)
			VALUES ('debt_reminder', $1, 'Friendly reminder', 'full text', 2, 'timeout', NOW() + INTERVAL '1 hour') RETURNING id`
	case "old-sent":
		query = `INSERT INTO email_outbox (kind, to_address, subject, body, attempts, sent_at)
			VALUES ('order_placed', $1, 'Order #1 received', 'full text', 1, NOW() - INTERVAL '40 days') RETURNING id`
	}
	var id string
	if err := testDB.QueryRow(query, to).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestAdminEmailQueue(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "queue-admin@example.com")
	register(t, "queue-customer@example.com")
	customer := login(t, "queue-customer@example.com", "password123")

	failed := queueRow(t, "queue-failed@example.com", "failed")
	later := queueRow(t, "queue-later@example.com", "later")

	// Access and filters.
	must(t, 403, "GET", "/api/admin/emails", customer, nil)
	must(t, 401, "GET", "/api/admin/emails/summary", "", nil)
	must(t, 400, "GET", "/api/admin/emails?status=lost", admin, nil)

	sum := obj(must(t, 200, "GET", "/api/admin/emails/summary", admin, nil))
	if str(sum["failed"]) == "0" || str(sum["pending"]) == "0" || sum["oldest_pending_at"] == nil {
		t.Fatalf("summary should count the failed and the pending email: %v", sum)
	}

	page := obj(must(t, 200, "GET", "/api/admin/emails?status=failed&to=QUEUE-FAILED", admin, nil))
	items := arr(page["items"])
	if str(page["total"]) != "1" || len(items) != 1 {
		t.Fatalf("filter by status and address (case-insensitive): %v", page)
	}
	row := obj(items[0])
	if row["status"] != "failed" || row["next_attempt_at"] != nil || !strings.Contains(str(row["last_error"]), "mailbox") || row["body"] != nil {
		t.Fatalf("failed row should show why, have no next attempt, and no body in the list: %v", row)
	}
	if obj(must(t, 200, "GET", "/api/admin/emails/"+later, admin, nil))["status"] != "pending" {
		t.Fatal("an email waiting for its next retry is pending")
	}
	if detail := obj(must(t, 200, "GET", "/api/admin/emails/"+failed, admin, nil)); detail["body"] != "full text" {
		t.Fatalf("the detail view shows the full text: %v", detail)
	}
	must(t, 404, "GET", "/api/admin/emails/999999", admin, nil)

	// Retry one failed email: it is delivered and becomes "sent".
	retried := obj(must(t, 200, "POST", "/api/admin/emails/"+failed+"/retry", admin, nil))
	if retried["status"] != "pending" || str(retried["attempts"]) != "0" {
		t.Fatalf("a retried email starts over: %v", retried)
	}
	waitForEmails(t, "queue-failed@example.com", "shipped", 1)
	waitStatus := func(id, want string) {
		t.Helper()
		for i := 0; i < 100; i++ {
			if obj(must(t, 200, "GET", "/api/admin/emails/"+id, admin, nil))["status"] == want {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("email %s never became %s", id, want)
	}
	waitStatus(failed, "sent")
	must(t, 409, "POST", "/api/admin/emails/"+failed+"/retry", admin, nil) // already sent
	must(t, 409, "DELETE", "/api/admin/emails/"+failed, admin, nil)
	must(t, 404, "POST", "/api/admin/emails/999999/retry", admin, nil)

	// Cancel a pending email.
	must(t, 200, "DELETE", "/api/admin/emails/"+later, admin, nil)
	must(t, 404, "GET", "/api/admin/emails/"+later, admin, nil)

	// Retry every failed email at once.
	queueRow(t, "queue-bulk1@example.com", "failed")
	queueRow(t, "queue-bulk2@example.com", "failed")
	msg := obj(must(t, 200, "POST", "/api/admin/emails/retry-failed", admin, nil))
	if !strings.HasPrefix(str(msg["message"]), "2 failed") {
		t.Fatalf("both failed emails should be retried: %v", msg)
	}
	waitForEmails(t, "queue-bulk1@example.com", "shipped", 1)
	waitForEmails(t, "queue-bulk2@example.com", "shipped", 1)

	// Delivered emails older than the retention period are removed.
	old := queueRow(t, "queue-old@example.com", "old-sent")
	n, err := notify.NewOutbox(testDB).DeleteOldSent(context.Background(), 30)
	if err != nil || n < 1 {
		t.Fatalf("old sent email should be deleted: n=%d err=%v", n, err)
	}
	must(t, 404, "GET", "/api/admin/emails/"+old, admin, nil)
	must(t, 200, "GET", "/api/admin/emails/"+failed, admin, nil) // recent sent emails stay
}
