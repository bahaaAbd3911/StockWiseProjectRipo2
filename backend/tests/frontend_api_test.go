package tests

import (
	"strings"
	"testing"
)

// The endpoints and response shapes the React frontend relies on.

func TestProductWithImagesAndColors(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "media-admin@example.com")

	p := obj(must(t, 201, "POST", "/api/products", admin, J{
		"name": "Kettle", "name_ar": "Kettle AR", "category": "home", "price": 25, "stock": 4,
		"images": []string{"https://cdn.example.com/k1.jpg", " https://cdn.example.com/k2.jpg ", "https://cdn.example.com/k1.jpg"},
		"colors": []J{{"name": "Red", "name_ar": "Red AR", "hex_code": "#FF0000"}, {"name": "Blue", "hex_code": "#0000FF"}},
	}))
	pid := id(p)
	if urls := arr(p["image_urls"]); len(urls) != 2 || p["primary_image_url"] != "https://cdn.example.com/k1.jpg" {
		t.Fatalf("images: first is primary, duplicates removed: %v / %v", urls, p["primary_image_url"])
	}
	if cols := arr(p["colors"]); len(cols) != 2 || obj(cols[1])["name_ar"] != "Blue" {
		t.Fatalf("colors (Arabic name defaults to the name): %v", cols)
	}

	// Lists carry colors and images too (the cart needs a color).
	found := false
	for _, it := range arr(obj(must(t, 200, "GET", "/api/products?limit=100", "", nil))["items"]) {
		if str(obj(it)["id"]) == pid {
			found = len(arr(obj(it)["colors"])) == 2 && len(arr(obj(it)["image_urls"])) == 2
		}
	}
	if !found {
		t.Fatal("the product list should include colors and image URLs")
	}

	// Update: change one color, drop the other, replace images; omitted lists are kept.
	upd := obj(must(t, 200, "PUT", "/api/products/"+pid, admin, J{
		"name": "Kettle", "name_ar": "Kettle AR", "category": "Home", "price": 25, "stock": 4,
		"colors": []J{{"name": "Red", "name_ar": "Ahmar", "hex_code": "#EE0000"}},
	}))
	if cols := arr(upd["colors"]); len(cols) != 1 || obj(cols[0])["hex_code"] != "#EE0000" || len(arr(upd["image_urls"])) != 2 {
		t.Fatalf("colors replaced, images kept: %v %v", cols, upd["image_urls"])
	}
	if upd["category"] != "home" {
		t.Fatalf("an existing category keeps its spelling: %v", upd["category"])
	}
	must(t, 400, "PUT", "/api/products/"+pid, admin, J{
		"name": "Kettle", "name_ar": "Kettle AR", "price": 25, "stock": 4, "images": []string{"not a url"},
	})
	must(t, 400, "PUT", "/api/products/"+pid, admin, J{
		"name": "Kettle", "name_ar": "Kettle AR", "price": 25, "stock": 4,
		"colors": []J{{"name": "Red", "hex_code": "#FF0000"}, {"name": "red", "hex_code": "#00FF00"}},
	})

	// Order items show the product image.
	register(t, "media-buyer@example.com")
	buyer := login(t, "media-buyer@example.com", "password123")
	addAddress(t, buyer)
	must(t, 200, "POST", "/api/cart/items", buyer, J{"product_id": n(pid), "selected_color": "Red"})
	order := obj(must(t, 201, "POST", "/api/orders", buyer, nil))
	if obj(arr(order["items"])[0])["product_image_url"] != "https://cdn.example.com/k1.jpg" {
		t.Fatalf("order item image: %v", arr(order["items"])[0])
	}
}

