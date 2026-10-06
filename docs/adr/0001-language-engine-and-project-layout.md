# 0001: Language, engine, project layout and balance configuration

## Context

Minedefence is a 2D side-scrolling co-op factory and tower-defence game
(`docs/design/vision.md`). The user has fixed the platform as Windows x64 only,
with no storefront or browser build. The simulation must run and be tested
without a window, and must be deterministic for a given seed and command stream.
The designer needs one commented data file holding every tunable value.

## Decision

**Language and engine.** Go (version pinned by `go.mod`) with Ebitengine
v2 (`github.com/hajimehoshi/ebiten/v2`). The user chose this. On Windows
Ebitengine needs no cgo or C toolchain, so a stock Go install builds the game.

**Module path.** `github.com/leedenison/minedefence`, matching the git remote, so
`go install` and tooling work without `replace` directives.

**Layout.**

| Path | Role |
|---|---|
| `cmd/minedefence/` | `main`: parse flags, load balance, call `app.Run`. Thin. |
| `internal/app/` | Implements `ebiten.Game`. Wires session, sim, render and ui. |
| `internal/sim/` | The simulation. Headless and deterministic. |
| `internal/balance/` | Parses the balance file into a typed `Config`. |
| `internal/netplay/` | Session (host or client), replication and wire messages. |
| `internal/render/` | Ebitengine drawing of sim state. |
| `internal/ui/` | Panels, HUD, input mapping to sim commands. |
| `internal/archtest/` | Tests that enforce the dependency rules. |
| `assets/` | Data files and a Go package embedding them via `embed`. |

Changes from the layout first proposed:

- `internal/app` was added so `main` stays trivial and the game loop and wiring
  sit in a package that can be tested.
- `net` was renamed `netplay` so it does not shadow the standard library `net`.
- `sim` starts as **one package** with one file per system (mining, minions,
  queue, pipelines, enemies, towers). These systems refer to each other
  (minions build towers, enemies attack buildings), so separate packages
  would create import cycles or interface layers too early. A subpackage is
  split out only when its dependencies are strictly one-way.

**Balance configuration.** The file is `assets/balance.toml`, in TOML. It is
embedded into the executable through the `assets` package and parsed by
`internal/balance` using `github.com/BurntSushi/toml`. Decoding is strict: a key
with no matching `Config` field is an error. TOML supports comments, has a
formal spec, and the designer can edit it without knowing Go. Embedding means
the game ships as one `.exe` and every build of a given commit uses the same
values. The designer tunes values by editing the file and re-running
`go run ./cmd/minedefence`.

**Tooling.** The project uses Go's standard commands only: `go build`,
`go test`, `go vet` and `gofmt`. Staticcheck is the linter. It is pinned as a
`tool` dependency in `go.mod` and run with `go tool staticcheck`, so nothing
needs installing separately. Make and task runners are not used.

**Testing strategy.** Unit tests use the standard `testing` package. Simulation
tests run headless and are the main safety net. `internal/archtest` runs
`go list` to enforce the package dependency rules in `docs/architecture.md`
and fails the build if they are broken.

## Alternatives considered

- **C# with Godot 4.** It has a mature editor, built-in UI controls and
  high-level multiplayer. Rejected because the user prefers Go. Go tooling is
  also simpler: one toolchain, no editor or .NET SDK versions to keep aligned.
  In Go, package imports enforce the headless simulation boundary, while in
  Godot it would rest on discipline.
- **Other balance formats.** JSON has no comments, and JSON with comments
  (JSONC) lacks a standard Go parser. YAML has comments, but its implicit
  typing (`no` → false, `1.10` → 1.1) and indentation-sensitive syntax make
  designer errors likely. HCL is heavier and less familiar. Go source would put
  data inside the code the designer must not edit.
- **Loading balance from disk at runtime.** Rejected for now. It adds a path
  and a failure mode to every launch and lets host and client run different
  values. Revisit this if rebuild time slows tuning.
- **golangci-lint.** It is a separate binary with a large configuration
  surface. Staticcheck plus `go vet` covers the need with one pinned tool.

## Consequences

- There is no visual editor. Levels, layouts and tuning are data or code.
- The game is UI-heavy (work queue, build menus, pipeline status). Ebitengine
  has no widget toolkit, so the rendering-ux engineer must choose between
  `ebitenui` and a custom immediate-mode UI. That choice needs its own ADR
  before UI work starts.
- Netcode is written by hand (see ADR 0002).
- The ecosystem is smaller than Godot's or Unity's. Expect to write more tooling.
- Go is garbage collected. Hot simulation and render paths must avoid
  per-tick allocation: reuse slices, prefer value types and index-based
  entity storage, and benchmark when frame time matters.
- Every balance value needs a matching `Config` field. The designer can change
  values without engineering help but cannot add new keys alone.
