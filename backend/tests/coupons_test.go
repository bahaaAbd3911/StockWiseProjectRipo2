package tests

import (
	"encoding/json"
	"strings"
	"testing"
)

func couponBody(code string, extra J) J {
	b := J{"code": code, "discount_type": "fixed", "discount_value": 5}
	for k, v := range extra {
		b[k] = v
	}
	return b
}

func codeError(t *testing.T, res any) string {
	t.Helper()
	return str(obj(obj(res)["fields"])["code"])
}

func TestCoupons(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "coupon-admin@example.com")
	pan := id(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Pan", "name_ar": "Pan AR", "category": "Kitchen", "price": 50, "discount_value": 10, "stock": 20,
	})) // sells at 45.00
	ball := id(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Ball", "name_ar": "Ball AR", "category": "Toys", "price": 30, "stock": 20,
	}))

	// --- admin: create and validate ---
	save10 := obj(must(t, 201, "POST", "/api/admin/coupons", admin, J{
		"code": " save10 ", "discount_type": "percentage", "discount_value": 10, "min_order_amount": 50,
		"description": "10% off orders over 50",
	}))
	if save10["code"] != "SAVE10" || save10["scope"] != "all" || str(save10["per_customer_limit"]) != "1" || save10["status"] != "live" {
		t.Fatalf("defaults: upper-case code, scope all, once per customer: %v", save10)
	}
	must(t, 201, "POST", "/api/admin/coupons", admin, couponBody("KITCHEN5", J{"scope": "category", "category": "kitchen"}))
	oneUse := id(must(t, 201, "POST", "/api/admin/coupons", admin, couponBody("ONEUSE", J{"discount_value": 20, "usage_limit": 1})))
	must(t, 409, "POST", "/api/admin/coupons", admin, couponBody("Save10", nil)) // codes are case-insensitive
	for name, body := range map[string]J{
		"bad characters":      couponBody("10% OFF!", nil),
		"too short":           couponBody("AB", nil),
		"percentage over 100": couponBody("BIG", J{"discount_type": "percentage", "discount_value": 150}),
		"cap on fixed":        couponBody("CAPPED", J{"max_discount": 10}),
		"category missing":    couponBody("CAT", J{"scope": "category"}),
		"products missing":    couponBody("PRODS", J{"scope": "products"}),
		"window backwards":    couponBody("WIN", J{"starts_at": "2026-12-10T00:00:00+03:00", "ends_at": "2026-12-01T00:00:00+03:00"}),
	} {
		if s, res := call(t, "POST", "/api/admin/coupons", admin, body); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}

	register(t, "rana@example.com")
	rana := login(t, "rana@example.com", "password123")
	addAddress(t, rana)
	must(t, 403, "POST", "/api/admin/coupons", rana, couponBody("MINE", nil))

	// --- applying to the cart ---
	must(t, 200, "POST", "/api/cart/items", rana, J{"product_id": n(ball)}) // 30.00
	if msg := codeError(t, must(t, 400, "POST", "/api/cart/coupon", rana, J{"code": "save10"})); !strings.Contains(msg, "add 20.00 more") {
		t.Fatalf("below the minimum the customer is told how much to add: %q", msg)
	}
	if msg := codeError(t, must(t, 400, "POST", "/api/cart/coupon", rana, J{"code": "NOPE"})); msg != "this code is not valid" {
		t.Fatalf("unknown code: %q", msg)
	}
	must(t, 400, "POST", "/api/cart/coupon", rana, J{"code": " "})

	must(t, 200, "POST", "/api/cart/items", rana, J{"product_id": n(pan)}) // + 45.00 = 75.00
	cart := obj(must(t, 200, "POST", "/api/cart/coupon", rana, J{"code": "save10"}))
	if str(cart["subtotal"]) != "75" || str(cart["discount"]) != "7.5" || str(cart["total"]) != "67.5" || obj(cart["coupon"])["code"] != "SAVE10" {
		t.Fatalf("10%% on top of the offer price: %v", cart)
	}

	// --- checkout uses it ---
	order := obj(must(t, 201, "POST", "/api/orders", rana, nil))
	if str(order["subtotal_amount"]) != "75" || str(order["discount_amount"]) != "7.5" || str(order["total_amount"]) != "67.5" || order["coupon_code"] != "SAVE10" {
		t.Fatalf("order should record the coupon: %v", order)
	}
	if obj(must(t, 200, "GET", "/api/cart", rana, nil))["coupon"] != nil {
		t.Fatal("the code is removed from the cart after checkout")
	}

	// Once per customer; cancelling gives the use back.
	must(t, 200, "POST", "/api/cart/items", rana, J{"product_id": n(pan), "quantity": 2})
	if msg := codeError(t, must(t, 400, "POST", "/api/cart/coupon", rana, J{"code": "SAVE10"})); msg != "you have already used this code" {
		t.Fatalf("second use: %q", msg)
	}
	must(t, 200, "POST", "/api/orders/"+id(order)+"/cancel", rana, nil)
	must(t, 200, "POST", "/api/cart/coupon", rana, J{"code": "SAVE10"})
	if c := obj(obj(must(t, 200, "GET", "/api/admin/coupons/"+id(save10), admin, nil))); str(c["times_used"]) != "0" {
		t.Fatalf("a cancelled order does not count as a use: %v", c)
	}
	must(t, 200, "DELETE", "/api/cart/coupon", rana, nil)

	// Category coupon: only the Kitchen line counts.
	must(t, 200, "POST", "/api/cart/items", rana, J{"product_id": n(ball)})
	cart = obj(must(t, 200, "POST", "/api/cart/coupon", rana, J{"code": "kitchen5"}))
	if str(cart["discount"]) != "5" {
		t.Fatalf("category coupon: %v", cart)
	}
	must(t, 200, "DELETE", "/api/cart", rana, nil)
	must(t, 200, "POST", "/api/cart/items", rana, J{"product_id": n(ball)})
	if msg := codeError(t, must(t, 400, "POST", "/api/cart/coupon", rana, J{"code": "KITCHEN5"})); !strings.Contains(msg, "does not apply") {
		t.Fatalf("category coupon on other products: %q", msg)
	}

	// Usage limit: someone else takes the only use.
	register(t, "sami@example.com")
	sami := login(t, "sami@example.com", "password123")
	addAddress(t, sami)
	must(t, 200, "POST", "/api/cart/items", sami, J{"product_id": n(pan)})
	must(t, 200, "POST", "/api/cart/coupon", sami, J{"code": "ONEUSE"})
	must(t, 200, "POST", "/api/cart/coupon", rana, J{"code": "ONEUSE"}) // both have it in the cart
	samiOrder := obj(must(t, 201, "POST", "/api/orders", sami, nil))
	if str(samiOrder["discount_amount"]) != "20" {
		t.Fatalf("fixed coupon: %v", samiOrder)
	}
	cart = obj(must(t, 200, "GET", "/api/cart", rana, nil))
	if p := str(obj(cart["coupon"])["problem"]); !strings.Contains(p, "usage limit") || str(cart["discount"]) != "0" {
		t.Fatalf("the cart should show why the code no longer works: %v", cart)
	}
	blocked := obj(must(t, 409, "POST", "/api/orders", rana, nil))
	if !strings.Contains(str(blocked["error"]), "ONEUSE cannot be used") {
		t.Fatalf("checkout must not silently drop the code: %v", blocked)
	}
	must(t, 200, "DELETE", "/api/cart/coupon", rana, nil)
	must(t, 201, "POST", "/api/orders", rana, nil)

	// Admin views.
	used := obj(must(t, 200, "GET", "/api/admin/coupons?status=used_up", admin, nil))
	if str(used["total"]) != "1" || obj(arr(used["items"])[0])["code"] != "ONEUSE" {
		t.Fatalf("status filter: %v", used)
	}
	detail := obj(must(t, 200, "GET", "/api/admin/coupons/"+oneUse, admin, nil))
	if str(detail["times_used"]) != "1" || str(detail["total_discount"]) != "20" {
		t.Fatalf("usage stats: %v", detail)
	}
	search := obj(must(t, 200, "GET", "/api/admin/coupons?search=kitch", admin, nil))
	if str(search["total"]) != "1" {
		t.Fatalf("search by code: %v", search)
	}

	// Disabling a coupon that is in a cart.
	must(t, 200, "POST", "/api/cart/items", sami, J{"product_id": n(pan)})
	must(t, 200, "POST", "/api/cart/coupon", sami, J{"code": "KITCHEN5"})
	must(t, 200, "PUT", "/api/admin/coupons/"+id(obj(arr(search["items"])[0])), admin,
		couponBody("KITCHEN5", J{"scope": "category", "category": "Kitchen", "is_active": false}))
	cart = obj(must(t, 200, "GET", "/api/cart", sami, nil))
	if p := str(obj(cart["coupon"])["problem"]); p != "this code is not valid" {
		t.Fatalf("a disabled coupon shows as not valid: %v", cart)
	}
	must(t, 409, "POST", "/api/orders", sami, nil)

	// Deleting a used coupon keeps the order history.
	must(t, 200, "DELETE", "/api/admin/coupons/"+oneUse, admin, nil)
	if o := obj(must(t, 200, "GET", "/api/orders/"+id(samiOrder), sami, nil)); o["coupon_code"] != "ONEUSE" || str(o["discount_amount"]) != "20" {
		t.Fatalf("orders keep the code and discount: %v", o)
	}
}

