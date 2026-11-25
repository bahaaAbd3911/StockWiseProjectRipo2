package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/notify"
)

// ---------- helpers ----------

func makeAdmin(t *testing.T, email string) string {
	t.Helper()
	register(t, email)
	if _, err := testDB.Exec(`UPDATE users SET role = 'admin' WHERE email = $1`, email); err != nil {
		t.Fatal(err)
	}
	return login(t, email, "password123")
}

func newProduct(t *testing.T, admin, name string, stock int) string {
	t.Helper()
	return id(must(t, 201, "POST", "/api/products", admin, J{
		"name": name, "name_ar": name + " AR", "price": 20, "stock": stock,
	}))
}

func addAddress(t *testing.T, token string) {
	t.Helper()
	must(t, 201, "POST", "/api/addresses", token, J{
		"recipient_name": "Buyer", "phone": "0791234567", "city": "Amman", "area": "Abdoun", "street": "Main St",
	})
}

// waitForEmails waits until at least n emails whose subject contains
// subject reached addr, and returns them.
func waitForEmails(t *testing.T, addr, subject string, n int) []mail.Message {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if msgs := outbox.about(addr, subject); len(msgs) >= n {
			return msgs
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected %d %q email(s) to %s, got %d", n, subject, addr, len(outbox.about(addr, subject)))
	return nil
}

// ---------- email verification ----------

func TestEmailVerification(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "verify-admin@example.com")
	product := newProduct(t, admin, "Kettle", 5)

	registerUnverified(t, "ivy@example.com")
	first := waitForMail(t, outbox, "ivy@example.com", verifySubject, 1)
	ivy := login(t, "ivy@example.com", "password123") // logging in works before confirming
	if obj(must(t, 200, "GET", "/api/profile", ivy, nil))["email_verified"] != false {
		t.Fatal("a new account starts unverified")
	}

	// Shopping works, ordering does not.
	must(t, 200, "POST", "/api/cart/items", ivy, J{"product_id": n(product)})
	addAddress(t, ivy)
	blocked := obj(must(t, 403, "POST", "/api/orders", ivy, nil))
	if !strings.Contains(str(blocked["error"]), "confirm your email") {
		t.Fatalf("checkout should explain why it is blocked: %v", blocked)
	}

	// Asking again sends a new link and the old one stops working.
	must(t, 200, "POST", "/api/email/verify/resend", ivy, nil)
	second := waitForMail(t, outbox, "ivy@example.com", verifySubject, 2)
	must(t, 400, "POST", "/api/email/verify", "", J{"token": first})
	must(t, 400, "POST", "/api/email/verify", "", J{"token": "made-up"})
	must(t, 400, "POST", "/api/email/verify", "", J{})

	must(t, 200, "POST", "/api/email/verify", "", J{"token": second})
	must(t, 400, "POST", "/api/email/verify", "", J{"token": second}) // single use
	if obj(must(t, 200, "GET", "/api/profile", ivy, nil))["email_verified"] != true {
		t.Fatal("the email should now be verified")
	}
	must(t, 201, "POST", "/api/orders", ivy, nil) // same token, no new login needed
	msg := obj(must(t, 200, "POST", "/api/email/verify/resend", ivy, nil))
	if !strings.Contains(str(msg["message"]), "already confirmed") {
		t.Fatalf("resending for a confirmed email should say so: %v", msg)
	}

	// Expired links are refused.
	registerUnverified(t, "jack@example.com")
	expired := waitForMail(t, outbox, "jack@example.com", verifySubject, 1)
	testDB.Exec(`UPDATE email_verification_tokens SET expires_at = NOW() - INTERVAL '1 minute'
		WHERE user_id = (SELECT id FROM users WHERE email = 'jack@example.com')`)
	must(t, 400, "POST", "/api/email/verify", "", J{"token": expired})

	// Accounts from before the feature were marked verified by the migration;
	// admins are never blocked.
	must(t, 200, "POST", "/api/cart/items", admin, J{"product_id": n(product)})
}

// ---------- order notifications ----------

