package tests

import (
	"context"
	"strings"
	"testing"

	"ecommerce-backend/internal/tenant"
)

// Multi-tenant isolation: one API, one database, stores that cannot see or
// touch each other's data - not through the API, not with edited IDs, not
// even with raw SQL on the API's own database connection.

func mustOn(t *testing.T, host string, want int, method, path, token string, body any) any {
	t.Helper()
	status, out := callOn(t, host, method, path, token, body)
	if status != want {
		t.Fatalf("[%s] %s %s: status %d, want %d; body: %v", host, method, path, status, want, out)
	}
	return out
}

// newStore creates a store directly in the database (the platform admin API
// comes in step 2) and returns its id and host name.
func newStore(t *testing.T, slug, status string) (int64, string) {
	t.Helper()
	var id int64
	if err := testDB.QueryRow(`INSERT INTO stores (slug, name, status) VALUES ($1, $2, $3) RETURNING id`,
		slug, strings.ToUpper(slug), status).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id, slug + "." + testPlatformDomain
}

// customerOn registers and logs in a confirmed customer of a store.
func customerOn(t *testing.T, host string, storeID int64, email string) string {
	t.Helper()
	mustOn(t, host, 201, "POST", "/api/register", "", J{
		"first_name": "Test", "last_name": "User", "email": email, "password": "password123", "phone": "0791234567",
	})
	if _, err := testDB.Exec(`UPDATE users SET email_verified_at = NOW() WHERE store_id = $1 AND email = $2`, storeID, email); err != nil {
		t.Fatal(err)
	}
	return str(obj(mustOn(t, host, 200, "POST", "/api/login", "", J{"email": email, "password": "password123"}))["token"])
}

func adminOn(t *testing.T, host string, storeID int64, email string) string {
	t.Helper()
	customerOn(t, host, storeID, email)
	if _, err := testDB.Exec(`UPDATE users SET role = 'admin' WHERE store_id = $1 AND email = $2`, storeID, email); err != nil {
		t.Fatal(err)
	}
	return str(obj(mustOn(t, host, 200, "POST", "/api/login", "", J{"email": email, "password": "password123"}))["token"])
}

func TestStoreFromHostName(t *testing.T) {
	requireServer(t)
	_, open := newStore(t, "hostcheck", "active")
	_, closed := newStore(t, "closedshop", "suspended")

	mustOn(t, open, 200, "GET", "/api/products", "", nil)
	if out := obj(mustOn(t, "nosuchstore."+testPlatformDomain, 404, "GET", "/api/products", "", nil)); out["error"] != "store not found" {
		t.Fatalf("unknown store: %v", out)
	}
	mustOn(t, closed, 403, "GET", "/api/products", "", nil)
	mustOn(t, "a.b."+testPlatformDomain, 200, "GET", "/api/products", "", nil) // not a store name: falls back to the default store
	mustOn(t, "HostCheck."+testPlatformDomain+":8443", 200, "GET", "/api/products", "", nil)
	mustOn(t, closed, 200, "GET", "/api/health", "", nil) // health checks do not need a store
}

