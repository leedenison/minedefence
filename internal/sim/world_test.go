package sim_test

import (
	"testing"

	"github.com/leedenison/minedefence/internal/balance"
	"github.com/leedenison/minedefence/internal/sim"
)

func TestStepAdvancesTick(t *testing.T) {
	w := sim.New(balance.Config{}, 1)
	for range 3 {
		w.Step()
	}
	if got := w.Tick(); got != 3 {
		t.Fatalf("Tick() = %d, want 3", got)
	}
}
