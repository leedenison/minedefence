// Package netplay runs a play session: the host-authoritative listen server,
// the client connection and the wire messages between them (ADR 0002).
//
// A Session owns the authoritative sim.World on the host (including solo
// play, which is a host with no remote players) and a replica on clients. It
// orders player commands into ticks, steps the world, and replicates state.
// Socket goroutines only move bytes; all sim access happens on the caller's
// goroutine.
//
// It may import sim and balance. It must not import Ebitengine, render, ui or
// app. Named netplay rather than net to avoid shadowing the standard library.
package netplay
