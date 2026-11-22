// Package notify writes and delivers the emails customers receive.
//
//   - templates.go: the text of every email (English, then Arabic)
//   - outbox.go:    the email queue (email_outbox table) and the worker that
//     delivers it, retries failures, and queues debt reminders
//
// Emails that carry a secret link (password reset, email verification) are
// NOT queued: the queue is a table, and storing the raw token there would undo
// the "only a hash is stored" protection. Those are sent directly.
package notify

import (
	"fmt"
	"net/url"
	"strings"

	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

// Templates holds what every email needs to know about the shop.
type Templates struct {
	Store       string // STORE_NAME, shown in subjects and signatures
	FrontendURL string // FRONTEND_URL, for "view your order" links
	Currency    string // CURRENCY, printed after amounts ("" = none)
}

// Email kinds, stored in email_outbox.kind.
const (
	KindOrderPlaced    = "order_placed"
	KindOrderStatus    = "order_status"
	KindDebtPayment    = "debt_payment"
	KindDebtAdjustment = "debt_adjustment"
	KindDebtReminder   = "debt_reminder"
	KindBackInStock    = "back_in_stock"
)

// Recipient is who an email goes to.
type Recipient struct {
	Email     string
	FirstName string
}

func (t Templates) money(d decimal.Decimal) string {
	if t.Currency == "" {
		return d.StringFixed(2)
	}
	return d.StringFixed(2) + " " + t.Currency
}

func (t Templates) link(path string) string {
	return strings.TrimRight(t.FrontendURL, "/") + path
}

// compose builds a bilingual message with the shop's signature.
func (t Templates) compose(to Recipient, subjectEn, subjectAr, en, ar string) mail.Message {
	return mail.Message{
		To:      to.Email,
		Subject: fmt.Sprintf("%s - %s / %s", t.Store, subjectEn, subjectAr),
		Text: fmt.Sprintf("Hello %s,\n\n%s\n\n%s\n\n---\n\nمرحباً %s،\n\n%s\n\n%s\n",
			to.FirstName, strings.TrimSpace(en), t.Store, to.FirstName, strings.TrimSpace(ar), t.Store),
	}
}

// ---------- account (sent directly, never queued) ----------

func (t Templates) VerifyEmail(to Recipient, link string, validHours int) mail.Message {
	return t.compose(to, "Confirm your email address", "تأكيد البريد الإلكتروني",
		fmt.Sprintf(`Welcome! Please confirm your email address by opening this link within %d hours:

%s

You need a confirmed email address to place orders. If you did not create an
account, ignore this email.`, validHours, link),
		fmt.Sprintf(`أهلاً بك! يرجى تأكيد بريدك الإلكتروني بفتح الرابط التالي خلال %d ساعة:

%s

تحتاج إلى بريد إلكتروني مؤكد لتتمكن من الطلب. إذا لم تنشئ حساباً، تجاهل هذه الرسالة.`, validHours, link))
}

func (t Templates) PasswordReset(to Recipient, link string, validMinutes int) mail.Message {
	return t.compose(to, "Reset your password", "إعادة تعيين كلمة المرور",
		fmt.Sprintf(`Someone asked to reset the password of your account. If it was you, open this
link within %d minutes:

%s

If you did not ask for this, ignore this email - your password stays the same.`, validMinutes, link),
		fmt.Sprintf(`طلب أحدهم إعادة تعيين كلمة المرور لحسابك. إذا كنت أنت، افتح الرابط التالي خلال %d دقيقة:

%s

إذا لم تطلب ذلك، تجاهل هذه الرسالة - لن تتغير كلمة المرور.`, validMinutes, link))
}

// ---------- orders ----------

func (t Templates) orderLines(o models.Order) string {
	var b strings.Builder
	for _, it := range o.Items {
		name := it.ProductName
		if it.SelectedColor != nil && *it.SelectedColor != "" {
			name += " (" + *it.SelectedColor + ")"
		}
		fmt.Fprintf(&b, "  %d x %s  %s\n", it.Quantity, name, t.money(it.Subtotal))
	}
	for _, bu := range o.Bundles {
		fmt.Fprintf(&b, "  %d x %s  %s\n", bu.Quantity, bu.BundleName, t.money(bu.Subtotal))
	}
	return strings.TrimRight(b.String(), "\n")
}

func payLater(o models.Order) (en, ar string) {
	if o.PaymentMethod == models.PaymentPayLater {
		return "Payment: pay later (added to your account balance)", "الدفع: لاحقاً (أضيف المبلغ إلى رصيد حسابك)"
	}
	return "Payment: cash on delivery", "الدفع: نقداً عند الاستلام"
}

func (t Templates) OrderPlaced(to Recipient, o models.Order) mail.Message {
	payEn, payAr := payLater(o)
	lines := t.orderLines(o)
	link := t.link(fmt.Sprintf("/orders/%d", o.ID))
	return t.compose(to, fmt.Sprintf("Order #%d received", o.ID), fmt.Sprintf("تم استلام الطلب رقم %d", o.ID),
		fmt.Sprintf(`Thank you for your order! We received order #%d and will confirm it soon.

%s

Total: %s
%s
Delivery to: %s, %s, %s

Follow your order: %s`, o.ID, lines, t.money(o.TotalAmount), payEn, o.Street, o.Area, o.City, link),
		fmt.Sprintf(`شكراً لطلبك! استلمنا الطلب رقم %d وسنؤكده قريباً.

%s

المجموع: %s
%s
التوصيل إلى: %s، %s، %s

تابع طلبك: %s`, o.ID, lines, t.money(o.TotalAmount), payAr, o.Street, o.Area, o.City, link))
}

var statusText = map[string][2]string{
	models.OrderConfirmed:  {"has been confirmed and will be prepared soon", "تم تأكيده وسيتم تجهيزه قريباً"},
	models.OrderProcessing: {"is being prepared", "قيد التجهيز"},
	models.OrderShipped:    {"is on its way", "في الطريق إليك"},
	models.OrderDelivered:  {"has been delivered. Thank you for shopping with us!", "تم توصيله. شكراً لتسوقك معنا!"},
	models.OrderCancelled:  {"has been cancelled", "تم إلغاؤه"},
}

// OrderStatus reports a status change. ok=false for statuses customers are
// not told about.
func (t Templates) OrderStatus(to Recipient, o models.Order) (mail.Message, bool) {
	text, ok := statusText[o.Status]
	if !ok {
		return mail.Message{}, false
	}
	extraEn, extraAr := "", ""
	if o.Status == models.OrderCancelled && o.PaymentMethod == models.PaymentPayLater {
		extraEn = "\nThe amount charged to your account balance has been removed."
		extraAr = "\nتم حذف المبلغ الذي أضيف إلى رصيد حسابك."
	}
	link := t.link(fmt.Sprintf("/orders/%d", o.ID))
	return t.compose(to, fmt.Sprintf("Order #%d: %s", o.ID, o.Status), fmt.Sprintf("تحديث الطلب رقم %d", o.ID),
		fmt.Sprintf("Your order #%d %s.%s\n\nTotal: %s\nDetails: %s", o.ID, text[0], extraEn, t.money(o.TotalAmount), link),
		fmt.Sprintf("طلبك رقم %d %s.%s\n\nالمجموع: %s\nالتفاصيل: %s", o.ID, text[1], extraAr, t.money(o.TotalAmount), link)), true
}

// ---------- debts ----------

func (t Templates) DebtPayment(to Recipient, amount, balance decimal.Decimal) mail.Message {
	return t.compose(to, "Payment received", "تم استلام دفعة",
		fmt.Sprintf("We received your payment of %s. Thank you!\n\nRemaining balance: %s\nStatement: %s",
			t.money(amount), t.money(balance), t.link("/account/debt")),
		fmt.Sprintf("استلمنا دفعتك بقيمة %s. شكراً لك!\n\nالرصيد المتبقي: %s\nكشف الحساب: %s",
			t.money(amount), t.money(balance), t.link("/account/debt")))
}

func (t Templates) DebtAdjustment(to Recipient, amount decimal.Decimal, note string, balance decimal.Decimal) mail.Message {
	sign := ""
	if amount.IsPositive() {
		sign = "+"
	}
	return t.compose(to, "Account balance updated", "تحديث رصيد الحساب",
		fmt.Sprintf("Your account balance was corrected by %s%s.\nReason: %s\n\nBalance now: %s\nStatement: %s",
			sign, t.money(amount), note, t.money(balance), t.link("/account/debt")),
		fmt.Sprintf("تم تعديل رصيد حسابك بمقدار %s%s.\nالسبب: %s\n\nالرصيد الآن: %s\nكشف الحساب: %s",
			sign, t.money(amount), note, t.money(balance), t.link("/account/debt")))
}

func (t Templates) DebtReminder(to Recipient, balance decimal.Decimal) mail.Message {
	return t.compose(to, "Friendly reminder: open balance", "تذكير: رصيد مستحق",
		fmt.Sprintf("This is a friendly reminder that your account has an open balance of %s.\n\nStatement: %s\n\nIf you already paid, thank you - please ignore this email.",
			t.money(balance), t.link("/account/debt")),
		fmt.Sprintf("نود تذكيرك بأن على حسابك رصيداً مستحقاً بقيمة %s.\n\nكشف الحساب: %s\n\nإذا كنت قد دفعت بالفعل، شكراً لك - تجاهل هذه الرسالة.",
			t.money(balance), t.link("/account/debt")))
}

// ---------- waiting list ----------

func (t Templates) BackInStock(to Recipient, productID int64, name, nameAr string) mail.Message {
	link := t.link(fmt.Sprintf("/products/%d", productID))
	return t.compose(to, name+" is back in stock", "المنتج متوفر الآن",
		fmt.Sprintf("Good news: %s is available again. Quantities may be limited.\n\n%s", name, link),
		fmt.Sprintf("خبر سار: %s متوفر من جديد. الكميات قد تكون محدودة.\n\n%s", nameAr, link))
}

// AddQuery appends ?key=value to a frontend page URL (for emailed links).
func AddQuery(base, key, value string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
