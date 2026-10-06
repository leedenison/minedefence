// Package render draws the simulation with Ebitengine: the block map, open
// regions, minions, buildings, items, enemies and towers.
//
// It reads sim state and never mutates it. It may import Ebitengine, sim,
// balance and assets. It must not import netplay, ui or app.
package render
