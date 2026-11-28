package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/tenant"
	"ecommerce-backend/internal/validators"
)

// ExportService builds the admin's CSV downloads (orders, customers).
//
// The files open correctly in Excel and Google Sheets:
//   - UTF-8 with a byte-order mark, so Arabic names are not garbled in Excel
//   - dates in the store's time zone ("2026-11-28 14:05")
//   - text that a spreadsheet would run as a formula (=, +, -, @ ...) is
//     prefixed with ' : a customer could otherwise name themselves
//     =HYPERLINK(...) and have it run on the admin's computer ("CSV injection")
type ExportService struct {
	repo       *repository.ExportRepository
	defaultLoc *time.Location
	now        func() time.Time
}

// defaultLoc is STORE_TIMEZONE, used when the store has no time zone of its own.
func NewExportService(repo *repository.ExportRepository, defaultLoc *time.Location) *ExportService {
	if defaultLoc == nil {
		defaultLoc = time.UTC
	}
	return &ExportService{repo: repo, defaultLoc: defaultLoc, now: time.Now}
}

// location is the store's own time zone, else the server's default.
func (s *ExportService) location(ctx context.Context) *time.Location {
	if st, ok := tenant.FromContext(ctx); ok && st.Timezone != "" {
		if loc, err := time.LoadLocation(st.Timezone); err == nil {
			return loc
		}
	}
	return s.defaultLoc
}

// ExportFile is a finished download.
type ExportFile struct {
	Name string // e.g. orders-2026-11-28.csv
	Data []byte
}

// Orders exports the orders placed between from and to (YYYY-MM-DD, both
// included, store time; empty = no limit), optionally with one status and
// one payment method.
func (s *ExportService) Orders(ctx context.Context, from, to, status, payment string) (ExportFile, error) {
	loc := s.location(ctx)
	c := validators.New()
	fromT := parseDay(c, "from", from, loc)
	toT := parseDay(c, "to", to, loc)
	status = strings.ToLower(strings.TrimSpace(status))
	c.Add(status == "" || models.IsOrderStatus(status), "status", "is not a valid order status")
	payment = strings.ToLower(strings.TrimSpace(payment))
	c.Add(payment == "" || payment == models.PaymentCashOnDelivery || payment == models.PaymentPayLater,
		"payment_method", "must be cash_on_delivery or pay_later")
	if !fromT.IsZero() && !toT.IsZero() {
		c.Add(!toT.Before(fromT), "to", "must not be before from")
	}
	if err := c.Err(); err != nil {
		return ExportFile{}, err
	}
	if !toT.IsZero() {
		toT = toT.AddDate(0, 0, 1) // "to" includes that whole day
	}
	rows, err := s.repo.Orders(ctx, fromT, toT, status, payment)
	if err != nil {
		return ExportFile{}, err
	}

	header := []string{"order_id", "date", "status", "payment_method", "payment_status",
		"customer", "customer_email", "recipient", "phone", "city", "area", "street",
		"items", "units", "subtotal", "discount", "coupon", "shipping", "total"}
	records := make([][]string, 0, len(rows))
	for _, o := range rows {
		records = append(records, []string{
			strconv.FormatInt(o.ID, 10), o.CreatedAt.In(loc).Format("2006-01-02 15:04"),
			cell(o.Status), cell(o.PaymentMethod), cell(o.PaymentStatus),
			cell(o.CustomerName), cell(o.CustomerEmail), cell(o.Recipient), cell(o.Phone),
			cell(o.City), cell(o.Area), cell(o.Street),
			cell(o.Items), strconv.Itoa(o.Units), money(o.Subtotal), money(o.Discount), cell(o.CouponCode),
			money(o.Shipping), money(o.Total),
		})
	}
	data, err := writeCSV(header, records)
	return ExportFile{Name: "orders-" + s.now().In(loc).Format("2006-01-02") + ".csv", Data: data}, err
}

// Customers exports every account of the store.
func (s *ExportService) Customers(ctx context.Context) (ExportFile, error) {
	loc := s.location(ctx)
	rows, err := s.repo.Customers(ctx)
	if err != nil {
		return ExportFile{}, err
	}
	header := []string{"customer_id", "first_name", "last_name", "email", "phone", "role", "active",
		"email_confirmed", "joined", "orders", "total_spent", "credit_limit", "owes"}
	records := make([][]string, 0, len(rows))
	for _, u := range rows {
		joined := ""
		if u.JoinedAt != nil {
			joined = u.JoinedAt.In(loc).Format("2006-01-02")
		}
		records = append(records, []string{
			strconv.FormatInt(u.ID, 10), cell(u.FirstName), cell(u.LastName), cell(u.Email), cell(u.Phone),
			cell(u.Role), yesNo(u.Active), yesNo(u.EmailVerified), joined,
			strconv.Itoa(u.Orders), money(u.Spent), money(u.CreditLimit), money(u.Balance),
		})
	}
	data, err := writeCSV(header, records)
	return ExportFile{Name: "customers-" + s.now().In(loc).Format("2006-01-02") + ".csv", Data: data}, err
}

func parseDay(c *validators.Checker, field, v string, loc *time.Location) time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}
	}
	t, err := time.ParseInLocation("2006-01-02", v, loc)
	c.Add(err == nil, field, "must be a date like 2026-11-28")
	return t
}

// cell neutralises text a spreadsheet would treat as a formula.
func cell(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func money(d decimal.Decimal) string { return d.StringFixed(2) }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// writeCSV: UTF-8 with a byte-order mark (for Excel), CRLF line ends.
func writeCSV(header []string, records [][]string) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 byte-order mark
	w := csv.NewWriter(&buf)
	w.UseCRLF = true
	if err := w.Write(header); err != nil {
		return nil, err
	}
	if err := w.WriteAll(records); err != nil {
		return nil, err
	}
	return buf.Bytes(), w.Error()
}
