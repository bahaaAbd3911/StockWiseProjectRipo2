package tests

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"ecommerce-backend/internal/services"
)

// ---------- step 3: store customization ----------

func TestStoreContentCustomization(t *testing.T) {
	requireServer(t)
	storeID, host := newStore(t, "decor", "active")
	admin := adminOn(t, host, storeID, "decor-admin@example.com")

	// Banner images are uploaded into the store's own folder.
	status, img := uploadOn(t, host, "/api/admin/settings/images", admin, "hero.png", pngBytes(t, 40, 20))
	heroURL := str(img["url"])
	if status != 201 || !strings.HasPrefix(heroURL, testPublicURL+"/uploads/"+strconv.FormatInt(storeID, 10)+"/") {
		t.Fatalf("banner image should be stored under the store's folder: %d %v", status, img)
	}

	content := J{
		"hero": J{"title": "Welcome to Decor", "titleAr": "أهلاً", "subtitle": "Lamps and vases", "highlight": "for every room",
			"buttonText": "Browse", "image": heroURL},
		"features": []J{{"icon": "fa-solid fa-truck", "name": "Fast delivery", "comment": "In Amman"}},
		"footer":   J{"tagline": "Light up your home", "social": J{"instagram": "https://instagram.com/decor"}},
	}
	mustOn(t, host, 200, "PUT", "/api/admin/settings", admin, J{"storeName": "Decor", "content": content})
	got := obj(mustOn(t, host, 200, "GET", "/api/settings", "", nil))
	if obj(obj(got["content"])["hero"])["title"] != "Welcome to Decor" || obj(obj(obj(got["content"])["footer"])["social"])["instagram"] == nil {
		t.Fatalf("content should be stored and returned to visitors: %v", got)
	}
	if c := obj(must(t, 200, "GET", "/api/settings", "", nil))["content"]; c != nil && obj(obj(c)["hero"])["title"] == "Welcome to Decor" {
		t.Fatal("another store must not show this store's content")
	}

	// Unsafe values are refused.
	for name, bad := range map[string]J{
		"script image":    {"hero": J{"image": "javascript:alert(1)"}},
		"html in icon":    {"features": []J{{"icon": `fa-solid fa-x" onclick="alert(1)`, "name": "x"}}},
		"unknown network": {"footer": J{"social": J{"myspace": "https://myspace.com/x"}}},
		"bad social link": {"footer": J{"social": J{"x": "javascript:alert(1)"}}},
		"too many tiles":  {"features": []J{{"icon": "fa-solid fa-a", "name": "1"}, {"icon": "fa-solid fa-a", "name": "2"}, {"icon": "fa-solid fa-a", "name": "3"}, {"icon": "fa-solid fa-a", "name": "4"}, {"icon": "fa-solid fa-a", "name": "5"}, {"icon": "fa-solid fa-a", "name": "6"}, {"icon": "fa-solid fa-a", "name": "7"}, {"icon": "fa-solid fa-a", "name": "8"}, {"icon": "fa-solid fa-a", "name": "9"}}},
	} {
		if s, res := callOn(t, host, "PUT", "/api/admin/settings", admin, J{"content": bad}); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}

	// Replacing the banner image deletes the old file.
	_, img2 := uploadOn(t, host, "/api/admin/settings/images", admin, "hero2.png", pngBytes(t, 30, 20))
	content["hero"].(J)["image"] = str(img2["url"])
	mustOn(t, host, 200, "PUT", "/api/admin/settings", admin, J{"storeName": "Decor", "content": content})
	if r, _ := fetchUpload(t, heroURL); r.StatusCode != 404 {
		t.Fatal("the replaced banner image should be deleted")
	}
	if r, _ := fetchUpload(t, str(img2["url"])); r.StatusCode != 200 {
		t.Fatal("the new banner image should be served")
	}

	// Customers cannot change settings or upload banners.
	customer := customerOn(t, host, storeID, "decor-buyer@example.com")
	mustOn(t, host, 403, "PUT", "/api/admin/settings", customer, J{"storeName": "Hacked"})
	if s, _ := uploadOn(t, host, "/api/admin/settings/images", customer, "x.png", pngBytes(t, 5, 5)); s != 403 {
		t.Fatal("customers cannot upload banner images")
	}
}

