# Architecture

Read this before touching code. Decisions are recorded in `docs/adr/`. This
file states the current result. If the code and this file disagree, fix one of
them in the same change.

- Platform: Windows x64 only.
- Language and engine: Go and Ebitengine v2 (ADR 0001).
- Multiplayer: host-authoritative listen server for 1 to 4 players, over TCP (ADR 0002).

## Commands

Run all commands from the repository root. CI (`.github/workflows/ci.yml`,
windows-latest) runs every one except Run, on every push.

| Task | Command |
|---|---|
| Build | `go build ./...` |
| Test | `go test ./...` |
| Lint | `go vet ./...` and `go tool staticcheck ./...` |
| Format | `gofmt -w .` (CI fails if `gofmt -l .` lists any file) |
| Run | `go run ./cmd/minedefence` |

The work is finished only when build, format, lint and test all pass.

## Modules

Module path: `github.com/leedenison/minedefence`.

| Package | Owns | May import (in this module) |
|---|---|---|
| `cmd/minedefence` | Flags, startup, calling `app.Run` | anything |
| `internal/app` | `ebiten.Game`, the game loop, wiring session, render and ui | anything below |
| `internal/sim` | All game state and rules: map and mining, minions and the work queue, pipelines and items, enemies and spawn regions, towers | `balance` |
| `internal/balance` | The typed `Config` and strict parsing of the balance file | `assets` |
| `internal/netplay` | Session (host, client, solo), command ordering, snapshots and deltas, wire messages, TCP transport | `sim`, `balance` |
| `internal/render` | Drawing sim state with Ebitengine, sprites, camera | `sim`, `balance`, `assets` |
| `internal/ui` | HUD, panels, menus, mapping input to `sim.Command` | `sim`, `balance`, `assets`, `render` |
| `assets` | Embedded data files as bytes (`embed`) | nothing |
| `internal/archtest` | Tests that enforce this table | n/a |

Third-party dependencies: Ebitengine (only in `app`, `render`, `ui`), BurntSushi/toml
(only in `balance`), and staticcheck (tool only). Adding any other
dependency needs an ADR.

### Enforced rules

`internal/archtest/boundaries_test.go` runs under `go test ./...` and fails if:

- `sim`, `balance` or `assets` has any transitive dependency outside its allow
  list. In particular, `sim` can never reach Ebitengine, `netplay`, `render`,
  `ui` or `app`.
- `netplay` reaches Ebitengine, `render`, `ui` or `app`. Or `render` reaches
  `netplay`, `ui` or `app`. Or `ui` reaches `netplay` or `app`.
- `sim` directly imports `time`, `math/rand` (v1), `crypto/rand`, `os`, `net`,
  `sync`, `runtime`, `unsafe` or `syscall`.

Changing these rules needs an ADR, and the test and this file must change together.

## Simulation model

- **Single owner.** Only one goroutine touches a `sim.World`: Ebitengine's
  update goroutine. Network goroutines pass messages through channels.
- **Fixed tick.** `sim.TicksPerSecond = 20`. The session calls `World.Step` once
  every 3 Ebitengine updates (60 TPS). The simulation measures time only in
  ticks. Balance durations are given in seconds and converted to ticks once,
  at load time.
- **Determinism.** State after tick N depends only on the balance `Config`, the
  seed and the commands for ticks 1 to N. Randomness comes only from the
  `math/rand/v2` PCG source held in `World`. Never iterate over a map where
  order affects the outcome: sort the keys, or use slices and integer IDs.
  Prefer integer tile coordinates.
- **Interface to the simulation** (grown as the slice needs it, ADR 0002):
  - `sim.New(cfg, seed)` creates a world.
  - `Step(cmds []Command)` advances one tick. Commands are plain data and
    carry a player slot and sequence number. The skeleton's `Step()` takes
    no commands yet. It gains the parameter when the first command type is
    added.
  - Read-only accessors serve `render` and `ui`.
  - A per-tick change set serves rendering and network deltas.
  - Snapshot encode/restore and a state hash serve `netplay`.

  Presentation packages never mutate sim state. They issue commands.
- **Session.** Solo play is a host with no listener. Host and solo both step the
  authoritative world. Clients apply snapshots and deltas to a replica and never
  call `Step`.

## Balance configuration

- **File:** `assets/balance.toml`. Format: TOML. Owner: game-designer.
- Every key has a `What`/`Why`/`Range` comment block directly above it. The
  file header documents the naming conventions: one table per mechanics spec,
  snake_case keys, and units in key names.
- The file is embedded into the executable and parsed by `internal/balance` into
  `balance.Config`. Unknown keys are an error, so each new key needs a matching
  `Config` field. Adding one is a gameplay-engineer task.
- After any edit, run `go test ./...`. `TestShippedConfigParses` checks the file.
- In multiplayer, the host and clients must have identical balance files. The
  handshake compares hashes.

## Testing

- Write simulation tests as standard Go tests next to the code
  (`internal/sim/*_test.go`). Prefer the external test package `sim_test`, so
  tests use only the public interface.
- QA acceptance tests go in `internal/sim/acceptance_<system>_test.go`. Build a
  world from a seed and a `balance.Config`, feed it commands, step it, and
  assert on the state.
- A replay or regression test is a seed, a balance config and a command log,
  checked against a final state hash.
- `render` and `ui` test headless logic only, such as layout, input mapping
  and interpolation. Never test pixels.
- Test helpers that build a `balance.Config` in code must not stand in for the
  shipped file in balance-sensitive assertions. Load `assets.BalanceTOML` for
  those.
