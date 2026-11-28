package tests

import (
	"encoding/json"
	"strings"
	"testing"
)

// Shipping fees: the store's rule (standard fee, per-city fees, free over an
// amount), the cart's estimate, the fee charged at checkout and kept on the
// order, and the validation of the settings.
func TestShippingFees(t *testing.T) {
	requireServer(t)
	admin := newAdmin(t, "ship-admin@example.com")

	// The tests share store 1's settings: put them back afterwards so other
	// tests keep free shipping.
	before := must(t, 200, "GET", "/api/settings", "", nil)
	t.Cleanup(func() {
		raw, _ := json.Marshal(before)
		var body J
		_ = json.Unmarshal(raw, &body)
		delete(body, "shipping")
		must(t, 200, "PUT", "/api/admin/settings", admin, body)
	})
	withShipping := func(shipping any) J {
		raw, _ := json.Marshal(before)
		var body J
		_ = json.Unmarshal(raw, &body)
		body["shipping"] = shipping
		return body
	}

	mug := product(t, admin, J{"name": "Ship Mug", "category": "ship-kitchen", "price": 10, "stock": 50})
	buyer := newCustomer(t, "ship-buyer@example.com") // default address in Amman

	// No shipping settings: free, as before shipping fees existed.
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(mug), "quantity": 2})
	cart := obj(must(t, 200, "GET", "/api/cart", buyer, nil))
	if num(t, cart["shipping"]) != 0 || num(t, cart["total"]) != 20 {
		t.Fatalf("no settings = free shipping: %v / %v", cart["shipping"], cart["total"])
	}

	// Bad settings are refused with the field named.
	for name, bad := range map[string]J{
		"negative fee":      {"fee": -1, "freeOver": nil, "cities": []J{}},
		"huge fee":          {"fee": 5000, "freeOver": nil, "cities": []J{}},
		"3 decimals":        {"fee": 1.234, "freeOver": nil, "cities": []J{}},
		"free over 0":       {"fee": 1, "freeOver": 0, "cities": []J{}},
		"city twice":        {"fee": 1, "freeOver": nil, "cities": []J{{"city": "Amman", "fee": 1}, {"city": "amman ", "fee": 2}}},
		"city without name": {"fee": 1, "freeOver": nil, "cities": []J{{"city": " ", "fee": 1}}},
	} {
		if s, res := call(t, "PUT", "/api/admin/settings", admin, withShipping(bad)); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}

	// The rule: 4.00 standard, Amman 2.00, free from 50.00.
	saved := obj(must(t, 200, "PUT", "/api/admin/settings", admin, withShipping(J{
		"fee": 4, "freeOver": 50, "cities": []J{{"city": " Amman ", "fee": 2}},
	})))
	if c := arr(obj(saved["shipping"])["cities"]); len(c) != 1 || obj(c[0])["city"] != "Amman" {
		t.Fatalf("city names are trimmed: %v", saved["shipping"])
	}
	if s := obj(obj(must(t, 200, "GET", "/api/settings", "", nil))["shipping"]); num(t, s["fee"]) != 4 {
		t.Fatalf("customers can read the shipping rule: %v", s)
	}

	// The cart: the default address's city, or the city the page asks about.
	cart = obj(must(t, 200, "GET", "/api/cart", buyer, nil))
	if num(t, cart["shipping"]) != 2 || cart["shipping_city"] != "Amman" || num(t, cart["total"]) != 22 ||
		num(t, cart["free_shipping_over"]) != 50 {
		t.Fatalf("cart for the default city: %v %v %v %v", cart["shipping"], cart["shipping_city"], cart["total"], cart["free_shipping_over"])
	}
	cart = obj(must(t, 200, "GET", "/api/cart?city=irbid", buyer, nil))
	if num(t, cart["shipping"]) != 4 || num(t, cart["total"]) != 24 {
		t.Fatalf("another city pays the standard fee: %v / %v", cart["shipping"], cart["total"])
	}
	cart = obj(must(t, 200, "GET", "/api/cart?city=AMMAN", buyer, nil))
	if num(t, cart["shipping"]) != 2 {
		t.Fatalf("cities match regardless of case: %v", cart["shipping"])
	}

	// Free from 50.00: the cart says so as soon as the items reach it.
	must(t, 200, "PUT", "/api/cart/items/"+id(arr(cart["items"])[0]), buyer, J{"quantity": 5})
	cart = obj(must(t, 200, "GET", "/api/cart?city=Irbid", buyer, nil))
	if num(t, cart["shipping"]) != 0 || num(t, cart["total"]) != 50 {
		t.Fatalf("free over 50: %v / %v", cart["shipping"], cart["total"])
	}
	must(t, 200, "PUT", "/api/cart/items/"+id(arr(cart["items"])[0]), buyer, J{"quantity": 2})

	// Checkout charges the delivery address's fee and the order keeps a copy.
	order := obj(must(t, 201, "POST", "/api/orders", buyer, nil))
	if num(t, order["subtotal_amount"]) != 20 || num(t, order["shipping_amount"]) != 2 || num(t, order["total_amount"]) != 22 {
		t.Fatalf("order totals: %v + %v = %v", order["subtotal_amount"], order["shipping_amount"], order["total_amount"])
	}

	// Changing the rule later does not change the order.
	must(t, 200, "PUT", "/api/admin/settings", admin, withShipping(J{"fee": 9, "freeOver": nil, "cities": []J{}}))
	again := obj(must(t, 200, "GET", "/api/orders/"+id(order), buyer, nil))
	if num(t, again["shipping_amount"]) != 2 || num(t, again["total_amount"]) != 22 {
		t.Fatalf("an order keeps the fee it was charged: %v", again)
	}

	// Another city's address pays the standard fee (no city fees now: 9.00).
	other := newCustomer(t, "ship-irbid@example.com")
	must(t, 200, "PUT", "/api/addresses/"+id(arr(must(t, 200, "GET", "/api/addresses", other, nil))[0]), other,
		J{"recipient_name": "C", "phone": "0791234567", "city": "Irbid", "area": "Center", "street": "Main St"})
	must(t, 200, "POST", "/api/cart/items", other, J{"product_id": n(mug), "quantity": 1})
	o2 := obj(must(t, 201, "POST", "/api/orders", other, nil))
	if num(t, o2["shipping_amount"]) != 9 || num(t, o2["total_amount"]) != 19 {
		t.Fatalf("standard fee: %v / %v", o2["shipping_amount"], o2["total_amount"])
	}

	// The order email lists the fee.
	placed := waitForEmails(t, "ship-irbid@example.com", "Order #"+id(o2)+" received", 1)[0]
	if !strings.Contains(placed.Text, "Shipping: 9.00") {
		t.Fatalf("the email should show the shipping fee:\n%s", placed.Text)
	}
}