func TestOrderNotifications(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "orders-admin@example.com")
	product := newProduct(t, admin, "Teapot", 10)
	register(t, "kim@example.com")
	kim := login(t, "kim@example.com", "password123")
	addAddress(t, kim)

	must(t, 200, "POST", "/api/cart/items", kim, J{"product_id": n(product), "quantity": 2})
	order := id(must(t, 201, "POST", "/api/orders", kim, nil))
	placed := waitForEmails(t, "kim@example.com", "Order #"+order+" received", 1)[0]
	for _, want := range []string{"Teapot", "40.00 JOD", "Main St", "http://localhost:5173/orders/" + order, "StockWise"} { // the main store's name (stores.name)
		if !strings.Contains(placed.Subject+placed.Text, want) {
			t.Errorf("order email should mention %q:\n%s", want, placed.Text)
		}
	}

	must(t, 200, "PUT", "/api/admin/orders/"+order+"/status", admin, J{"status": "confirmed"})
	waitForEmails(t, "kim@example.com", "Order #"+order+": confirmed", 1)
	must(t, 200, "PUT", "/api/admin/orders/"+order+"/status", admin, J{"status": "processing"})
	must(t, 200, "PUT", "/api/admin/orders/"+order+"/status", admin, J{"status": "shipped"})
	waitForEmails(t, "kim@example.com", "Order #"+order+": shipped", 1)

	must(t, 200, "POST", "/api/cart/items", kim, J{"product_id": n(product)})
	second := id(must(t, 201, "POST", "/api/orders", kim, nil))
	must(t, 200, "POST", "/api/orders/"+second+"/cancel", kim, nil)
	waitForEmails(t, "kim@example.com", "Order #"+second+": cancelled", 1)

	// Every queued email is marked as sent.
	var unsent int
	testDB.QueryRow(`SELECT COUNT(*) FROM email_outbox WHERE to_address = 'kim@example.com' AND sent_at IS NULL`).Scan(&unsent)
	if unsent != 0 {
		t.Fatalf("%d email(s) still unsent", unsent)
	}
}

// ---------- delivery retries ----------

func TestEmailDeliveryIsRetried(t *testing.T) {
	requireServer(t)
	const addr = "retry@example.com"
	outbox.failFor(addr, 1) // the provider is down for the first attempt
	if err := notify.NewOutbox(testDB).Enqueue(context.Background(), "test", mail.Message{To: addr, Subject: "Retry me", Text: "hi"}); err != nil {
		t.Fatal(err)
	}

	// The failure is recorded and the email is rescheduled about a minute later.
	var attempts int
	var lastError *string
	var delay float64
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		testDB.QueryRow(`SELECT attempts, last_error, EXTRACT(EPOCH FROM next_attempt_at - NOW())
			FROM email_outbox WHERE to_address = $1`, addr).Scan(&attempts, &lastError, &delay)
		if attempts == 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if attempts != 1 || lastError == nil || !strings.Contains(*lastError, "outage") || delay < 30 || delay > 61 {
		t.Fatalf("failure not recorded properly: attempts=%d error=%v delay=%.0fs", attempts, lastError, delay)
	}
	if len(outbox.to(addr)) != 0 {
		t.Fatal("nothing should have been delivered yet")
	}

	// Pretend the minute has passed: the next attempt delivers it, once.
	testDB.Exec(`UPDATE email_outbox SET next_attempt_at = NOW() WHERE to_address = $1`, addr)
	waitForEmails(t, addr, "Retry me", 1)
	time.Sleep(200 * time.Millisecond)
	if got := len(outbox.to(addr)); got != 1 {
		t.Fatalf("the email must be delivered exactly once, got %d", got)
	}
}

// ---------- debts ----------