// multipartImage builds a form with the file in field "image".
func multipartImage(filename string, data []byte) (*bytes.Buffer, string) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("image", filename)
	fw.Write(data)
	mw.Close()
	return &body, mw.FormDataContentType()
}

// uploadOn uploads a file to a store's host.
func uploadOn(t *testing.T, host, path, token, filename string, data []byte) (int, J) {
	t.Helper()
	body, contentType := multipartImage(filename, data)
	req, _ := http.NewRequest("POST", server.URL+path, body)
	req.Host = host
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out J
	_ = jsonDecode(res.Body, &out)
	return res.StatusCode, out
}

func TestUploadsArePerStore(t *testing.T) {
	requireServer(t)
	storeID, host := newStore(t, "uploadsplit", "active")
	admin := adminOn(t, host, storeID, "split-admin@example.com")
	product := id(mustOn(t, host, 201, "POST", "/api/products", admin, J{"name": "Cup", "name_ar": "Cup AR", "price": 3, "stock": 1}))
	status, img := uploadOn(t, host, "/api/admin/products/"+product+"/images/upload", admin, "cup.png", pngBytes(t, 10, 10))
	url := str(img["image_url"])
	folder := strconv.FormatInt(storeID, 10)
	if status != 201 || !strings.Contains(url, "/uploads/"+folder+"/") {
		t.Fatalf("product image should be in the store's folder: %d %v", status, img)
	}
	name := url[strings.LastIndex(url, "/")+1:]
	if _, err := os.Stat(filepath.Join(uploadDir, folder, name)); err != nil {
		t.Fatalf("file should be on disk in folder %s: %v", folder, err)
	}

	// The main store's admin puts that URL on one of its own products and
	// deletes it: the other store's file must survive.
	mainAdmin := makeAdmin(t, "split-main-admin@example.com")
	mainProduct := newProduct(t, mainAdmin, "Thief Cup", 1)
	stolen := obj(must(t, 201, "POST", "/api/products/"+mainProduct+"/images", mainAdmin, J{"image_url": url}))
	must(t, 200, "DELETE", "/api/product-images/"+id(stolen), mainAdmin, nil)
	if r, _ := fetchUpload(t, url); r.StatusCode != 200 {
		t.Fatal("a store deleted another store's uploaded file")
	}
}

// ---------- step 4: custom domains ----------

