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
	"ecommerce-backend/internal/tenant"

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

// enqueueFor queues an email for a given store (used in platform mode, where
// store_id has no default).
func enqueueFor(ctx context.Context, q database.Querier, storeID int64, kind string, m mail.Message) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO email_outbox (store_id, kind, to_address, from_name, subject, body) VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)`,
		storeID, kind, m.To, m.FromName, m.Subject, m.Text)
	return err
}

// enqueue queues an email for the current store (store_id defaults to it).
func enqueue(ctx context.Context, q database.Querier, kind string, m mail.Message) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO email_outbox (kind, to_address, from_name, subject, body) VALUES ($1, $2, NULLIF($3, ''), $4, $5)`,
		kind, m.To, m.FromName, m.Subject, m.Text)
	return err
}

// ---------- worker ----------

const (
	batchSize   = 20
	maxAttempts = 10 // about 2 days of retries with the backoff below
	// maxAttemptsSQL is maxAttempts for SQL text (admin.go); a test keeps them equal.
	maxAttemptsSQL = "10"
	sendTimeout    = 30 * time.Second
)

// WorkerOptions configure the background delivery.
type WorkerOptions struct {
	Interval time.Duration // how often the queue is checked (default 5s)
	// DebtReminderEvery: customers with an open balance get a reminder this
	// long after their last charge, then again every period. 0 = off.
	DebtReminderEvery time.Duration
	// Branding gives each reminder its store's name, links and currency.
	Branding *Branding
	// KeepSentDays: delivered emails are deleted after this many days (default 30).
	KeepSentDays int
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
	if opts.KeepSentDays <= 0 {
		opts.KeepSentDays = 30
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
		var lastReminders, lastCleanup time.Time
		for {
			ctx := tenant.AsPlatform(context.Background()) // the worker serves every store
			if time.Since(lastCleanup) >= time.Hour {
				lastCleanup = time.Now()
				if n, err := (&Outbox{db: w.db}).DeleteOldSent(ctx, w.opts.KeepSentDays); err != nil {
					slog.Error("clean up sent emails", "error", err)
				} else if n > 0 {
					slog.Info("deleted old sent emails", "count", n, "older_than_days", w.opts.KeepSentDays)
				}
			}
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
	ctx = tenant.AsPlatform(ctx)
	sent := 0
	err := database.WithTx(ctx, w.db, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, to_address, COALESCE(from_name, ''), subject, body, attempts FROM email_outbox
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
			if err := rows.Scan(&j.id, &j.msg.To, &j.msg.FromName, &j.msg.Subject, &j.msg.Text, &j.attempts); err != nil {
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
	ctx = tenant.AsPlatform(ctx)
	if w.opts.DebtReminderEvery <= 0 {
		return 0, nil
	}
	queued := 0
	err := database.WithTx(ctx, w.db, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT u.id, u.store_id, u.email, u.first_name, b.balance,
			       s.slug, s.name, s.currency, s.timezone
			FROM users u
			JOIN stores s ON s.id = u.store_id AND s.status = 'active'
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
			store   int64
			shop    tenant.Store
			to      Recipient
			balance decimal.Decimal
		}
		var list []due
		for rows.Next() {
			var d due
			if err := rows.Scan(&d.id, &d.store, &d.to.Email, &d.to.FirstName, &d.balance,
				&d.shop.Slug, &d.shop.Name, &d.shop.Currency, &d.shop.Timezone); err != nil {
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
			// Platform mode has no current store, so the reminder names its store.
			if err := enqueueFor(ctx, tx, d.store, KindDebtReminder, w.opts.Branding.ForStore(d.shop).DebtReminder(d.to, d.balance)); err != nil {
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