func TestDebtNotificationsAndReminders(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "debt-mail-admin@example.com")
	product := newProduct(t, admin, "Blender", 10)
	register(t, "lena@example.com")
	lena := login(t, "lena@example.com", "password123")
	lenaID := str(obj(must(t, 200, "GET", "/api/profile", lena, nil))["id"])
	addAddress(t, lena)
	acct := "/api/admin/customers/" + lenaID
	must(t, 200, "PUT", acct+"/credit-limit", admin, J{"credit_limit": 500})

	must(t, 200, "POST", "/api/cart/items", lena, J{"product_id": n(product), "quantity": 5})
	must(t, 201, "POST", "/api/orders", lena, J{"payment_method": "pay_later"}) // owes 100

	must(t, 201, "POST", acct+"/debt/payments", admin, J{"amount": 30})
	paid := waitForEmails(t, "lena@example.com", "Payment received", 1)[0]
	if !strings.Contains(paid.Text, "30.00 JOD") || !strings.Contains(paid.Text, "70.00 JOD") {
		t.Fatalf("payment email should show the amount and the remaining balance:\n%s", paid.Text)
	}
	must(t, 201, "POST", acct+"/debt/adjustments", admin, J{"amount": -5, "note": "loyalty discount"})
	adj := waitForEmails(t, "lena@example.com", "Account balance updated", 1)[0]
	if !strings.Contains(adj.Text, "loyalty discount") || !strings.Contains(adj.Text, "65.00 JOD") {
		t.Fatalf("adjustment email should show the reason and the new balance:\n%s", adj.Text)
	}

	// Reminders: due once the last charge is older than the reminder period.
	reminders := notify.NewWorker(testDB, nil, notify.WorkerOptions{
		DebtReminderEvery: time.Hour,
		Branding:          notify.NewBranding(notify.Templates{Store: "TestShop", FrontendURL: "http://localhost:5173", Currency: "JOD"}, nil, "", ""),
	})
	ctx := context.Background()
	if _, err := reminders.QueueDebtReminders(ctx); err != nil {
		t.Fatal(err)
	}
	countReminders := func() (n int) {
		testDB.QueryRow(`SELECT COUNT(*) FROM email_outbox WHERE kind = 'debt_reminder' AND to_address = 'lena@example.com'`).Scan(&n)
		return n
	}
	if countReminders() != 0 {
		t.Fatal("no reminder right after buying")
	}
	testDB.Exec(`UPDATE customer_debt_transactions SET created_at = created_at - INTERVAL '2 hours' WHERE user_id = $1`, lenaID)
	if _, err := reminders.QueueDebtReminders(ctx); err != nil {
		t.Fatal(err)
	}
	reminder := waitForEmails(t, "lena@example.com", "Friendly reminder", 1)[0]
	if !strings.Contains(reminder.Text, "65.00 JOD") {
		t.Fatalf("reminder should show the open balance:\n%s", reminder.Text)
	}
	if _, err := reminders.QueueDebtReminders(ctx); err != nil {
		t.Fatal(err)
	}
	if countReminders() != 1 {
		t.Fatal("a customer must not be reminded twice in the same period")
	}

	// Once everything is paid, no more reminders.
	must(t, 201, "POST", acct+"/debt/payments", admin, J{"amount": 65})
	testDB.Exec(`UPDATE users SET debt_reminded_at = NOW() - INTERVAL '2 hours' WHERE id = $1`, lenaID)
	reminders.QueueDebtReminders(ctx)
	if countReminders() != 1 {
		t.Fatal("no reminder when nothing is owed")
	}
}

// ---------- waiting list ----------

func TestBackInStockEmails(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "stock-admin@example.com")
	product := newProduct(t, admin, "Toaster", 0)
	register(t, "mona@example.com")
	register(t, "nour@example.com")
	mona := login(t, "mona@example.com", "password123")
	nour := login(t, "nour@example.com", "password123")
	must(t, 201, "POST", "/api/waiting-list", mona, J{"product_id": n(product)})
	cancelled := id(must(t, 201, "POST", "/api/waiting-list", nour, J{"product_id": n(product)}))
	must(t, 200, "DELETE", "/api/waiting-list/"+cancelled, nour, nil)

	// Editing without stock tells nobody.
	update := J{"name": "Toaster", "name_ar": "Toaster AR", "price": 25, "stock": 0}
	must(t, 200, "PUT", "/api/products/"+product, admin, update)
	time.Sleep(200 * time.Millisecond)
	if len(outbox.about("mona@example.com", "back in stock")) != 0 {
		t.Fatal("no email while still out of stock")
	}

	update["stock"] = 4
	must(t, 200, "PUT", "/api/products/"+product, admin, update)
	mail := waitForEmails(t, "mona@example.com", "Toaster is back in stock", 1)[0]
	if !strings.Contains(mail.Text, "/products/"+product) {
		t.Fatalf("the email should link to the product:\n%s", mail.Text)
	}
	if st := obj(arr(must(t, 200, "GET", "/api/waiting-list", mona, nil))[0])["status"]; st != "notified" {
		t.Fatalf("the request should now be notified, got %v", st)
	}
	must(t, 200, "PUT", "/api/products/"+product, admin, update) // another edit: no second email
	time.Sleep(200 * time.Millisecond)
	if len(outbox.about("mona@example.com", "back in stock")) != 1 || len(outbox.about("nour@example.com", "back in stock")) != 0 {
		t.Fatal("each waiting customer is told once; cancelled requests are not")
	}
}

// ---------- image uploads ----------

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{G: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// upload sends a multipart form with the file in field "image".
func upload(t *testing.T, method, path, token, filename string, data []byte, fields map[string]string) (int, J) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		mw.WriteField(k, v)
	}
	if data != nil {
		fw, _ := mw.CreateFormFile("image", filename)
		fw.Write(data)
	}
	mw.Close()
	req, _ := http.NewRequest(method, server.URL+path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out J
	_ = jsonDecode(res.Body, &out)
	return res.StatusCode, out
}

// fetchUpload GETs an uploaded file by its public URL.
func fetchUpload(t *testing.T, fileURL string) (*http.Response, []byte) {
	t.Helper()
	path, ok := strings.CutPrefix(fileURL, testPublicURL)
	if !ok {
		t.Fatalf("upload URL %q should start with PUBLIC_URL", fileURL)
	}
	res, err := http.Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res, b
}

