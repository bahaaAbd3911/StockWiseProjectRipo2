package tests

import (
	"context"
	"strings"
	"testing"

	"ecommerce-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var platformHost = "admin." + testPlatformDomain

// platformLogin creates a platform admin (as the command line tool does) and logs in.
func platformLogin(t *testing.T, email string) string {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("platform-pass-1"), bcrypt.MinCost)
	if _, err := repository.NewPlatformRepository(testDB).CreateAdmin(context.Background(), email, "Plat", "Form", string(hash)); err != nil {
		t.Fatal(err)
	}
	return str(obj(mustOn(t, platformHost, 200, "POST", "/api/platform/login", "",
		J{"email": email, "password": "platform-pass-1"}))["token"])
}

func TestPlatformAPIOnlyOnPlatformHost(t *testing.T) {
	requireServer(t)
	pt := platformLogin(t, "root1@platform.test")
	creds := J{"email": "root1@platform.test", "password": "platform-pass-1"}

	must(t, 404, "POST", "/api/platform/login", "", creds) // the main store's host
	_, storeHost := newStore(t, "hostguard", "active")
	mustOn(t, storeHost, 404, "GET", "/api/platform/stores", pt, nil)
	mustOn(t, platformHost, 404, "GET", "/api/products", "", nil) // the platform host is not a store
	mustOn(t, platformHost, 401, "POST", "/api/platform/login", "", J{"email": "root1@platform.test", "password": "wrong-pass"})

	// Tokens do not cross: a store admin's token is useless on the platform
	// API, and a platform token is useless on a store.
	storeAdmin := makeAdmin(t, "plat-store-admin@example.com")
	mustOn(t, platformHost, 401, "GET", "/api/platform/stores", storeAdmin, nil)
	must(t, 401, "GET", "/api/admin/orders", pt, nil)
	mustOn(t, platformHost, 401, "GET", "/api/platform/stores", "", nil)
	if me := obj(mustOn(t, platformHost, 200, "GET", "/api/platform/me", pt, nil)); me["email"] != "root1@platform.test" {
		t.Fatalf("me: %v", me)
	}
}