func TestCategories(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "cat-admin@example.com")
	register(t, "cat-customer@example.com")
	customer := login(t, "cat-customer@example.com", "password123")

	must(t, 403, "POST", "/api/admin/categories", customer, J{"name": "X"})
	cat := obj(must(t, 201, "POST", "/api/admin/categories", admin, J{"name": "Gaming", "name_ar": "Gaming AR", "icon": "fa-solid fa-gamepad"}))
	must(t, 409, "POST", "/api/admin/categories", admin, J{"name": "gaming"})
	must(t, 400, "POST", "/api/admin/categories", admin, J{"name": " "})

	console := id(must(t, 201, "POST", "/api/products", admin, J{"name": "Console", "name_ar": "Console AR", "category": "GAMING", "price": 300, "stock": 2}))
	lookup := func(list []any, name string) J {
		for _, c := range list {
			if obj(c)["name"] == name {
				return obj(c)
			}
		}
		return nil
	}
	pub := lookup(arr(must(t, 200, "GET", "/api/categories", "", nil)), "Gaming")
	if pub == nil || str(pub["product_count"]) != "1" || pub["icon"] != "fa-solid fa-gamepad" || pub["is_active"] != true {
		t.Fatalf("public category list: %v", pub)
	}

	// Renaming renames the products' category too.
	must(t, 200, "PUT", "/api/admin/categories/"+id(cat), admin, J{"name": "Video Games", "name_ar": "VG AR", "icon": "fa-solid fa-gamepad"})
	if p := obj(must(t, 200, "GET", "/api/products/"+console, "", nil)); p["category"] != "Video Games" || p["category_ar"] != "VG AR" {
		t.Fatalf("rename should cascade to products: %v %v", p["category"], p["category_ar"])
	}

	// Inactive categories are only in the admin list.
	must(t, 200, "PUT", "/api/admin/categories/"+id(cat), admin, J{"name": "Video Games", "is_active": false})
	if lookup(arr(must(t, 200, "GET", "/api/categories", "", nil)), "Video Games") != nil {
		t.Fatal("inactive categories are hidden from the public list")
	}
	if lookup(arr(must(t, 200, "GET", "/api/admin/categories", admin, nil)), "Video Games") == nil {
		t.Fatal("admins see inactive categories")
	}

	// Deleting hides its products.
	must(t, 200, "DELETE", "/api/admin/categories/"+id(cat), admin, nil)
	must(t, 404, "GET", "/api/products/"+console, "", nil)
	if p := obj(must(t, 200, "GET", "/api/admin/products/"+console, admin, nil)); p["is_active"] != false {
		t.Fatalf("products of a deleted category are deactivated, not deleted: %v", p["is_active"])
	}
	must(t, 404, "DELETE", "/api/admin/categories/"+id(cat), admin, nil)
}

func TestStoreSettings(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "settings-admin@example.com")
	register(t, "settings-customer@example.com")
	customer := login(t, "settings-customer@example.com", "password123")

	must(t, 200, "GET", "/api/settings", "", nil) // public, empty at first
	body := J{
		"storeName": "  My Shop ",
		"logo":      "data:image/png;base64,iVBORw0KGgo=",
		"colors":    J{"light": J{"text": "#17231D", "button": "#22C55E"}, "dark": J{"text": "#E8F5F0"}},
	}
	must(t, 403, "PUT", "/api/admin/settings", customer, body)
	saved := obj(must(t, 200, "PUT", "/api/admin/settings", admin, body))
	if saved["storeName"] != "My Shop" {
		t.Fatalf("saved settings: %v", saved)
	}
	got := obj(must(t, 200, "GET", "/api/settings", "", nil))
	if got["storeName"] != "My Shop" || obj(obj(got["colors"])["light"])["button"] != "#22C55E" || !strings.HasPrefix(str(got["logo"]), "data:image/png") {
		t.Fatalf("settings should be returned to everyone: %v", got)
	}

	for name, bad := range map[string]J{
		"script logo":  {"logo": "javascript:alert(1)"},
		"html logo":    {"logo": "data:text/html;base64,PHNjcmlwdD4="},
		"bad color":    {"colors": J{"light": J{"text": "red"}}},
		"unknown mode": {"colors": J{"sepia": J{"text": "#000000"}}},
		"css in key":   {"colors": J{"light": J{"text;}body{": "#000000"}}},
	} {
		if s, res := call(t, "PUT", "/api/admin/settings", admin, bad); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}
	// A logo of about 1 MB fits (bigger than the normal 1 MiB body limit).
	big := "data:image/png;base64," + strings.Repeat("A", 1_300_000)
	must(t, 200, "PUT", "/api/admin/settings", admin, J{"storeName": "My Shop", "logo": big})
	must(t, 400, "PUT", "/api/admin/settings", admin, J{"logo": "data:image/png;base64," + strings.Repeat("A", 1_600_000)})
}

