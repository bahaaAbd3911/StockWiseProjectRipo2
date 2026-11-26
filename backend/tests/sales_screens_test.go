package tests

import (
	"testing"
	"time"
)

// The request bodies the React admin forms send (Front-End/src/api/sales.js:
// fromCoupon, fromOffer, fromBundle) and the fields the customer and admin
// screens read. If one of these changes, the screens break.

func requireFields(t *testing.T, what string, v any, fields ...string) {
	t.Helper()
	m := obj(v)
	for _, f := range fields {
		if _, ok := m[f]; !ok {
			t.Errorf("%s: missing field %q in %v", what, f, m)
		}
	}
}

func TestSalesScreens(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "screens-admin@example.com")
	mug := product(t, admin, J{"name": "Screens Mug", "category": "screens-kitchen", "price": 10, "stock": 20,
		"colors": []J{{"name": "White", "hex_code": "#FFFFFF"}}})
	plate := product(t, admin, J{"name": "Screens Plate", "category": "screens-kitchen", "price": 6, "stock": 20})
	gone := product(t, admin, J{"name": "Screens Sold Out", "category": "screens-kitchen", "price": 4, "stock": 0})

	// ---- admin forms: the exact shapes the screens send ----
	coupon := obj(must(t, 201, "POST", "/api/admin/coupons", admin, J{
		"code": "SCREENS10", "description": "", "description_ar": "", "discount_type": "percentage",
		"discount_value": 10, "scope": "all", "product_ids": []int{}, "category": nil,
		"min_order_amount": 0, "max_discount": nil, "usage_limit": nil, "per_customer_limit": nil,
		"is_active": true, "starts_at": nil, "ends_at": nil,
	}))
	requireFields(t, "coupon", coupon, "id", "code", "discount_type", "discount_value", "scope", "product_ids",
		"min_order_amount", "max_discount", "usage_limit", "per_customer_limit", "is_active", "status",
		"times_used", "total_discount", "starts_at", "ends_at")
	// Editing sends the whole form again.
	must(t, 200, "PUT", "/api/admin/coupons/"+id(coupon), admin, J{
		"code": "SCREENS10", "description": "Ten off", "description_ar": "", "discount_type": "percentage",
		"discount_value": 10, "scope": "category", "product_ids": []int{}, "category": "screens-kitchen",
		"min_order_amount": 5, "max_discount": 50, "usage_limit": 100, "per_customer_limit": 2,
		"is_active": true, "starts_at": nil, "ends_at": time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339),
	})
	// A bad form comes back with per-field messages (the modals show them).
	bad := obj(must(t, 400, "POST", "/api/admin/coupons", admin, J{
		"code": "x", "discount_type": "percentage", "discount_value": 150, "scope": "all", "product_ids": []int{},
	}))
	if f := obj(bad["fields"]); f["code"] == nil || f["discount_value"] == nil {
		t.Fatalf("field errors for the form: %v", bad)
	}

	offer := obj(must(t, 201, "POST", "/api/admin/offers", admin, J{
		"name": "Screens Sale", "name_ar": "تخفيض", "description": "", "description_ar": "",
		"discount_type": "fixed", "discount_value": 2, "scope": "products", "product_ids": []any{n(plate)},
		"category": nil, "is_active": true, "starts_at": rfc(-time.Minute), "ends_at": rfc(7 * 24 * time.Hour),
	}))
	// Offers need both dates (the form fills them in).
	must(t, 400, "POST", "/api/admin/offers", admin, J{
		"name": "No Dates", "name_ar": "بدون", "discount_type": "fixed", "discount_value": 1, "scope": "products",
		"product_ids": []any{n(plate)}, "category": nil, "is_active": true, "starts_at": nil, "ends_at": nil,
	})
	requireFields(t, "offer", offer, "id", "name", "name_ar", "discount_type", "discount_value", "scope",
		"product_ids", "category", "starts_at", "ends_at", "is_active", "status")
	if offer["status"] != "live" {
		t.Fatalf("offer status: %v", offer["status"])
	}

	bundle := obj(must(t, 201, "POST", "/api/admin/bundles", admin, J{
		"name": "Screens Set", "name_ar": "طقم", "description": "", "description_ar": "", "image_url": nil,
		"price": 12, "is_active": true, "starts_at": nil, "ends_at": nil,
		"items": []J{
			{"product_id": n(mug), "quantity": 1, "selected_color": "White"},
			{"product_id": n(plate), "quantity": 2, "selected_color": nil},
		},
	}))
	requireFields(t, "bundle", bundle, "id", "name", "name_ar", "price", "original_total", "savings",
		"available_stock", "is_available", "is_active", "items")
	requireFields(t, "bundle item", arr(bundle["items"])[0], "product_id", "product_name", "product_name_ar",
		"quantity", "selected_color", "primary_image_url")

	for _, list := range []string{"/api/admin/coupons", "/api/admin/offers", "/api/admin/bundles",
		"/api/admin/debts", "/api/admin/waiting-list?status=waiting"} {
		requireFields(t, list, must(t, 200, "GET", list, admin, nil), "items", "total")
	}

	// ---- customer screens ----
	buyer := newCustomer(t, "screens-buyer@example.com")

	// Offers page: list, then one offer with its products (a page).
	live := obj(must(t, 200, "GET", "/api/offers", "", nil))
	if len(arr(live["items"])) == 0 {
		t.Fatal("the live offer should be listed")
	}
	detail := obj(must(t, 200, "GET", "/api/offers/"+id(offer), "", nil))
	if len(arr(obj(detail["products"])["items"])) != 1 || obj(detail["offer"])["id"] == nil {
		t.Fatalf("offer detail is {offer, products: page}: %v", detail)
	}

	// Bundles page and the cart.
	if len(arr(obj(must(t, 200, "GET", "/api/bundles", "", nil))["items"])) == 0 {
		t.Fatal("the bundle should be listed")
	}
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(mug), "quantity": 1, "selected_color": "White"})
	must(t, 200, "POST", "/api/cart/bundles", buyer, J{"bundle_id": n(id(bundle)), "quantity": 1})
	must(t, 200, "POST", "/api/cart/coupon", buyer, J{"code": "screens10"})
	cart := obj(must(t, 200, "GET", "/api/cart", buyer, nil))
	requireFields(t, "cart", cart, "items", "bundles", "subtotal", "discount", "total", "coupon")
	requireFields(t, "cart bundle", arr(cart["bundles"])[0], "id", "bundle_id", "name", "name_ar", "image_url",
		"unit_price", "quantity", "line_total", "items", "is_available")
	requireFields(t, "cart coupon", cart["coupon"], "code", "discount_amount")
	// 10 (mug) + 12 (bundle) = 22; the coupon is 10% of the category products.
	if num(t, cart["subtotal"]) != 22 || num(t, cart["discount"]) <= 0 || num(t, cart["total"]) >= 22 {
		t.Fatalf("cart totals: subtotal %v discount %v total %v", cart["subtotal"], cart["discount"], cart["total"])
	}
	line := id(arr(cart["bundles"])[0])
	must(t, 200, "PUT", "/api/cart/bundles/"+line, buyer, J{"quantity": 2})
	must(t, 200, "DELETE", "/api/cart/bundles/"+line, buyer, nil)
	must(t, 200, "DELETE", "/api/cart/coupon", buyer, nil)

	// My Balance: no credit yet, so no pay later.
	debt := obj(must(t, 200, "GET", "/api/debt", buyer, nil))
	requireFields(t, "my debt account", debt["account"], "credit_limit", "balance", "available_credit", "pay_later_allowed")
	if obj(debt["account"])["pay_later_allowed"] != false {
		t.Fatal("pay later needs a credit limit")
	}
	must(t, 409, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"})

	// The admin "Allow a customer" flow, then checkout with Pay Later.
	var buyerID string
	if err := testDB.QueryRow(`SELECT id::text FROM users WHERE email = 'screens-buyer@example.com'`).Scan(&buyerID); err != nil {
		t.Fatal(err)
	}
	must(t, 200, "PUT", "/api/admin/customers/"+buyerID+"/credit-limit", admin, J{"credit_limit": 100})
	if obj(obj(must(t, 200, "GET", "/api/debt", buyer, nil))["account"])["pay_later_allowed"] != true {
		t.Fatal("pay later should be allowed after a credit limit")
	}
	order := obj(must(t, 201, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"}))
	statement := obj(must(t, 200, "GET", "/api/admin/customers/"+buyerID+"/debt?limit=50", admin, nil))
	tx := arr(obj(statement["transactions"])["items"])
	if len(tx) != 1 {
		t.Fatalf("one charge for the order: %v", tx)
	}
	requireFields(t, "debt transaction", tx[0], "id", "type", "amount", "balance_after", "note", "order_id", "created_at")
	if str(obj(tx[0])["order_id"]) != id(order) {
		t.Fatalf("the charge points at the order: %v", tx[0])
	}
	must(t, 201, "POST", "/api/admin/customers/"+buyerID+"/debt/payments", admin, J{"amount": 5, "note": nil})
	must(t, 201, "POST", "/api/admin/customers/"+buyerID+"/debt/adjustments", admin, J{"amount": -1, "note": "rounding"})
	must(t, 400, "POST", "/api/admin/customers/"+buyerID+"/debt/adjustments", admin, J{"amount": -1, "note": ""})

	// Waiting list: "Notify me" on a sold-out product, the tab, cancel.
	w := obj(must(t, 201, "POST", "/api/waiting-list", buyer, J{"product_id": n(gone), "quantity": 1, "selected_color": nil}))
	mine := arr(must(t, 200, "GET", "/api/waiting-list", buyer, nil))
	if len(mine) != 1 {
		t.Fatalf("my waiting list: %v", mine)
	}
	requireFields(t, "waiting item", mine[0], "id", "product_id", "product_name", "product_name_ar",
		"product_stock", "quantity", "selected_color", "status", "requested_at")
	adminList := arr(obj(must(t, 200, "GET", "/api/admin/waiting-list?status=waiting", admin, nil))["items"])
	found := false
	for _, it := range adminList {
		if id(it) == id(w) {
			found = true
			requireFields(t, "admin waiting item", it, "customer")
			requireFields(t, "admin waiting customer", obj(it)["customer"], "first_name", "last_name", "email")
		}
	}
	if !found {
		t.Fatal("the admin waiting list should show the request")
	}
	must(t, 200, "DELETE", "/api/waiting-list/"+id(w), buyer, nil)
}
