// Package sim is the authoritative game simulation: the block map and mining,
// minions and the shared work queue, factory pipelines and items, enemies and
// spawn regions, and towers.
//
// Rules for this package (enforced by internal/archtest):
//   - It must run headless. It may import only the standard library,
//     internal/balance and their dependencies; never Ebitengine, netplay,
//     render, ui or app.
//   - It must be deterministic. State changes only in World.Step, from the
//     previous state, the commands for that tick and the seeded RNG held in
//     World. No wall-clock time, no global or unseeded randomness, no
//     goroutines, and no logic that depends on map iteration order.
//
// Each system lives in its own file (mining.go, minions.go, queue.go, ...)
// within this one package. Split a system into a subpackage only when its
// dependencies are strictly one-way.
package sim