func TestAdminDashboard(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "dash-admin@example.com")
	register(t, "dash-buyer@example.com")
	buyer := login(t, "dash-buyer@example.com", "password123")
	addAddress(t, buyer)
	must(t, 403, "GET", "/api/admin/dashboard", buyer, nil)
	must(t, 400, "GET", "/api/admin/dashboard?days=0", admin, nil)
	must(t, 400, "GET", "/api/admin/dashboard?days=400", admin, nil)

	before := obj(must(t, 200, "GET", "/api/admin/dashboard?days=7", admin, nil))

	scarce := id(must(t, 201, "POST", "/api/products", admin, J{"name": "Rare Vase", "name_ar": "Vase AR", "price": 40, "stock": 3}))
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(scarce), "quantity": 2})
	must(t, 201, "POST", "/api/orders", buyer, nil) // 80.00, stock 3 -> 1

	d := obj(must(t, 200, "GET", "/api/admin/dashboard?days=7", admin, nil))
	if d["timezone"] != "Asia/Amman" || len(arr(d["sales_by_day"])) != 7 {
		t.Fatalf("7 local days expected: %v %v", d["timezone"], len(arr(d["sales_by_day"])))
	}
	delta := func(path ...string) float64 {
		get := func(m J) float64 {
			v := any(m)
			for _, p := range path {
				v = obj(v)[p]
			}
			return num(t, v)
		}
		return get(d) - get(before)
	}
	if delta("today", "orders") != 1 || delta("today", "revenue") != 80 || delta("period", "items_sold") != 2 {
		t.Fatalf("today's figures should include the new order: %v", d["today"])
	}
	if delta("orders_by_status", "pending") != 1 {
		t.Fatalf("orders by status: %v", d["orders_by_status"])
	}
	last := obj(arr(d["sales_by_day"])[6])
	if num(t, last["orders"]) < 1 {
		t.Fatalf("the last day of the chart is today: %v", last)
	}
	found := false
	for _, p := range arr(obj(d["low_stock"])["products"]) {
		if str(obj(p)["product_id"]) == scarce && str(obj(p)["stock"]) == "1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the product with 1 left should be in low stock: %v", d["low_stock"])
	}
	if len(arr(d["best_sellers"])) == 0 || d["debts"] == nil || d["customers"] == nil || d["payment_methods"] == nil {
		t.Fatalf("sections missing: %v", d)
	}
}

func num(t *testing.T, v any) float64 {
	t.Helper()
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			t.Fatal(err)
		}
		return f
	case nil:
		return 0
	}
	t.Fatalf("not a number: %#v", v)
	return 0
}
