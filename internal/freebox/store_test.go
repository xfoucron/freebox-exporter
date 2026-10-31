package freebox_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

func TestTokenStore(t *testing.T) {
	store := freebox.TokenStore{Path: filepath.Join(t.TempDir(), "nested", "token")}

	got, err := store.Load()
	if err != nil || got != "" {
		t.Fatalf("Load on missing file = %q, %v; want empty, nil", got, err)
	}

	if err := store.Save("secret"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err = store.Load()
	if err != nil || got != "secret" {
		t.Errorf("Load = %q, %v; want secret, nil", got, err)
	}

	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions = %v, want 0600", perm)
	}
}