func TestCustomDomains(t *testing.T) {
	requireServer(t)
	storeID, host := newStore(t, "domainshop", "active")
	admin := adminOn(t, host, storeID, "domain-admin@example.com")
	mustOn(t, host, 200, "PUT", "/api/admin/settings", admin, J{"storeName": "Domain Shop"})

	// Stand in for DNS: which TXT records exist.
	txt := map[string][]string{}
	realLookup := services.LookupTXT
	services.LookupTXT = func(_ context.Context, name string) ([]string, error) { return txt[name], nil }
	defer func() { services.LookupTXT = realLookup }()

	for name, bad := range map[string]string{
		"not a domain":   "not a domain",
		"platform's own": "other." + testPlatformDomain,
		"an IP":          "203.0.113.5",
		"no top level":   "localhost",
	} {
		if s, res := callOn(t, host, "POST", "/api/admin/domains", admin, J{"domain": bad}); s != 400 {
			t.Errorf("%s: want 400, got %d %v", name, s, res)
		}
	}
	d := obj(mustOn(t, host, 201, "POST", "/api/admin/domains", admin, J{"domain": "https://Shop.Example.TEST/"}))
	dns := obj(d["dns"])
	if d["domain"] != "shop.example.test" || d["verified"] != false ||
		dns["txt_name"] != "_stockwise-verify.shop.example.test" || !strings.HasPrefix(str(dns["txt_value"]), "stockwise-verify=") ||
		dns["cname_target"] != "domainshop."+testPlatformDomain {
		t.Fatalf("added domain with DNS instructions: %v", d)
	}
	mustOn(t, host, 409, "POST", "/api/admin/domains", admin, J{"domain": "shop.example.test"})

	// Not verified yet: the domain serves nothing and gets no certificate.
	mustOn(t, "shop.example.test", 200, "GET", "/api/settings", "", nil) // falls back to the default store...
	if s := obj(mustOn(t, "shop.example.test", 200, "GET", "/api/settings", "", nil)); s["storeName"] == "Domain Shop" {
		t.Fatal("an unverified domain must not serve the store")
	}
	tls := func(domain string) int {
		s, _ := call(t, "GET", "/internal/tls-check?domain="+domain, "", nil)
		return s
	}
	if tls("shop.example.test") != 404 {
		t.Fatal("no certificate before verification")
	}
	mustOn(t, host, 409, "POST", "/api/admin/domains/"+id(d)+"/verify", admin, nil)  // no TXT record yet
	mustOn(t, host, 409, "POST", "/api/admin/domains/"+id(d)+"/primary", admin, nil) // must be verified first

	// The owner publishes the TXT record; verification succeeds.
	txt[str(dns["txt_name"])] = []string{"other=1", str(dns["txt_value"])}
	if v := obj(mustOn(t, host, 200, "POST", "/api/admin/domains/"+id(d)+"/verify", admin, nil)); v["verified"] != true {
		t.Fatalf("verification: %v", v)
	}
	if s := obj(mustOn(t, "shop.example.test", 200, "GET", "/api/settings", "", nil)); s["storeName"] != "Domain Shop" {
		t.Fatalf("a verified domain serves its store: %v", s)
	}
	if tls("shop.example.test") != 200 || tls("domainshop."+testPlatformDomain) != 200 || tls(platformHost) != 200 ||
		tls("random.example.org") != 404 || tls("nosuchstore."+testPlatformDomain) != 404 {
		t.Fatal("certificates only for hosts the platform serves")
	}

	// Main domain: emails link to it.
	mustOn(t, host, 200, "POST", "/api/admin/domains/"+id(d)+"/primary", admin, nil)
	mustOn(t, "shop.example.test", 201, "POST", "/api/register", "", J{
		"first_name": "Dana", "last_name": "M", "email": "dana@example.com", "password": "password123", "phone": "0791234567",
	})
	waitForMail(t, outbox, "dana@example.com", "Confirm your email", 1)
	if m := outbox.about("dana@example.com", "Confirm your email")[0]; !strings.Contains(m.Text, "http://shop.example.test/verify-email?token=") {
		t.Fatalf("emails should link to the main domain: %s", m.Text)
	}

	// Another store cannot see or remove it.
	otherID, other := newStore(t, "domainthief", "active")
	thief := adminOn(t, other, otherID, "thief-admin@example.com")
	if list := arr(mustOn(t, other, 200, "GET", "/api/admin/domains", thief, nil)); len(list) != 0 {
		t.Fatalf("another store's domains are visible: %v", list)
	}
	mustOn(t, other, 404, "DELETE", "/api/admin/domains/"+id(d), thief, nil)
	mustOn(t, other, 404, "POST", "/api/admin/domains/"+id(d)+"/verify", thief, nil)

	// Suspended stores get no certificates; removing the domain stops it serving.
	testDB.Exec(`UPDATE stores SET status = 'suspended' WHERE id = $1`, storeID)
	if tls("shop.example.test") != 404 {
		t.Fatal("no certificate for a suspended store")
	}
	testDB.Exec(`UPDATE stores SET status = 'active' WHERE id = $1`, storeID)
	mustOn(t, host, 200, "DELETE", "/api/admin/domains/"+id(d), admin, nil)
	if s := obj(mustOn(t, "shop.example.test", 200, "GET", "/api/settings", "", nil)); s["storeName"] == "Domain Shop" {
		t.Fatal("a removed domain must stop serving the store")
	}
}
