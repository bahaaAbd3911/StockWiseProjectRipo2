package tests

import (
	"strings"
	"testing"
	"time"
)

// newAdmin registers an account and promotes it to admin.
func newAdmin(t *testing.T, email string) string {
	t.Helper()
	register(t, email)
	if _, err := testDB.Exec(`UPDATE users SET role = 'admin' WHERE email = $1`, email); err != nil {
		t.Fatal(err)
	}
	return login(t, email, "password123")
}

// newCustomer registers, logs in and saves an address.
func newCustomer(t *testing.T, email string) string {
	t.Helper()
	register(t, email)
	tok := login(t, email, "password123")
	must(t, 201, "POST", "/api/addresses", tok, J{"recipient_name": "C", "phone": "0791234567",
		"city": "Amman", "area": "Abdoun", "street": "Main St"})
	return tok
}

func product(t *testing.T, admin string, body J) string {
	t.Helper()
	if body["name_ar"] == nil {
		body["name_ar"] = "منتج"
	}
	return id(must(t, 201, "POST", "/api/products", admin, body))
}

func rfc(d time.Duration) string { return time.Now().Add(d).Format(time.RFC3339) }

// ---------------------------------------------------------------------------

func TestOffers(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "offer-admin@example.com")
	buyer := newCustomer(t, "offer-buyer@example.com")

	phone := product(t, admin, J{"name": "Offer Phone", "category": "Gadgets", "price": 100, "discount_value": 10, "stock": 10})
	caseP := product(t, admin, J{"name": "Offer Case", "category": "Cases", "price": 200, "stock": 10})
	cable := product(t, admin, J{"name": "Plain Cable", "category": "Gadgets", "price": 80, "stock": 10})

	// --- validation ---
	bad := obj(must(t, 400, "POST", "/api/admin/offers", admin, J{"name": "x", "name_ar": "x",
		"discount_type": "percentage", "discount_value": 150, "scope": "products",
		"starts_at": rfc(time.Hour), "ends_at": rfc(-time.Hour)}))
	f := obj(bad["fields"])
	if f["discount_value"] == nil || f["product_ids"] == nil || f["ends_at"] == nil {
		t.Fatalf("offer validation should list every problem: %v", bad)
	}
	must(t, 404, "POST", "/api/admin/offers", admin, J{"name": "x", "name_ar": "x", "discount_type": "fixed",
		"discount_value": 5, "scope": "products", "product_ids": []any{n(phone), 999999},
		"starts_at": rfc(-time.Hour), "ends_at": rfc(time.Hour)})
	must(t, 403, "POST", "/api/admin/offers", buyer, J{})

	// --- create offers ---
	spring := obj(must(t, 201, "POST", "/api/admin/offers", admin, J{"name": "Spring sale", "name_ar": "تخفيضات الربيع",
		"discount_type": "percentage", "discount_value": 25, "scope": "products", "product_ids": []any{n(phone)},
		"starts_at": rfc(-time.Hour), "ends_at": rfc(48 * time.Hour)}))
	if spring["status"] != "live" {
		t.Fatalf("offer should be live: %v", spring)
	}
	must(t, 201, "POST", "/api/admin/offers", admin, J{"name": "Small coupon", "name_ar": "خصم صغير",
		"discount_type": "fixed", "discount_value": 10, "scope": "products", "product_ids": []any{n(phone)},
		"starts_at": rfc(-time.Hour), "ends_at": rfc(48 * time.Hour)}) // 90: worse than the spring sale
	must(t, 201, "POST", "/api/admin/offers", admin, J{"name": "Cases week", "name_ar": "أسبوع الأغطية",
		"discount_type": "fixed", "discount_value": 50, "scope": "category", "category": "cases",
		"starts_at": rfc(-time.Hour), "ends_at": rfc(48 * time.Hour)})
	must(t, 201, "POST", "/api/admin/offers", admin, J{"name": "Future", "name_ar": "لاحقاً",
		"discount_type": "percentage", "discount_value": 50, "scope": "products", "product_ids": []any{n(cable)},
		"starts_at": rfc(24 * time.Hour), "ends_at": rfc(72 * time.Hour)})

	// --- prices: the single best discount wins, no stacking ---
	p := obj(must(t, 200, "GET", "/api/products/"+phone, "", nil))
	if str(p["final_price"]) != "75" || str(obj(p["offer"])["name"]) != "Spring sale" {
		t.Fatalf("phone should cost 75 via the spring sale (not 90, not stacked): %v %v", p["final_price"], p["offer"])
	}
	if c := obj(must(t, 200, "GET", "/api/products/"+caseP, "", nil)); str(c["final_price"]) != "150" {
		t.Fatalf("category offer should take 50 off the case: %v", c["final_price"])
	}
	if c := obj(must(t, 200, "GET", "/api/products/"+cable, "", nil)); str(c["final_price"]) != "80" || c["offer"] != nil {
		t.Fatalf("scheduled offer must not apply yet: %v", c)
	}

	// --- listing ---
	if live := obj(must(t, 200, "GET", "/api/offers", "", nil)); str(live["total"]) != "3" {
		t.Fatalf("3 offers are live: %v", live["total"])
	}
	sched := obj(must(t, 200, "GET", "/api/admin/offers?status=scheduled", admin, nil))
	if str(sched["total"]) != "1" {
		t.Fatalf("1 offer is scheduled: %v", sched)
	}
	detail := obj(must(t, 200, "GET", "/api/offers/"+id(spring), "", nil))
	if items := arr(obj(detail["products"])["items"]); len(items) != 1 || id(items[0]) != phone {
		t.Fatalf("spring sale covers only the phone: %v", detail)
	}
	sorted := arr(obj(must(t, 200, "GET", "/api/products?category=gadgets&sort=price_asc", "", nil))["items"])
	if len(sorted) != 2 || id(sorted[0]) != phone {
		t.Fatalf("price sort must use the offer price (phone 75 < cable 80): %v", sorted)
	}
	onOffer := obj(must(t, 200, "GET", "/api/products?on_offer=true&category=gadgets", "", nil))
	if str(onOffer["total"]) != "1" {
		t.Fatalf("only the phone is on offer in gadgets: %v", onOffer)
	}

	// --- cart & checkout keep the offer ---
	cart := obj(must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(phone), "quantity": 2}))
	line := obj(arr(cart["items"])[0])
	if str(line["unit_price"]) != "75" || str(cart["subtotal"]) != "150" || obj(line["offer"])["name"] != "Spring sale" {
		t.Fatalf("cart should use the offer price: %v", cart)
	}
	order := obj(must(t, 201, "POST", "/api/orders", buyer, nil))
	item := obj(arr(order["items"])[0])
	if str(order["total_amount"]) != "150" || item["offer_name"] != "Spring sale" {
		t.Fatalf("order should record the offer: %v", order)
	}

	// --- switching the offer off / deleting it ---
	must(t, 200, "PUT", "/api/admin/offers/"+id(spring), admin, J{"name": "Spring sale", "name_ar": "تخفيضات الربيع",
		"discount_type": "percentage", "discount_value": 25, "scope": "products", "product_ids": []any{n(phone)},
		"starts_at": rfc(-time.Hour), "ends_at": rfc(48 * time.Hour), "is_active": false})
	if p := obj(must(t, 200, "GET", "/api/products/"+phone, "", nil)); str(p["final_price"]) != "90" || p["offer"] != nil {
		t.Fatalf("with the sale off the product discount (90) applies; the 90 coupon ties and loses: %v %v", p["final_price"], p["offer"])
	}
	must(t, 200, "DELETE", "/api/admin/offers/"+id(spring), admin, nil)
	must(t, 404, "GET", "/api/admin/offers/"+id(spring), admin, nil)
	kept := obj(arr(obj(must(t, 200, "GET", "/api/orders/"+id(order), buyer, nil))["items"])[0])
	if kept["offer_name"] != "Spring sale" || kept["offer_id"] != nil {
		t.Fatalf("order history keeps the offer name after deletion: %v", kept)
	}
}