func TestAdminUsers(t *testing.T) {
	requireServer(t)
	admin := makeAdmin(t, "users-admin@example.com")
	adminID := str(obj(must(t, 200, "GET", "/api/profile", admin, nil))["id"])
	register(t, "users-customer@example.com")
	customer := login(t, "users-customer@example.com", "password123")
	must(t, 403, "GET", "/api/admin/users", customer, nil)

	// Create: email counts as confirmed, role defaults to customer.
	created := obj(must(t, 201, "POST", "/api/admin/users", admin, J{
		"first_name": "Nadia", "last_name": "Haddad", "email": "Nadia@Example.com", "password": "password123",
	}))
	nid := id(created)
	if created["email"] != "nadia@example.com" || created["role"] != "customer" || created["email_verified"] != true || created["is_active"] != true {
		t.Fatalf("created user: %v", created)
	}
	must(t, 400, "POST", "/api/admin/users", admin, J{"first_name": "A", "last_name": "B", "email": "nadia@example.com", "password": "password123"})
	must(t, 400, "POST", "/api/admin/users", admin, J{"first_name": "A", "last_name": "B", "email": "x@example.com", "password": "short"})
	must(t, 400, "POST", "/api/admin/users", admin, J{"first_name": "A", "last_name": "B", "email": "y@example.com", "password": "password123", "role": "staff"})
	nadia := login(t, "nadia@example.com", "password123")

	// List and search.
	page := obj(must(t, 200, "GET", "/api/admin/users?search=haddad", admin, nil))
	if str(page["total"]) != "1" || obj(arr(page["items"])[0])["email"] != "nadia@example.com" {
		t.Fatalf("search: %v", page)
	}

	// Edit: a new password logs her out everywhere; blank keeps it.
	must(t, 200, "PUT", "/api/admin/users/"+nid, admin, J{"first_name": "Nadia", "last_name": "H", "email": "nadia@example.com", "password": "newpassword1"})
	must(t, 401, "GET", "/api/profile", nadia, nil)
	login(t, "nadia@example.com", "newpassword1")
	must(t, 200, "PUT", "/api/admin/users/"+nid, admin, J{"first_name": "Nadia", "last_name": "H", "email": "nadia@example.com", "role": "admin"})
	login(t, "nadia@example.com", "newpassword1")

	// Safety rules.
	must(t, 409, "PUT", "/api/admin/users/"+adminID, admin, J{"first_name": "A", "last_name": "B", "email": "users-admin@example.com", "role": "customer"})
	must(t, 409, "DELETE", "/api/admin/users/"+adminID, admin, nil)

	// Details: user, orders, favorites.
	custID := str(obj(must(t, 200, "GET", "/api/profile", customer, nil))["id"])
	product := newProduct(t, admin, "Lamp X", 5)
	must(t, 201, "POST", "/api/favorites", customer, J{"product_id": n(product)})
	addAddress(t, customer)
	must(t, 200, "POST", "/api/cart/items", customer, J{"product_id": n(product), "quantity": 2})
	must(t, 201, "POST", "/api/orders", customer, nil)
	d := obj(must(t, 200, "GET", "/api/admin/users/"+custID, admin, nil))
	if str(obj(d["user"])["orders_count"]) != "1" || str(obj(d["user"])["total_spent"]) != "40" ||
		len(arr(d["orders"])) != 1 || len(arr(d["favorites"])) != 1 {
		t.Fatalf("user details: %v", d)
	}

	// Delete: a user with orders is deactivated; one without is removed.
	msg := obj(must(t, 200, "DELETE", "/api/admin/users/"+custID, admin, nil))
	if !strings.Contains(str(msg["message"]), "deactivated") {
		t.Fatalf("users with orders are deactivated: %v", msg)
	}
	must(t, 401, "GET", "/api/profile", customer, nil)
	if u := obj(obj(must(t, 200, "GET", "/api/admin/users/"+custID, admin, nil))["user"]); u["is_active"] != false {
		t.Fatalf("still listed, as inactive: %v", u)
	}
	must(t, 200, "PUT", "/api/admin/users/"+nid, admin, J{"first_name": "Nadia", "last_name": "H", "email": "nadia@example.com", "role": "customer"})
	if m := obj(must(t, 200, "DELETE", "/api/admin/users/"+nid, admin, nil)); m["message"] != "user deleted" {
		t.Fatalf("a user without orders is deleted: %v", m)
	}
	must(t, 404, "GET", "/api/admin/users/"+nid, admin, nil)
}