func TestPlatformCreatesReadyStore(t *testing.T) {
	requireServer(t)
	pt := platformLogin(t, "root2@platform.test")

	// Validation.
	for name, body := range map[string]J{
		"reserved slug": {"slug": "admin", "name": "X", "admin": J{"first_name": "A", "last_name": "B", "email": "a@x.com"}},
		"bad slug":      {"slug": "Ahmad Store!", "name": "X", "admin": J{"first_name": "A", "last_name": "B", "email": "a@x.com"}},
		"taken slug":    {"slug": "main", "name": "X", "admin": J{"first_name": "A", "last_name": "B", "email": "a@x.com"}},
		"no admin":      {"slug": "noadmin", "name": "X"},
		"bad timezone":  {"slug": "badtz", "name": "X", "timezone": "Mars/Base", "admin": J{"first_name": "A", "last_name": "B", "email": "a@x.com"}},
	} {
		if s, res := callOn(t, platformHost, "POST", "/api/platform/stores", pt, body); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}

	// Create: store + settings + admin, invitation emailed.
	res := obj(mustOn(t, platformHost, 201, "POST", "/api/platform/stores", pt, J{
		"slug": "Omar-Shop ", "name": "Omar Shop", "currency": "jod",
		"admin": J{"first_name": "Omar", "last_name": "K", "email": "Omar@Example.com"},
	}))
	st := obj(res["store"])
	host := "omar-shop." + testPlatformDomain
	if st["slug"] != "omar-shop" || st["currency"] != "JOD" || st["status"] != "active" ||
		st["url"] != "http://"+host || st["admin_email"] != "omar@example.com" {
		t.Fatalf("created store: %v", st)
	}
	if s := obj(mustOn(t, host, 200, "GET", "/api/settings", "", nil)); s["storeName"] != "Omar Shop" {
		t.Fatalf("the new store has its own settings: %v", s)
	}
	mustOn(t, host, 200, "GET", "/api/products", "", nil) // live right away

	invite := waitForMail(t, outbox, "omar@example.com", "Your store is ready", 1)
	msgs := outbox.about("omar@example.com", "Your store is ready")
	if !strings.Contains(msgs[0].Text, "http://"+host+"/reset-password?token=") || msgs[0].FromName != "Omar Shop" {
		t.Fatalf("the invite links to the store's own address and is sent as the store: %q / %s", msgs[0].FromName, msgs[0].Text)
	}

	// Before choosing a password the admin cannot log in; the invite link
	// only works on the new store's host.
	creds := J{"email": "omar@example.com", "password": "omar-password1"}
	mustOn(t, host, 401, "POST", "/api/login", "", creds)
	must(t, 400, "POST", "/api/password/reset", "", J{"token": invite, "new_password": "omar-password1"})
	mustOn(t, host, 200, "POST", "/api/password/reset", "", J{"token": invite, "new_password": "omar-password1"})
	admin := str(obj(mustOn(t, host, 200, "POST", "/api/login", "", creds))["token"])
	must(t, 401, "POST", "/api/login", "", creds) // the account belongs to Omar's store only

	// The admin manages only their store.
	pid := id(mustOn(t, host, 201, "POST", "/api/products", admin, J{"name": "Omar Lamp", "name_ar": "Lamp AR", "price": 12, "stock": 3}))
	must(t, 404, "GET", "/api/products/"+pid, "", nil)

	// Customers sign up in the new store; their emails carry its name and links.
	mustOn(t, host, 201, "POST", "/api/register", "", J{
		"first_name": "Lina", "last_name": "S", "email": "lina@example.com", "password": "password123", "phone": "0791234567",
	})
	waitForMail(t, outbox, "lina@example.com", "Confirm your email", 1)
	confirm := outbox.about("lina@example.com", "Confirm your email")[0]
	if !strings.HasPrefix(confirm.Subject, "Omar Shop") || !strings.Contains(confirm.Text, "http://"+host+"/verify-email?token=") {
		t.Fatalf("customer emails are branded per store: %s / %s", confirm.Subject, confirm.Text)
	}

	// The platform list shows the numbers.
	list := obj(mustOn(t, platformHost, 200, "GET", "/api/platform/stores?search=omar", pt, nil))
	row := obj(arr(list["items"])[0])
	if str(list["total"]) != "1" || str(row["products_count"]) != "1" || str(row["customers_count"]) != "1" {
		t.Fatalf("store list: %v", list)
	}

	// Edit, suspend, reactivate.
	mustOn(t, platformHost, 200, "PUT", "/api/platform/stores/"+id(st), pt, J{"name": "Omar Store", "currency": "USD"})
	mustOn(t, platformHost, 200, "POST", "/api/platform/stores/"+id(st)+"/suspend", pt, nil)
	mustOn(t, host, 403, "GET", "/api/products", "", nil)
	mustOn(t, host, 403, "POST", "/api/login", "", creds)
	if s := obj(mustOn(t, platformHost, 200, "GET", "/api/platform/stores/"+id(st), pt, nil)); s["status"] != "suspended" || s["name"] != "Omar Store" {
		t.Fatalf("store after edit + suspend: %v", s)
	}
	mustOn(t, platformHost, 200, "POST", "/api/platform/stores/"+id(st)+"/activate", pt, nil)
	mustOn(t, host, 200, "POST", "/api/login", "", creds)

	// Re-sending the invite replaces the link.
	mustOn(t, platformHost, 200, "POST", "/api/platform/stores/"+id(st)+"/invite", pt, nil)
	second := waitForMail(t, outbox, "omar@example.com", "Your store is ready", 2)
	mustOn(t, host, 200, "POST", "/api/password/reset", "", J{"token": second, "new_password": "omar-password2"})
	mustOn(t, platformHost, 404, "POST", "/api/platform/stores/999999/suspend", pt, nil)
}