func TestBundles(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "bundle-admin@example.com")
	buyer := newCustomer(t, "bundle-buyer@example.com")

	keyboard := product(t, admin, J{"name": "Keyboard", "price": 50, "stock": 3})
	must(t, 201, "POST", "/api/products/"+keyboard+"/colors", admin, J{"name": "Black", "name_ar": "أسود", "hex_code": "#000000"})
	mouse := product(t, admin, J{"name": "Mouse Two", "price": 20, "stock": 4})

	must(t, 400, "POST", "/api/admin/bundles", admin, J{"name": "x", "name_ar": "x", "price": 10,
		"items": []any{J{"product_id": n(keyboard), "quantity": 1}, J{"product_id": n(mouse), "quantity": 1}}}) // keyboard needs a color
	must(t, 400, "POST", "/api/admin/bundles", admin, J{"name": "x", "name_ar": "x", "price": 10,
		"items": []any{J{"product_id": n(mouse), "quantity": 1}}}) // a single item is not a bundle

	bundleBody := J{"name": "Desk set", "name_ar": "طقم المكتب", "price": 75, "items": []any{
		J{"product_id": n(keyboard), "quantity": 1, "selected_color": "black"},
		J{"product_id": n(mouse), "quantity": 2},
	}}
	b := obj(must(t, 201, "POST", "/api/admin/bundles", admin, bundleBody))
	bundle := id(b)
	if str(b["original_total"]) != "90" || str(b["savings"]) != "15" || str(b["available_stock"]) != "2" || b["is_available"] != true {
		t.Fatalf("bundle totals wrong (50 + 2x20 = 90, save 15, stock min(3/1, 4/2) = 2): %v", b)
	}
	if public := obj(must(t, 200, "GET", "/api/bundles", "", nil)); str(public["total"]) != "1" {
		t.Fatalf("bundle should be listed: %v", public)
	}

	// --- cart ---
	must(t, 409, "POST", "/api/cart/bundles", buyer, J{"bundle_id": n(bundle), "quantity": 3}) // needs 6 mice
	cart := obj(must(t, 200, "POST", "/api/cart/bundles", buyer, J{"bundle_id": n(bundle)}))
	if len(arr(cart["bundles"])) != 1 || str(cart["subtotal"]) != "75" {
		t.Fatalf("cart should hold the bundle: %v", cart)
	}
	cart = obj(must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(mouse), "quantity": 3}))
	rejected := obj(must(t, 409, "POST", "/api/orders", buyer, nil)) // 2 in bundle + 3 alone = 5 > 4
	if !strings.Contains(str(rejected["error"]), "Mouse Two: only 4 left in stock, but your cart needs 5") {
		t.Fatalf("checkout should count the bundle and the single item together: %v", rejected)
	}
	mouseLine := id(arr(cart["items"])[0])
	must(t, 200, "PUT", "/api/cart/items/"+mouseLine, buyer, J{"quantity": 2})

	order := obj(must(t, 201, "POST", "/api/orders", buyer, nil))
	ob := arr(order["bundles"])
	if str(order["total_amount"]) != "115" || len(ob) != 1 || len(arr(obj(ob[0])["items"])) != 2 {
		t.Fatalf("order should contain the bundle with its 2 products (75 + 2x20 = 115): %v", order)
	}
	if s := str(obj(must(t, 200, "GET", "/api/admin/products/"+mouse, admin, nil))["stock"]); s != "0" {
		t.Fatalf("mouse stock 4 - 2 (bundle) - 2 (single) = 0, got %s", s)
	}
	must(t, 200, "POST", "/api/orders/"+id(order)+"/cancel", buyer, nil)
	if s := str(obj(must(t, 200, "GET", "/api/admin/products/"+mouse, admin, nil))["stock"]); s != "4" {
		t.Fatalf("cancelling must give the bundle's stock back too, got %s", s)
	}
	if s := str(obj(must(t, 200, "GET", "/api/admin/products/"+keyboard, admin, nil))["stock"]); s != "3" {
		t.Fatalf("keyboard stock should be back to 3, got %s", s)
	}

	// --- buying a bundle makes reviews of its products verified ---
	must(t, 200, "POST", "/api/cart/bundles", buyer, J{"bundle_id": n(bundle)})
	order2 := id(must(t, 201, "POST", "/api/orders", buyer, nil))
	for _, s := range []string{"confirmed", "processing", "shipped", "delivered"} {
		must(t, 200, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": s})
	}
	review := obj(must(t, 201, "POST", "/api/products/"+keyboard+"/reviews", buyer, J{"comment": "Nice", "rating": 5}))
	if review["verified_purchase"] != true {
		t.Fatal("a product bought inside a bundle counts as a verified purchase")
	}

	// --- deactivating ---
	must(t, 200, "POST", "/api/cart/bundles", buyer, J{"bundle_id": n(bundle)})
	must(t, 200, "DELETE", "/api/admin/bundles/"+bundle, admin, nil)
	must(t, 404, "GET", "/api/bundles/"+bundle, "", nil)
	if a := obj(must(t, 200, "GET", "/api/admin/bundles/"+bundle, admin, nil)); a["is_active"] != false {
		t.Fatalf("admin still sees the inactive bundle: %v", a)
	}
	c := obj(must(t, 200, "GET", "/api/cart", buyer, nil))
	if obj(arr(c["bundles"])[0])["is_available"] != false || c["has_unavailable_items"] != true {
		t.Fatalf("an inactive bundle in the cart is flagged: %v", c)
	}
	must(t, 200, "DELETE", "/api/cart/bundles/"+id(arr(c["bundles"])[0]), buyer, nil)
}

