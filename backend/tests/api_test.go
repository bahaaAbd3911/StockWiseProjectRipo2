// Package tests contains end-to-end tests that run the real HTTP API against a
// real PostgreSQL database.
//
// They only run when INTEGRATION_DATABASE_URL points to a PostgreSQL server
// whose user may create databases, for example:
//
//	$env:INTEGRATION_DATABASE_URL="postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
//	go test ./tests -v
//
// A brand-new throwaway database (ecommerce_it_<timestamp>) is created, the
// schema migration is applied, the tests run, and the database is dropped.
// Your real database is never touched.
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/migrate"
	"ecommerce-backend/internal/routes"
)

var (
	server  *httptest.Server
	testDB  *sql.DB
	baseURL string // INTEGRATION_DATABASE_URL
)

func TestMain(m *testing.M) {
	base := os.Getenv("INTEGRATION_DATABASE_URL")
	if base == "" {
		fmt.Println("integration tests skipped: INTEGRATION_DATABASE_URL is not set")
		os.Exit(m.Run())
	}
	baseURL = base
	code, err := runWithTempDB(m, base)
	if err != nil {
		fmt.Println("integration setup failed:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runWithTempDB(m *testing.M, base string) (int, error) {
	ctx := context.Background()
	adminDB, err := database.Connect(ctx, base)
	if err != nil {
		return 0, err
	}
	defer adminDB.Close()

	name := fmt.Sprintf("ecommerce_it_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+name); err != nil {
		return 0, fmt.Errorf("create test database: %w", err)
	}
	defer adminDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)") //nolint:errcheck

	u, err := url.Parse(base)
	if err != nil {
		return 0, err
	}
	u.Path = "/" + name
	testDB, err = database.Connect(ctx, u.String())
	if err != nil {
		return 0, err
	}
	defer testDB.Close()

	// Apply every migration exactly like `go run ./cmd/migrate up` does.
	if _, err := migrate.Up(ctx, testDB, "../migrations"); err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}

	server = httptest.NewServer(routes.New(testDB, config.Config{
		JWTSecret:      "integration-test-secret-0123456789abcdef",
		JWTTTL:         time.Hour,
		AllowedOrigins: []string{"http://localhost:5173"},
	}))
	defer server.Close()
	return m.Run(), nil
}

// ---------- tiny HTTP client helpers ----------

type J = map[string]any

func call(t *testing.T, method, path, token string, body any) (int, any) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, server.URL+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	var out any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber() // keep money exact: "289.99" stays "289.99"
	_ = dec.Decode(&out)
	return res.StatusCode, out
}

// must calls the API and fails the test unless the status matches.
func must(t *testing.T, want int, method, path, token string, body any) any {
	t.Helper()
	status, out := call(t, method, path, token, body)
	if status != want {
		t.Fatalf("%s %s: status %d, want %d; body: %v", method, path, status, want, out)
	}
	return out
}

func obj(v any) J      { m, _ := v.(J); return m }
func arr(v any) []any  { a, _ := v.([]any); return a }
func str(v any) string { return fmt.Sprint(v) }
func id(v any) string  { return str(obj(v)["id"]) }

// n turns an ID string back into a JSON number for request bodies.
func n(s string) json.Number { return json.Number(s) }

func register(t *testing.T, email string) {
	must(t, 201, "POST", "/api/register", "", J{
		"first_name": "Test", "last_name": "User", "email": email,
		"password": "password123", "phone": "0791234567",
	})
}

func login(t *testing.T, email, password string) string {
	return str(obj(must(t, 200, "POST", "/api/login", "", J{"email": email, "password": password}))["token"])
}

func requireServer(t *testing.T) {
	if server == nil {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}
}

// ---------- the scenario ----------