func TestImageUploads(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "upload-admin@example.com")
	product := newProduct(t, admin, "Lamp", 3)
	register(t, "omar@example.com")
	omar := login(t, "omar@example.com", "password123")
	path := "/api/admin/products/" + product + "/images/upload"

	// Product image: stored, linked, served.
	status, img := upload(t, "POST", path, admin, "lamp.png", pngBytes(t, 30, 20), map[string]string{"is_primary": "true"})
	if status != 201 || img["is_primary"] != true {
		t.Fatalf("upload failed: %d %v", status, img)
	}
	fileURL := str(img["image_url"])
	res, body := fetchUpload(t, fileURL)
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "image/png" ||
		!strings.Contains(res.Header.Get("Cache-Control"), "immutable") || res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("serving the image: %d %v", res.StatusCode, res.Header)
	}
	if _, err := png.Decode(bytes.NewReader(body)); err != nil {
		t.Fatal("served file should be the picture")
	}
	if detail := obj(must(t, 200, "GET", "/api/products/"+product, "", nil)); detail["primary_image_url"] != fileURL {
		t.Fatalf("the product should show the uploaded image: %v", detail["primary_image_url"])
	}

	// Refusals.
	if s, out := upload(t, "POST", path, admin, "evil.png", []byte("<script>alert(1)</script>"), nil); s != 400 || obj(out["fields"])["image"] == nil {
		t.Fatalf("a fake image must be refused on the image field: %d %v", s, out)
	}
	big := append(pngBytes(t, 10, 10), make([]byte, 1<<20)...) // over the 1 MB test limit
	if s, out := upload(t, "POST", path, admin, "big.png", big, nil); s != 400 || !strings.Contains(str(obj(out["fields"])["image"]), "1 MB") {
		t.Fatalf("an oversized file must be refused: %d %v", s, out)
	}
	if s, _ := upload(t, "POST", path, admin, "", nil, nil); s != 400 {
		t.Fatalf("a missing file must be refused, got %d", s)
	}
	if s, _ := upload(t, "POST", path, omar, "lamp.png", pngBytes(t, 5, 5), nil); s != 403 {
		t.Fatalf("customers cannot upload product images, got %d", s)
	}
	must(t, 400, "POST", path, admin, J{"image": "x"}) // JSON instead of a form
	if s, _ := upload(t, "POST", "/api/admin/products/999999/images/upload", admin, "a.png", pngBytes(t, 5, 5), nil); s != 404 {
		t.Fatalf("unknown product should be 404, got %d", s)
	}
	for _, p := range []string{"/uploads/../go.mod", "/uploads/notours.png", "/uploads/" + strings.Repeat("a", 32) + ".png"} {
		if r, _ := http.Get(server.URL + p); r.StatusCode != 404 {
			t.Fatalf("%s should be 404, got %d", p, r.StatusCode)
		}
	}

	// Deleting the image deletes the file.
	must(t, 200, "DELETE", "/api/product-images/"+id(img), admin, nil)
	if r, _ := fetchUpload(t, fileURL); r.StatusCode != 404 {
		t.Fatalf("deleted image should be gone, got %d", r.StatusCode)
	}

	// Profile picture: upload, replace (old file removed), delete.
	status, user := upload(t, "POST", "/api/profile/image", omar, "me.png", pngBytes(t, 8, 8), nil)
	first := str(user["profile_image_url"])
	if status != 200 || !strings.HasPrefix(first, testPublicURL+"/uploads/") {
		t.Fatalf("profile upload failed: %d %v", status, user)
	}
	_, user = upload(t, "POST", "/api/profile/image", omar, "me2.png", pngBytes(t, 9, 9), nil)
	second := str(user["profile_image_url"])
	if r, _ := fetchUpload(t, first); r.StatusCode != 404 {
		t.Fatal("the replaced picture's file should be deleted")
	}
	if r, _ := fetchUpload(t, second); r.StatusCode != 200 {
		t.Fatal("the new picture should be served")
	}
	if out := obj(must(t, 200, "DELETE", "/api/profile/image", omar, nil)); out["profile_image_url"] != nil {
		t.Fatalf("picture should be removed: %v", out)
	}
	if r, _ := fetchUpload(t, second); r.StatusCode != 404 {
		t.Fatal("the removed picture's file should be deleted")
	}
	if s, _ := upload(t, "POST", "/api/profile/image", "", "me.png", pngBytes(t, 8, 8), nil); s != 401 {
		t.Fatalf("uploading a profile picture needs a login, got %d", s)
	}

	// Only finished files are left in the folder (no .tmp leftovers).
	entries, _ := os.ReadDir(uploadDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temporary file left behind: %s", e.Name())
		}
	}
}

func jsonDecode(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }
