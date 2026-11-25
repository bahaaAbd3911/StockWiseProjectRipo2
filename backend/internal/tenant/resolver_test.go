package tenant

import "testing"

func TestNormalizeHost(t *testing.T) {
	cases := map[string]string{
		"Ahmad.MyPlatform.com":      "ahmad.myplatform.com",
		"ahmad.myplatform.com:8443": "ahmad.myplatform.com",
		"shop.example.com.":         "shop.example.com",
		"[::1]:8080":                "::1",
		"127.0.0.1:8080":            "127.0.0.1",
	}
	for in, want := range cases {
		if got := NormalizeHost(in); got != want {
			t.Errorf("NormalizeHost(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSlugFor(t *testing.T) {
	r := NewResolver(nil, "MyPlatform.com.", "")
	ok := map[string]string{"ahmad.myplatform.com": "ahmad", "omar-1.myplatform.com": "omar-1"}
	for host, want := range ok {
		if got, found := r.SlugFor(host); !found || got != want {
			t.Errorf("SlugFor(%q) = %q, %v", host, got, found)
		}
	}
	for _, host := range []string{"myplatform.com", "a.b.myplatform.com", "ahmadmyplatform.com", "evilmyplatform.com", "ahmad.myplatform.com.evil.com"} {
		if got, found := r.SlugFor(host); found {
			t.Errorf("SlugFor(%q) should not match, got %q", host, got)
		}
	}
	if _, found := NewResolver(nil, "", "").SlugFor("ahmad.myplatform.com"); found {
		t.Error("no platform domain configured: no subdomain stores")
	}
}
