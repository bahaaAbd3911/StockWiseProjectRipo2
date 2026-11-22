package notify

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/mail"

	"github.com/shopspring/decimal"
)

// Outbox queues emails in the email_outbox table. Queueing never sends
// anything itself, so a slow or broken mail provider can never slow down or
// fail a checkout: the Worker delivers in the background.
type Outbox struct {
	db *sql.DB
}

func NewOutbox(db *sql.DB) *Outbox { return &Outbox{db: db} }

// Enqueue stores one email for delivery.
func (o *Outbox) Enqueue(ctx context.Context, kind string, m mail.Message) error {
	return enqueue(ctx, o.db, kind, m)
}

// EnqueueOrLog is Enqueue for "nice to have" emails: a failure is logged and
// never turns a successful action (an order, a payment) into an error.
func (o *Outbox) EnqueueOrLog(ctx context.Context, kind string, m mail.Message) {
	if o == nil || m.To == "" { // nil outbox: notifications off
		return
	}
	if err := o.Enqueue(context.WithoutCancel(ctx), kind, m); err != nil {
		slog.Error("could not queue email", "kind", kind, "error", err)
	}
}

func enqueue(ctx context.Context, q database.Querier, kind string, m mail.Message) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO email_outbox (kind, to_address, subject, body) VALUES ($1, $2, $3, $4)`,
		kind, m.To, m.Subject, m.Text)
	return err
}

// ---------- worker ----------

const (
	batchSize   = 20
	maxAttempts = 10 // about 2 days of retries with the backoff below
	sendTimeout = 30 * time.Second
)

// WorkerOptions configure the background delivery.
type WorkerOptions struct {
	Interval time.Duration // how often the queue is checked (default 5s)
	// DebtReminderEvery: customers with an open balance get a reminder this
	// long after their last charge, then again every period. 0 = off.
	DebtReminderEvery time.Duration
	Templates         Templates
}

// Worker delivers queued emails and queues debt reminders.
type Worker struct {
	db     *sql.DB
	mailer mail.Mailer
	opts   WorkerOptions

	stop chan struct{}
	done sync.WaitGroup
	once sync.Once
}

func NewWorker(db *sql.DB, mailer mail.Mailer, opts WorkerOptions) *Worker {
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}
	return &Worker{db: db, mailer: mailer, opts: opts, stop: make(chan struct{})}
}

// Start runs the worker until Stop.
func (w *Worker) Start() {
	w.done.Add(1)
	go func() {
		defer w.done.Done()
		ticker := time.NewTicker(w.opts.Interval)
		defer ticker.Stop()
		var lastReminders time.Time
		for {
			ctx := context.Background()
			// Debt reminders are checked at most every 10 minutes: they are
			// counted in days, there is no point asking the database more often.
			if w.opts.DebtReminderEvery > 0 && time.Since(lastReminders) >= min(10*time.Minute, w.opts.DebtReminderEvery) {
				lastReminders = time.Now()
				if _, err := w.QueueDebtReminders(ctx); err != nil {
					slog.Error("queue debt reminders", "error", err)
				}
			}
			if _, err := w.RunOnce(ctx); err != nil {
				slog.Error("deliver queued emails", "error", err)
			}
			select {
			case <-w.stop:
				return
			case <-ticker.C:
			}
		}
	}()
}

// Stop ends the worker after the batch in progress. Unsent emails stay queued
// and go out after the next start.
func (w *Worker) Stop() {
	w.once.Do(func() { close(w.stop) })
	w.done.Wait()
}

// RunOnce delivers up to one batch of due emails and returns how many were
// sent. The rows stay locked (FOR UPDATE SKIP LOCKED) while they are sent, so
// a second API instance picks different rows and nothing goes out twice.
func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	sent := 0
	err := database.WithTx(ctx, w.db, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, to_address, subject, body, attempts FROM email_outbox
			WHERE sent_at IS NULL AND attempts < $1 AND next_attempt_at <= NOW()
			ORDER BY id
			LIMIT $2
			FOR UPDATE SKIP LOCKED`, maxAttempts, batchSize)
		if err != nil {
			return err
		}
		type job struct {
			id       int64
			msg      mail.Message
			attempts int
		}
		var jobs []job
		for rows.Next() {
			var j job
			if err := rows.Scan(&j.id, &j.msg.To, &j.msg.Subject, &j.msg.Text, &j.attempts); err != nil {
				rows.Close()
				return err
			}
			jobs = append(jobs, j)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		for _, j := range jobs {
			sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
			sendErr := w.mailer.Send(sendCtx, j.msg)
			cancel()
			if sendErr == nil {
				sent++
				if _, err := tx.ExecContext(ctx,
					`UPDATE email_outbox SET sent_at = NOW(), attempts = attempts + 1, last_error = NULL WHERE id = $1`, j.id); err != nil {
					return err
				}
				continue
			}
			attempts := j.attempts + 1
			slog.Warn("email delivery failed", "id", j.id, "attempt", attempts, "error", sendErr)
			if _, err := tx.ExecContext(ctx, `
				UPDATE email_outbox SET attempts = $2, last_error = $3, next_attempt_at = NOW() + make_interval(secs => $4)
				WHERE id = $1`, j.id, attempts, sendErr.Error(), Backoff(attempts).Seconds()); err != nil {
				return err
			}
		}
		return nil
	})
	return sent, err
}

// Backoff is the wait before retry number `attempts`: 1, 2, 4, 8 ... minutes,
// at most 6 hours.
func Backoff(attempts int) time.Duration {
	d := time.Minute
	for i := 1; i < attempts && d < 6*time.Hour; i++ {
		d *= 2
	}
	return min(d, 6*time.Hour)
}

// QueueDebtReminders queues a reminder for every active customer who owes
// money and has not been reminded (or charged) for DebtReminderEvery. The
// reminder and users.debt_reminded_at are written in the same transaction, so
// a customer is never reminded twice for the same period.
func (w *Worker) QueueDebtReminders(ctx context.Context) (int, error) {
	if w.opts.DebtReminderEvery <= 0 {
		return 0, nil
	}
	queued := 0
	err := database.WithTx(ctx, w.db, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT u.id, u.email, u.first_name, b.balance
			FROM users u
			CROSS JOIN LATERAL (
				SELECT COALESCE(SUM(t.amount), 0) AS balance,
				       MAX(t.created_at) FILTER (WHERE t.amount > 0) AS last_charge
				FROM customer_debt_transactions t WHERE t.user_id = u.id
			) b
			WHERE u.is_active
			  AND b.balance > 0
			  AND COALESCE(u.debt_reminded_at, b.last_charge) <= NOW() - make_interval(secs => $1)
			ORDER BY u.id
			FOR UPDATE OF u SKIP LOCKED`, w.opts.DebtReminderEvery.Seconds())
		if err != nil {
			return err
		}
		type due struct {
			id      int64
			to      Recipient
			balance decimal.Decimal
		}
		var list []due
		for rows.Next() {
			var d due
			if err := rows.Scan(&d.id, &d.to.Email, &d.to.FirstName, &d.balance); err != nil {
				rows.Close()
				return err
			}
			list = append(list, d)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		for _, d := range list {
			if err := enqueue(ctx, tx, KindDebtReminder, w.opts.Templates.DebtReminder(d.to, d.balance)); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE users SET debt_reminded_at = NOW() WHERE id = $1`, d.id); err != nil {
				return err
			}
			queued++
		}
		return nil
	})
	if errors.Is(err, context.Canceled) {
		return queued, nil
	}
	return queued, err
}
