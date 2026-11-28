package tests

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"
	"testing"
)

// getRaw fetches a non-JSON response (a CSV download).
func getRaw(t *testing.T, path, token string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest("GET", server.URL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

// Private admin notes on orders: saved, shown to admins, never to the customer.
func TestOrderAdminNotes(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "notes-admin@example.com")
	pid := product(t, admin, J{"name": "Notes Lamp", "category": "notes", "price": 15, "stock": 10})
	buyer := newCustomer(t, "notes-buyer@example.com")
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(pid)})
	order := id(must(t, 201, "POST", "/api/orders", buyer, nil))

	// A new order has an empty note; admins see the field, customers do not.
	if obj(must(t, 200, "GET", "/api/admin/orders/"+order, admin, nil))["admin_note"] != "" {
		t.Fatal("a new order has an empty admin note")
	}
	saved := obj(must(t, 200, "PUT", "/api/admin/orders/"+order+"/note", admin, J{"note": "  Deliver after 5 pm  "}))
	if saved["admin_note"] != "Deliver after 5 pm" || saved["admin_note_updated_at"] == nil {
		t.Fatalf("note saved and trimmed: %v / %v", saved["admin_note"], saved["admin_note_updated_at"])
	}
	for _, path := range []string{"/api/orders/" + order, "/api/orders"} {
		_, res := call(t, "GET", path, buyer, nil)
		if strings.Contains(str(res), "Deliver after 5 pm") || strings.Contains(str(res), "admin_note") {
			t.Fatalf("the customer must never see the admin note (%s): %v", path, res)
		}
	}
	list := obj(must(t, 200, "GET", "/api/admin/orders", admin, nil))
	found := false
	for _, o := range arr(list["items"]) {
		if id(o) == order {
			found = obj(o)["admin_note"] == "Deliver after 5 pm"
		}
	}
	if !found {
		t.Fatal("the admin order list shows the note")
	}

	// Limits and permissions.
	must(t, 400, "PUT", "/api/admin/orders/"+order+"/note", admin, J{"note": strings.Repeat("x", 2001)})
	must(t, 403, "PUT", "/api/admin/orders/"+order+"/note", buyer, J{"note": "mine"})
	must(t, 404, "PUT", "/api/admin/orders/999999/note", admin, J{"note": "x"})

	// An empty note clears it.
	cleared := obj(must(t, 200, "PUT", "/api/admin/orders/"+order+"/note", admin, J{"note": ""}))
	if cleared["admin_note"] != "" || cleared["admin_note_updated_at"] != nil {
		t.Fatalf("note cleared: %v", cleared)
	}
}

// CSV exports of orders and customers: Excel-friendly, store-scoped, safe.
func TestCSVExports(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "export-admin@example.com")
	pid := product(t, admin, J{"name": "Export Mug", "category": "export", "price": 12.5, "stock": 10})
	buyer := newCustomer(t, "export-buyer@example.com")
	// A name a spreadsheet would run as a formula, and an Arabic street.
	if _, err := testDB.Exec(`UPDATE users SET first_name = '=HYPERLINK("http://evil.example")', last_name = 'سارة'
		WHERE email = 'export-buyer@example.com'`); err != nil {
		t.Fatal(err)
	}
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(pid), "quantity": 2})
	order := id(must(t, 201, "POST", "/api/orders", buyer, nil))

	res, body := getRaw(t, "/api/admin/exports/orders.csv", admin)
	if res.StatusCode != 200 {
		t.Fatalf("orders.csv: %d %s", res.StatusCode, body)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("content type: %s", ct)
	}
	if cd := res.Header.Get("Content-Disposition"); !strings.Contains(cd, `filename="orders-`) {
		t.Fatalf("download file name: %s", cd)
	}
	if !strings.HasPrefix(body, "\xEF\xBB\xBF") {
		t.Fatal("the file starts with a UTF-8 byte-order mark (Excel and Arabic)")
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(body, "\xEF\xBB\xBF"))).ReadAll()
	if err != nil {
		t.Fatalf("valid CSV: %v", err)
	}
	if rows[0][0] != "order_id" || rows[0][len(rows[0])-1] != "total" {
		t.Fatalf("header: %v", rows[0])
	}
	var line []string
	for _, r := range rows[1:] {
		if r[0] == order {
			line = r
		}
	}
	if line == nil {
		t.Fatalf("the new order is in the export")
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		col[h] = i
	}
	if !strings.HasPrefix(line[col["customer"]], `'=HYPERLINK`) {
		t.Fatalf("formulas are neutralised: %q", line[col["customer"]])
	}
	if !strings.Contains(line[col["customer"]], "سارة") {
		t.Fatalf("Arabic text survives: %q", line[col["customer"]])
	}
	if line[col["units"]] != "2" || line[col["total"]] != "25.00" || line[col["status"]] != "pending" {
		t.Fatalf("order figures: units %s total %s status %s", line[col["units"]], line[col["total"]], line[col["status"]])
	}

	if line[col["items"]] != "Export Mug x2" {
		t.Fatalf("items column: %q", line[col["items"]])
	}

	// Filters: status, payment method and dates; bad input is refused.
	_, body = getRaw(t, "/api/admin/exports/orders.csv?payment_method=pay_later", admin)
	if strings.Contains(body, "\r\n"+order+",") {
		t.Fatal("the payment filter leaves out cash orders")
	}
	_, body = getRaw(t, "/api/admin/exports/orders.csv?status=delivered", admin)
	if strings.Contains(body, "\r\n"+order+",") {
		t.Fatal("the status filter leaves out pending orders")
	}
	_, body = getRaw(t, "/api/admin/exports/orders.csv?from=2000-01-01&to=2000-12-31", admin)
	if strings.Count(body, "\r\n") != 1 {
		t.Fatalf("no orders in 2000, only the header: %q", body)
	}
	for _, bad := range []string{"?from=28-11-2026", "?status=lost", "?from=2026-12-01&to=2026-11-01", "?payment_method=card"} {
		if res, _ := getRaw(t, "/api/admin/exports/orders.csv"+bad, admin); res.StatusCode != 400 {
			t.Errorf("%s: want 400, got %d", bad, res.StatusCode)
		}
	}

	// Customers.
	res, body = getRaw(t, "/api/admin/exports/customers.csv", admin)
	if res.StatusCode != 200 || !strings.Contains(body, "export-buyer@example.com") || !strings.Contains(body, "'=HYPERLINK") {
		t.Fatalf("customers.csv: %d", res.StatusCode)
	}

	// Admins only, and only their own store's data.
	if res, _ := getRaw(t, "/api/admin/exports/orders.csv", buyer); res.StatusCode != 403 {
		t.Fatalf("customers cannot export: %d", res.StatusCode)
	}
	if res, _ := getRaw(t, "/api/admin/exports/customers.csv", ""); res.StatusCode != 401 {
		t.Fatalf("not logged in: %d", res.StatusCode)
	}
}