func TestCustomerDebts(t *testing.T) {
	requireServer(t)
	adminEmail := "debt-admin@example.com"
	admin := newAdmin(t, adminEmail)
	adminID := str(obj(obj(must(t, 200, "POST", "/api/login", "", J{"email": adminEmail, "password": "password123"}))["user"])["id"])
	buyer := newCustomer(t, "debt-buyer@example.com")
	buyerID := str(obj(must(t, 200, "GET", "/api/profile", buyer, nil))["id"])
	tv := product(t, admin, J{"name": "TV", "price": 100, "stock": 20})
	acct := "/api/admin/customers/" + buyerID

	// --- no credit limit: pay later refused ---
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(tv), "quantity": 2})
	must(t, 409, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"})
	must(t, 400, "POST", "/api/orders", buyer, J{"payment_method": "card"})

	// --- credit limit ---
	must(t, 400, "PUT", acct+"/credit-limit", admin, J{"credit_limit": -1})
	must(t, 403, "PUT", acct+"/credit-limit", buyer, J{"credit_limit": 1000})
	acc := obj(must(t, 200, "PUT", acct+"/credit-limit", admin, J{"credit_limit": 250}))
	if str(acc["available_credit"]) != "250" || acc["pay_later_allowed"] != true {
		t.Fatalf("account after setting the limit: %v", acc)
	}

	// --- pay later within the limit ---
	order1 := obj(must(t, 201, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"}))
	if order1["payment_method"] != "pay_later" || order1["payment_status"] != "deferred" {
		t.Fatalf("pay-later order: %v", order1)
	}
	st := obj(must(t, 200, "GET", "/api/debt", buyer, nil))
	if a := obj(st["account"]); str(a["balance"]) != "200" || str(a["available_credit"]) != "50" {
		t.Fatalf("balance 200 / available 50 expected: %v", a)
	}
	first := obj(arr(obj(st["transactions"])["items"])[0])
	if first["type"] != "charge" || str(first["amount"]) != "200" || str(first["balance_after"]) != "200" {
		t.Fatalf("checkout should add a charge: %v", first)
	}

	// --- over the limit ---
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(tv), "quantity": 1})
	over := obj(must(t, 409, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"}))
	if !strings.Contains(str(over["error"]), "available credit (50.00)") {
		t.Fatalf("should explain the available credit: %v", over)
	}

	// --- cancelling reverses the charge ---
	must(t, 200, "POST", "/api/orders/"+id(order1)+"/cancel", buyer, nil)
	st = obj(must(t, 200, "GET", "/api/debt", buyer, nil))
	rev := obj(arr(obj(st["transactions"])["items"])[0])
	if str(obj(st["account"])["balance"]) != "0" || rev["type"] != "reversal" || str(rev["amount"]) != "-200" {
		t.Fatalf("cancel should reverse the debt: %v", st)
	}

	// --- payments and adjustments ---
	order2 := id(must(t, 201, "POST", "/api/orders", buyer, J{"payment_method": "pay_later"})) // 100, now fits
	must(t, 409, "POST", acct+"/debt/payments", admin, J{"amount": 150})                       // more than owed
	must(t, 400, "POST", acct+"/debt/payments", admin, J{"amount": 0})
	paid := obj(must(t, 201, "POST", acct+"/debt/payments", admin, J{"amount": 60, "note": "cash at the shop", "order_id": n(order2)}))
	pay := obj(arr(obj(paid["transactions"])["items"])[0])
	if str(obj(paid["account"])["balance"]) != "40" || pay["type"] != "payment" || str(pay["amount"]) != "-60" ||
		str(pay["balance_after"]) != "40" || str(pay["created_by"]) != adminID {
		t.Fatalf("payment should lower the balance to 40: %v", paid)
	}
	must(t, 400, "POST", acct+"/debt/adjustments", admin, J{"amount": 10})                      // note required
	must(t, 409, "POST", acct+"/debt/adjustments", admin, J{"amount": -50, "note": "too much"}) // would go negative
	adj := obj(must(t, 201, "POST", acct+"/debt/adjustments", admin, J{"amount": 10, "note": "late fee"}))
	if str(obj(adj["account"])["balance"]) != "50" {
		t.Fatalf("adjustment should raise the balance to 50: %v", adj["account"])
	}

	// --- delivering a pay-later order keeps it deferred ---
	for _, s := range []string{"confirmed", "processing", "shipped", "delivered"} {
		must(t, 200, "PUT", "/api/admin/orders/"+order2+"/status", admin, J{"status": s})
	}
	if o := obj(must(t, 200, "GET", "/api/admin/orders/"+order2, admin, nil)); o["payment_status"] != "deferred" {
		t.Fatalf("pay-later order stays deferred after delivery: %v", o["payment_status"])
	}

	// --- cash order delivered without collecting the money -> debt ---
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(tv), "quantity": 1})
	cod := id(must(t, 201, "POST", "/api/orders", buyer, nil))
	must(t, 400, "PUT", "/api/admin/orders/"+cod+"/status", admin, J{"status": "confirmed", "payment_collected": false})
	for _, s := range []string{"confirmed", "processing", "shipped"} {
		must(t, 200, "PUT", "/api/admin/orders/"+cod+"/status", admin, J{"status": s})
	}
	delivered := obj(must(t, 200, "PUT", "/api/admin/orders/"+cod+"/status", admin, J{"status": "delivered", "payment_collected": false}))
	if delivered["payment_status"] != "deferred" {
		t.Fatalf("uncollected cash should be deferred: %v", delivered["payment_status"])
	}
	st = obj(must(t, 200, "GET", acct+"/debt", admin, nil))
	if str(obj(st["account"])["balance"]) != "150" {
		t.Fatalf("balance 50 + 100 uncollected = 150: %v", st["account"])
	}

	// --- admin overview ---
	list := obj(must(t, 200, "GET", "/api/admin/debts?outstanding=true", admin, nil))
	found := false
	for _, a := range arr(list["items"]) {
		if str(obj(a)["user_id"]) == buyerID && str(obj(a)["balance"]) == "150" {
			found = true
		}
	}
	if !found {
		t.Fatalf("customer with debt missing from the overview: %v", list)
	}
	must(t, 403, "GET", "/api/admin/debts", buyer, nil)
}
