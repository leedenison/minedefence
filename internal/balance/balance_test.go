package balance_test

import (
	"testing"

	"github.com/leedenison/minedefence/assets"
	"github.com/leedenison/minedefence/internal/balance"
)

// TestShippedConfigParses guards the file the designer edits: run after every
// balance change.
func TestShippedConfigParses(t *testing.T) {
	if _, err := balance.Parse(assets.BalanceTOML); err != nil {
		t.Fatalf("assets/balance.toml: %v", err)
	}
}

func TestUnknownKeyRejected(t *testing.T) {
	if _, err := balance.Parse([]byte("not_a_real_key = 1\n")); err == nil {
		t.Fatal("expected error for unknown key")
	}
}