func TestFullShoppingFlow(t *testing.T) {
	requireServer(t)

	// --- accounts ---
	register(t, "alice@example.com")
	register(t, "bob@example.com")
	register(t, "admin@example.com")
	if _, err := testDB.Exec(`UPDATE users SET role = 'admin' WHERE email = 'admin@example.com'`); err != nil {
		t.Fatal(err)
	}
	status, dup := call(t, "POST", "/api/register", "", J{"first_name": "A", "last_name": "B",
		"email": "ALICE@example.com", "password": "password123", "phone": "0791234567"})
	if status != 409 {
		t.Fatalf("duplicate email (different case) should be 409, got %d %v", status, dup)
	}
	alice := login(t, "Alice@Example.com", "password123") // email is case-insensitive
	bob := login(t, "bob@example.com", "password123")
	admin := login(t, "admin@example.com", "password123")
	must(t, 401, "POST", "/api/login", "", J{"email": "alice@example.com", "password": "wrong-password"})

	// --- permissions ---
	must(t, 401, "POST", "/api/products", "", J{})
	must(t, 403, "POST", "/api/products", alice, J{"name": "x"})

	// --- catalog (admin) ---
	laptop := id(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Laptop", "name_ar": "حاسوب", "category": "Electronics", "category_ar": "إلكترونيات",
		"price": 100, "discount_value": 10, "stock": 5,
	}))
	mouse := id(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Mouse", "name_ar": "فأرة", "category": "Electronics", "price": "19.99", "stock": 3,
	}))
	soldOut := id(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Camera", "name_ar": "كاميرا", "price": 250, "stock": 0,
	}))
	bad := obj(must(t, 400, "POST", "/api/products", admin, J{"name": "", "price": -5, "discount_value": 150}))
	if f := obj(bad["fields"]); f["name"] == nil || f["price"] == nil || f["discount_value"] == nil {
		t.Fatalf("validation should report every bad field, got %v", bad)
	}

	must(t, 201, "POST", "/api/products/"+laptop+"/colors", admin, J{"name": "Black", "name_ar": "أسود", "hex_code": "#000000"})
	must(t, 201, "POST", "/api/products/"+laptop+"/colors", admin, J{"name": "White", "name_ar": "أبيض", "hex_code": "#FFFFFF"})
	must(t, 409, "POST", "/api/products/"+laptop+"/colors", admin, J{"name": "Black", "name_ar": "أسود", "hex_code": "#111111"})
	must(t, 400, "POST", "/api/products/"+laptop+"/colors", admin, J{"name": "Red", "name_ar": "أحمر", "hex_code": "red"})

	img1 := obj(must(t, 201, "POST", "/api/products/"+laptop+"/images", admin, J{"image_url": "https://cdn.example.com/1.webp"}))
	if img1["is_primary"] != true {
		t.Fatal("the first image should automatically be primary")
	}
	must(t, 201, "POST", "/api/products/"+laptop+"/images", admin, J{"image_url": "https://cdn.example.com/2.webp", "is_primary": true, "display_order": 1})
	images := arr(must(t, 200, "GET", "/api/products/"+laptop+"/images", "", nil))
	primaries := 0
	for _, im := range images {
		if obj(im)["is_primary"] == true {
			primaries++
		}
	}
	if len(images) != 2 || primaries != 1 {
		t.Fatalf("want 2 images with exactly one primary, got %v", images)
	}

	// --- public catalog ---
	page := obj(must(t, 200, "GET", "/api/products?category=electronics&sort=price_asc", "", nil))
	if str(page["total"]) != "2" || str(obj(arr(page["items"])[0])["name"]) != "Mouse" {
		t.Fatalf("category filter / price sort wrong: %v", page)
	}
	detail := obj(must(t, 200, "GET", "/api/products/"+laptop, "", nil))
	if str(detail["final_price"]) != "90" || len(arr(detail["colors"])) != 2 ||
		str(detail["primary_image_url"]) != "https://cdn.example.com/2.webp" {
		t.Fatalf("product detail wrong: %v", detail)
	}
	if got := obj(must(t, 200, "GET", "/api/products?search=lap", "", nil)); str(got["total"]) != "1" {
		t.Fatalf("search failed: %v", got)
	}
	must(t, 400, "GET", "/api/products?sort=random", "", nil)
	must(t, 400, "GET", "/api/products?limit=1000", "", nil)
	cats := arr(must(t, 200, "GET", "/api/categories", "", nil))
	if len(cats) != 1 || str(obj(cats[0])["product_count"]) != "2" {
		t.Fatalf("categories wrong: %v", cats)
	}

	// --- cart ---
	must(t, 400, "POST", "/api/cart/items", alice, J{"product_id": n(laptop), "quantity": 1}) // color required
	cart := obj(must(t, 200, "POST", "/api/cart/items", alice, J{"product_id": n(laptop), "quantity": 2, "selected_color": "black"}))
	if str(cart["subtotal"]) != "180" {
		t.Fatalf("2 x 90.00 should be 180, got %v", cart["subtotal"])
	}
	if str(obj(arr(cart["items"])[0])["selected_color"]) != "Black" {
		t.Fatal("color should be stored with its canonical spelling")
	}
	must(t, 409, "POST", "/api/cart/items", alice, J{"product_id": n(laptop), "quantity": 1, "selected_color": "White"})
	must(t, 409, "POST", "/api/cart/items", alice, J{"product_id": n(laptop), "quantity": 4, "selected_color": "Black"}) // 2+4 > 5
	must(t, 409, "POST", "/api/cart/items", alice, J{"product_id": n(soldOut)})
	must(t, 400, "POST", "/api/cart/items", alice, J{"product_id": n(mouse), "selected_color": "Black"}) // mouse has no colors
	cart = obj(must(t, 200, "POST", "/api/cart/items", alice, J{"product_id": n(mouse)}))
	laptopItem := ""
	for _, it := range arr(cart["items"]) {
		if str(obj(it)["product_id"]) == laptop {
			laptopItem = id(it)
		}
	}
	cart = obj(must(t, 200, "PUT", "/api/cart/items/"+laptopItem, alice, J{"quantity": 3}))
	if str(cart["subtotal"]) != "289.99" || str(cart["item_count"]) != "4" {
		t.Fatalf("3 x 90 + 19.99 = 289.99, got %v", cart)
	}
	must(t, 404, "PUT", "/api/cart/items/"+laptopItem, bob, J{"quantity": 1}) // not bob's item

	// --- addresses ---
	must(t, 400, "POST", "/api/orders", alice, nil) // no address yet
	addr := obj(must(t, 201, "POST", "/api/addresses", alice, J{
		"recipient_name": "Alice", "phone": "0791234567", "city": "Amman", "area": "Abdoun", "street": "Main St",
	}))
	if addr["is_default"] != true {
		t.Fatal("the first address should become the default")
	}
	must(t, 404, "GET", "/api/addresses/"+id(addr), bob, nil)
	must(t, 404, "POST", "/api/orders", bob, J{"address_id": addr["id"]}) // cannot use alice's address

	// --- checkout ---
	order := obj(must(t, 201, "POST", "/api/orders", alice, nil))
	if str(order["total_amount"]) != "289.99" || len(arr(order["items"])) != 2 || order["status"] != "pending" {
		t.Fatalf("order wrong: %v", order)
	}
	if str(obj(arr(order["items"])[0])["unit_price"]) != "90" {
		t.Fatalf("order item should store the discounted price, got %v", order["items"])
	}
	if items := arr(obj(must(t, 200, "GET", "/api/cart", alice, nil))["items"]); len(items) != 0 {
		t.Fatal("cart should be empty after checkout")
	}
	if s := str(obj(must(t, 200, "GET", "/api/products/"+laptop, "", nil))["stock"]); s != "2" {
		t.Fatalf("laptop stock should be 5-3=2, got %s", s)
	}
	must(t, 404, "GET", "/api/orders/"+id(order), bob, nil)

	// --- customer cancel restores stock ---
	cancelled := obj(must(t, 200, "POST", "/api/orders/"+id(order)+"/cancel", alice, nil))
	if cancelled["status"] != "cancelled" {
		t.Fatalf("cancel failed: %v", cancelled)
	}
	must(t, 409, "POST", "/api/orders/"+id(order)+"/cancel", alice, nil)
	if s := str(obj(must(t, 200, "GET", "/api/products/"+laptop, "", nil))["stock"]); s != "5" {
		t.Fatalf("stock should be restored to 5, got %s", s)
	}

	// --- admin order life cycle ---
	must(t, 200, "POST", "/api/cart/items", alice, J{"product_id": n(laptop), "quantity": 1, "selected_color": "White"})
	order2 := id(must(t, 201, "POST", "/api/orders", alice, J{"address_id": addr["id"]}))
	must(t, 409, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": "shipped"}) // cannot skip steps
	must(t, 400, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": "lost"})
	for _, s := range []string{"confirmed", "processing", "shipped"} {
		must(t, 200, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": s})
	}
	must(t, 409, "POST", "/api/orders/"+order2+"/cancel", alice, nil) // no longer pending

	// bob reviews before alice's order is delivered: not verified
	bobReview := obj(must(t, 201, "POST", "/api/products/"+laptop+"/reviews", bob, J{"comment": "Looks nice", "rating": 4}))
	if bobReview["verified_purchase"] != false || bobReview["reviewer_name"] != "Test U." {
		t.Fatalf("bob's review wrong: %v", bobReview)
	}

	delivered := obj(must(t, 200, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": "delivered"}))
	if delivered["payment_status"] != "paid" || obj(delivered["customer"])["email"] != "alice@example.com" {
		t.Fatalf("delivery should mark cash-on-delivery as paid: %v", delivered)
	}
	adminPage := obj(must(t, 200, "GET", "/api/admin/orders?status=delivered", admin, nil))
	if str(adminPage["total"]) != "1" {
		t.Fatalf("admin status filter wrong: %v", adminPage)
	}
	must(t, 403, "GET", "/api/admin/orders", alice, nil)

	// --- reviews & helpful votes ---
	aliceReview := obj(must(t, 201, "POST", "/api/products/"+laptop+"/reviews", alice, J{"comment": "Great laptop", "rating": 5}))
	if aliceReview["verified_purchase"] != true {
		t.Fatal("alice bought and received the laptop: her review must be verified")
	}
	must(t, 409, "POST", "/api/products/"+laptop+"/reviews", alice, J{"comment": "Again", "rating": 5})
	must(t, 400, "POST", "/api/products/"+laptop+"/reviews", bob, J{"comment": "", "rating": 9})
	must(t, 400, "POST", "/api/reviews/"+id(aliceReview)+"/helpful", alice, nil) // own review
	voted := obj(must(t, 200, "POST", "/api/reviews/"+id(aliceReview)+"/helpful", bob, nil))
	if str(voted["helpful_count"]) != "1" {
		t.Fatalf("helpful_count should be 1: %v", voted)
	}
	must(t, 409, "POST", "/api/reviews/"+id(aliceReview)+"/helpful", bob, nil)
	unvoted := obj(must(t, 200, "DELETE", "/api/reviews/"+id(aliceReview)+"/helpful", bob, nil))
	if str(unvoted["helpful_count"]) != "0" {
		t.Fatalf("helpful_count should be back to 0: %v", unvoted)
	}
	must(t, 404, "PUT", "/api/reviews/"+id(aliceReview), bob, J{"comment": "hacked", "rating": 1})
	edited := obj(must(t, 200, "PUT", "/api/reviews/"+id(aliceReview), alice, J{"comment": "Still great", "rating": 5}))
	if edited["is_edited"] != true {
		t.Fatal("edited review should be flagged")
	}
	if p := obj(must(t, 200, "GET", "/api/products/"+laptop, "", nil)); str(p["rating_average"]) != "4.5" || str(p["review_count"]) != "2" {
		t.Fatalf("rating average should be 4.5 over 2 reviews: %v %v", p["rating_average"], p["review_count"])
	}

	// --- favorites ---
	must(t, 201, "POST", "/api/favorites", bob, J{"product_id": n(laptop)})
	must(t, 200, "POST", "/api/favorites", bob, J{"product_id": n(laptop)})
	must(t, 404, "POST", "/api/favorites", bob, J{"product_id": 999999})
	favs := arr(must(t, 200, "GET", "/api/favorites", bob, nil))
	if len(favs) != 1 || str(obj(obj(favs[0])["product"])["final_price"]) != "90" {
		t.Fatalf("favorites wrong: %v", favs)
	}
	must(t, 200, "DELETE", "/api/favorites/"+laptop, bob, nil)
	must(t, 404, "DELETE", "/api/favorites/"+laptop, bob, nil)

	// --- waiting list ---
	must(t, 409, "POST", "/api/waiting-list", bob, J{"product_id": n(laptop), "selected_color": "Black"}) // in stock
	wl := obj(must(t, 201, "POST", "/api/waiting-list", bob, J{"product_id": n(soldOut), "quantity": 2}))
	must(t, 409, "POST", "/api/waiting-list", bob, J{"product_id": n(soldOut)})
	adminWL := obj(must(t, 200, "GET", "/api/admin/waiting-list?status=waiting", admin, nil))
	if str(adminWL["total"]) != "1" {
		t.Fatalf("admin waiting list wrong: %v", adminWL)
	}
	notified := obj(must(t, 200, "PUT", "/api/admin/waiting-list/"+id(wl)+"/status", admin, J{"status": "notified"}))
	if notified["notified_at"] == nil {
		t.Fatal("notified_at should be set")
	}
	must(t, 200, "DELETE", "/api/waiting-list/"+id(wl), bob, nil)
	must(t, 409, "DELETE", "/api/waiting-list/"+id(wl), bob, nil)
	must(t, 409, "PUT", "/api/admin/waiting-list/"+id(wl)+"/status", admin, J{"status": "waiting"}) // cancelled is final

	// --- soft-deleting a product ---
	must(t, 200, "POST", "/api/cart/items", bob, J{"product_id": n(mouse)})
	must(t, 200, "DELETE", "/api/products/"+mouse, admin, nil)
	must(t, 404, "GET", "/api/products/"+mouse, "", nil)
	must(t, 200, "GET", "/api/admin/products/"+mouse, admin, nil)
	bobCart := obj(must(t, 200, "GET", "/api/cart", bob, nil))
	if bobCart["has_unavailable_items"] != true || str(bobCart["subtotal"]) != "0" {
		t.Fatalf("inactive product should be flagged in the cart: %v", bobCart)
	}
	must(t, 201, "POST", "/api/addresses", bob, J{"recipient_name": "Bob", "phone": "0791234567", "city": "Irbid", "area": "Center", "street": "King St"})
	rejected := obj(must(t, 409, "POST", "/api/orders", bob, nil)) // the cart contains an inactive product
	if !strings.Contains(str(rejected["error"]), "Mouse: this product is no longer available") {
		t.Fatalf("checkout should explain which item is the problem: %v", rejected)
	}
}

func TestLastUnitCanOnlyBeSoldOnce(t *testing.T) {
	requireServer(t)
	register(t, "racer1@example.com")
	register(t, "racer2@example.com")
	register(t, "racer-admin@example.com")
	testDB.Exec(`UPDATE users SET role = 'admin' WHERE email = 'racer-admin@example.com'`)
	admin := login(t, "racer-admin@example.com", "password123")
	product := id(must(t, 201, "POST", "/api/products", admin, J{"name": "Last one", "name_ar": "الأخير", "price": 10, "stock": 1}))

	var tokens []string
	for _, email := range []string{"racer1@example.com", "racer2@example.com"} {
		tok := login(t, email, "password123")
		must(t, 201, "POST", "/api/addresses", tok, J{"recipient_name": "R", "phone": "0791234567", "city": "A", "area": "B", "street": "C"})
		must(t, 200, "POST", "/api/cart/items", tok, J{"product_id": n(product)})
		tokens = append(tokens, tok)
	}

	var wg sync.WaitGroup
	statuses := make([]int, len(tokens))
	for i, tok := range tokens {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses[i], _ = call(t, "POST", "/api/orders", tok, nil)
		}()
	}
	wg.Wait()

	joined := fmt.Sprint(statuses)
	if !(strings.Contains(joined, "201") && strings.Contains(joined, "409")) {
		t.Fatalf("exactly one checkout should win, got %v", statuses)
	}
	if s := str(obj(must(t, 200, "GET", "/api/admin/products/"+product, admin, nil))["stock"]); s != "0" {
		t.Fatalf("stock must never go below zero, got %s", s)
	}
}

func TestAccountSecurity(t *testing.T) {
	requireServer(t)
	register(t, "carol@example.com")
	carol := login(t, "carol@example.com", "password123")

	must(t, 400, "PUT", "/api/profile/password", carol, J{"current_password": "nope-nope", "new_password": "newpassword1"})
	must(t, 200, "PUT", "/api/profile/password", carol, J{"current_password": "password123", "new_password": "newpassword1"})
	must(t, 401, "POST", "/api/login", "", J{"email": "carol@example.com", "password": "password123"})
	carol = login(t, "carol@example.com", "newpassword1")

	profile := obj(must(t, 200, "PUT", "/api/profile", carol, J{"first_name": "Carol", "last_name": "Smith", "phone": "+962791234567"}))
	if profile["first_name"] != "Carol" || profile["password_hash"] != nil {
		t.Fatalf("profile update wrong (and must never expose password_hash): %v", profile)
	}

	must(t, 200, "DELETE", "/api/profile", carol, nil)
	must(t, 401, "GET", "/api/profile", carol, nil) // the old token stops working immediately
	must(t, 401, "POST", "/api/login", "", J{"email": "carol@example.com", "password": "newpassword1"})
}

func TestCORSPreflightAndHealth(t *testing.T) {
	requireServer(t)
	req, _ := http.NewRequest(http.MethodOptions, server.URL+"/api/cart", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 204 || res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("preflight failed: %d %v", res.StatusCode, res.Header)
	}
	if h := obj(must(t, 200, "GET", "/api/health", "", nil)); h["database"] != "ok" {
		t.Fatalf("health: %v", h)
	}
}
