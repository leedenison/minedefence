// Package app implements ebiten.Game and wires the other packages together:
// it owns the session (local or networked), advances the simulation on the
// fixed tick, forwards input from ui as commands and asks render to draw.
//
// app is the only package besides cmd/minedefence that may import every other
// internal package. No internal package may import app.
package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/leedenison/minedefence/internal/balance"
	"github.com/leedenison/minedefence/internal/sim"
)

const (
	screenWidth  = 1280
	screenHeight = 720

	// updatesPerTick is the number of Ebitengine updates (60 per second)
	// per simulation tick. 3 gives the 20 Hz tick set in ADR 0002.
	updatesPerTick = ebiten.DefaultTPS / sim.TicksPerSecond
)

// Game is the top-level ebiten.Game.
type Game struct {
	world   *sim.World
	updates int
}

// Update runs at a fixed ebiten.DefaultTPS and steps the simulation every
// updatesPerTick calls.
func (g *Game) Update() error {
	g.updates++
	if g.updates%updatesPerTick == 0 {
		g.world.Step()
	}
	return nil
}

// Draw is a placeholder until internal/render exists.
func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Minedefence")
}

// Layout returns a fixed logical resolution.
func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

// Run opens the window and blocks until it closes.
func Run(cfg balance.Config) error {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Minedefence")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(&Game{world: sim.New(cfg, 1)})
}
