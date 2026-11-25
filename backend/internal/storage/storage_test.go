package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ecommerce-backend/internal/tenant"
)

func storeCtx(id int64) context.Context {
	return tenant.WithStore(context.Background(), tenant.Store{ID: id})
}

func TestFilesLiveInTheirStoresFolder(t *testing.T) {
	dir := t.TempDir()
	s, err := New(Config{Driver: "local", Dir: dir, PublicURL: "https://api.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	url, err := s.Save(storeCtx(7), ".png", "image/png", []byte("png"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://api.example.com/uploads/7/") {
		t.Fatalf("store 7's file should be in folder 7: %s", url)
	}
	name := strings.TrimPrefix(url, "https://api.example.com/uploads/7/")
	if _, err := os.Stat(filepath.Join(dir, "7", name)); err != nil {
		t.Fatalf("file not on disk: %v", err)
	}

	// Another store cannot delete it, even knowing the URL.
	if err := s.Delete(storeCtx(8), url); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "7", name)); err != nil {
		t.Fatal("store 8 deleted store 7's file")
	}
	// Its own store can.
	if err := s.Delete(storeCtx(7), url); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "7", name)); !os.IsNotExist(err) {
		t.Fatal("store 7 could not delete its own file")
	}
}

func TestOwnedKey(t *testing.T) {
	base := "https://api.example.com/uploads/"
	name := strings.Repeat("a", 32) + ".jpg"
	cases := []struct {
		ctx  context.Context
		url  string
		want bool
	}{
		{storeCtx(3), base + "3/" + name, true},
		{storeCtx(3), base + "4/" + name, false}, // another store's
		{storeCtx(1), base + name, true},         // old files belong to store 1
		{storeCtx(2), base + name, false},
		{storeCtx(3), base + "3/../4/" + name, false},           // no path tricks
		{storeCtx(3), base + "3/evil.html", false},              // only names Save makes
		{storeCtx(3), "https://cdn.other.com/3/" + name, false}, // not our storage
		{tenant.AsPlatform(context.Background()), base + "4/" + name, true},
	}
	for _, c := range cases {
		if _, ok := ownedKey(c.ctx, base, c.url); ok != c.want {
			t.Errorf("ownedKey(%s) = %v, want %v", c.url, ok, c.want)
		}
	}
}