func TestStoresAreIsolated(t *testing.T) {
	requireServer(t)
	otherID, other := newStore(t, "isolated", "active")
	mainAdmin := makeAdmin(t, "iso-main-admin@example.com")
	otherAdmin := adminOn(t, other, otherID, "iso-other-admin@example.com")

	// Catalog: each store sees only its own products.
	mainProduct := newProduct(t, mainAdmin, "Main Only Lamp", 5)
	otherProduct := id(mustOn(t, other, 201, "POST", "/api/products", otherAdmin, J{
		"name": "Other Only Vase", "name_ar": "Vase AR", "category": "Decor", "price": 30, "stock": 4,
	}))
	for _, it := range arr(obj(mustOn(t, other, 200, "GET", "/api/products?limit=100", "", nil))["items"]) {
		if str(obj(it)["id"]) == mainProduct {
			t.Fatal("another store's product is listed")
		}
	}
	mustOn(t, other, 404, "GET", "/api/products/"+mainProduct, "", nil)
	must(t, 404, "GET", "/api/products/"+otherProduct, "", nil)

	// Edited IDs: an admin cannot change or delete another store's rows.
	mustOn(t, other, 404, "PUT", "/api/products/"+mainProduct, otherAdmin, J{"name": "Hacked", "name_ar": "x", "price": 1, "stock": 1})
	mustOn(t, other, 404, "DELETE", "/api/products/"+mainProduct, otherAdmin, nil)
	if p := obj(must(t, 200, "GET", "/api/products/"+mainProduct, "", nil)); p["name"] != "Main Only Lamp" || p["is_active"] != true {
		t.Fatalf("main's product must be untouched: %v", p)
	}

	// A login only works on its own store.
	out := obj(mustOn(t, other, 401, "GET", "/api/profile", mainAdmin, nil))
	if !strings.Contains(str(out["error"]), "another store") {
		t.Fatalf("a main-store token on another store: %v", out)
	}
	must(t, 401, "GET", "/api/admin/orders", otherAdmin, nil)

	// The same email is two separate accounts.
	mainCustomer := customerOn(t, "", 1, "same-person@example.com")
	otherCustomer := customerOn(t, other, otherID, "same-person@example.com")
	must(t, 200, "POST", "/api/cart/items", mainCustomer, J{"product_id": n(mainProduct)})
	mustOn(t, other, 404, "POST", "/api/cart/items", otherCustomer, J{"product_id": n(mainProduct)}) // not this store's product
	mustOn(t, other, 200, "POST", "/api/cart/items", otherCustomer, J{"product_id": n(otherProduct)})
	if items := arr(obj(mustOn(t, other, 200, "GET", "/api/cart", otherCustomer, nil))["items"]); len(items) != 1 || str(obj(items[0])["product_id"]) != otherProduct {
		t.Fatalf("each account has its own cart: %v", items)
	}

	// Orders, users and the dashboard are per store.
	addAddress(t, mainCustomer)
	mainOrder := id(must(t, 201, "POST", "/api/orders", mainCustomer, nil))
	mustOn(t, other, 404, "GET", "/api/admin/orders/"+mainOrder, otherAdmin, nil)
	mustOn(t, other, 404, "PUT", "/api/admin/orders/"+mainOrder+"/status", otherAdmin, J{"status": "confirmed"})
	if page := obj(mustOn(t, other, 200, "GET", "/api/admin/orders", otherAdmin, nil)); str(page["total"]) != "0" {
		t.Fatalf("another store's orders are visible: %v", page)
	}
	for _, u := range arr(obj(mustOn(t, other, 200, "GET", "/api/admin/users?limit=100", otherAdmin, nil))["items"]) {
		if obj(u)["email"] == "iso-main-admin@example.com" {
			t.Fatal("another store's users are visible")
		}
	}
	if d := obj(mustOn(t, other, 200, "GET", "/api/admin/dashboard?days=7", otherAdmin, nil)); str(obj(d["today"])["orders"]) != "0" {
		t.Fatalf("the dashboard counts another store's orders: %v", d["today"])
	}

	// Names that must be unique are unique per store.
	must(t, 201, "POST", "/api/admin/coupons", mainAdmin, J{"code": "WELCOME", "discount_type": "fixed", "discount_value": 5})
	mustOn(t, other, 201, "POST", "/api/admin/coupons", otherAdmin, J{"code": "WELCOME", "discount_type": "fixed", "discount_value": 9})
	mustOn(t, other, 201, "POST", "/api/admin/categories", otherAdmin, J{"name": "Electronics"})
	mustOn(t, other, 400, "POST", "/api/cart/coupon", otherCustomer, J{"code": "NOPE"})
	if c := obj(mustOn(t, other, 200, "POST", "/api/cart/coupon", otherCustomer, J{"code": "welcome"})); str(c["discount"]) != "9" {
		t.Fatalf("each store's code has its own value: %v", c)
	}

	// Store settings are per store.
	mustOn(t, other, 200, "PUT", "/api/admin/settings", otherAdmin, J{"storeName": "Other Shop"})
	must(t, 200, "PUT", "/api/admin/settings", mainAdmin, J{"storeName": "Main Shop"})
	if s := obj(mustOn(t, other, 200, "GET", "/api/settings", "", nil)); s["storeName"] != "Other Shop" {
		t.Fatalf("settings leaked between stores: %v", s)
	}
}

// The database enforces the isolation itself, below the application code.
func TestDatabaseEnforcesIsolation(t *testing.T) {
	requireServer(t)
	otherID, other := newStore(t, "dbguard", "active")
	mainAdmin := makeAdmin(t, "guard-main-admin@example.com")
	mainProduct := newProduct(t, mainAdmin, "Guarded Lamp", 5)
	otherCustomer := customerOn(t, other, otherID, "guard-customer@example.com")
	_ = otherCustomer
	mustOn(t, other, 200, "GET", "/api/cart", otherCustomer, nil) // creates the cart
	var otherCart int64
	if err := testDB.QueryRow(`SELECT c.id FROM carts c JOIN users u ON u.id = c.user_id
		WHERE u.store_id = $1 AND u.email = 'guard-customer@example.com'`, otherID).Scan(&otherCart); err != nil {
		t.Fatal(err)
	}

	asOther := tenant.WithStore(context.Background(), tenant.Store{ID: otherID})
	asMain := tenant.WithStore(context.Background(), tenant.Store{ID: 1})

	// 1. Without a store, the API's connection sees nothing at all.
	var n int
	if err := appDB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM products`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("no store set must mean no rows (got %d, %v) - is the API running as stockwise_app?", n, err)
	}
	// 2. With a store, only its rows.
	if err := appDB.QueryRowContext(asOther, `SELECT COUNT(*) FROM products WHERE id = $1`, mainProduct).Scan(&n); err != nil || n != 0 {
		t.Fatalf("another store's product is readable with raw SQL: %d, %v", n, err)
	}
	if err := appDB.QueryRowContext(asMain, `SELECT COUNT(*) FROM products WHERE id = $1`, mainProduct).Scan(&n); err != nil || n != 1 {
		t.Fatalf("own product should be readable: %d, %v", n, err)
	}
	// 3. Writes into another store are refused...
	if _, err := appDB.ExecContext(asOther, `UPDATE products SET price = 1 WHERE id = $1`, mainProduct); err != nil {
		t.Fatal(err)
	}
	var price string
	testDB.QueryRow(`SELECT price FROM products WHERE id = $1`, mainProduct).Scan(&price)
	if price == "1.00" {
		t.Fatal("an UPDATE reached another store's row")
	}
	if _, err := appDB.ExecContext(asOther,
		`INSERT INTO categories (store_id, name) VALUES (1, 'Planted')`); err == nil {
		t.Fatal("inserting a row for another store must be refused (row-level security)")
	}
	// ...and so are links to another store's rows, even by id.
	if _, err := appDB.ExecContext(asOther,
		`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, 1)`, otherCart, mainProduct); err == nil {
		t.Fatal("a cart line pointing at another store's product must be refused (store-aware foreign key)")
	}
	// 4. A transaction stays in its store.
	tx, err := appDB.BeginTx(asOther, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(asMain, `SELECT 1`); err == nil {
		t.Fatal("a transaction must not switch to another store")
	}
}
