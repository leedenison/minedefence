package sim

import (
	"math/rand/v2"

	"github.com/leedenison/minedefence/internal/balance"
)

// TicksPerSecond is the fixed simulation rate (ADR 0002). It is an
// architectural constant, not a balance value: durations in the balance
// configuration are converted to ticks when loaded.
const TicksPerSecond = 20

// World is the complete simulation state.
type World struct {
	cfg  balance.Config
	rng  *rand.Rand
	tick uint64
}

// New returns a world built from cfg whose randomness derives only from seed.
func New(cfg balance.Config, seed uint64) *World {
	return &World{
		cfg: cfg,
		rng: rand.New(rand.NewPCG(seed, 0)),
	}
}

// Tick returns the number of completed simulation steps.
func (w *World) Tick() uint64 { return w.tick }

// Step advances the simulation by one tick.
func (w *World) Step() {
	w.tick++
}
