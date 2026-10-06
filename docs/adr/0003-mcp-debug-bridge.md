# 0003: MCP debug bridge for agent-driven testing

## Context

The development agents (QA playtester, engineers, designer) can already test the
simulation through `go test`, because the simulation is headless and
deterministic (ADR 0001, ADR 0002). That covers assertions that someone thought
to write in advance. It does not cover:

- **Exploratory play.** Driving a world step by step, looking at the result and
  deciding what to try next, which is how balance problems and edge cases are
  found.
- **Visual checks.** Looking at the HUD and the scene to judge readability and
  whether feedback matches the Feedback sections of the specs. Agents can read
  PNG images. `render` and `ui` still never test pixels in `go test`.
- **Bug reproduction.** Loading a seed, replaying a command log, pausing at a
  given tick and inspecting state.

The Model Context Protocol (MCP) lets Claude Code call tools exposed by a local
process. A bridge that exposes the game through MCP gives agents these
capabilities without screen-scraping the window.

Constraints from existing decisions:

- Only the Ebitengine update goroutine touches a `sim.World`. The simulation
  has no locks (ADR 0002).
- All player intent is a `sim.Command`. Presentation never mutates sim state.
- The simulation runs at 20 ticks per second. An agent takes seconds per
  action, so it cannot play in real time.
- `sim` may not import `os`, `net`, `sync`, `time` or Ebitengine.

## Decision

Build the bridge in two parts, in this order. Neither part starts until `sim`
has its first `Command` type, `Step(cmds []Command)` and read-only accessors.

### Part 1: headless bridge

A new binary, `cmd/mdbridge`, that runs an MCP server over stdio. It holds its
own `sim.World` in process. There is no window and no Ebitengine.

Tools (initial set, grown as the simulation grows):

| Tool | Effect |
|---|---|
| `new_world(seed)` | Create a world from the embedded balance file and a seed. |
| `step(ticks, commands)` | Apply `commands` on the next tick, then advance `ticks` ticks in total. |
| `get_state(filter)` | Return a JSON view of state, optionally limited to a region or entity kind. |
| `state_hash()` | Return the state hash used by `netplay`. |
| `snapshot()` / `restore(data)` | Save and restore a whole world, once `sim` provides snapshots. |
| `command_log()` | Return every command applied since `new_world`, in the replay format. |

A command log from the bridge can be pasted into a replay test, so anything an
agent finds by exploring can become a regression test.

### Part 2: live game bridge

The game gains a `-debug-bridge 127.0.0.1:PORT` flag, compiled only under the
build tag `debugbridge`. Release builds contain no bridge code. In multiplayer
it is refused on clients. Only solo or host sessions can use it.

The game listens on that port for a small request/response protocol. Requests
are JSON, one per line. `cmd/mdbridge -attach 127.0.0.1:PORT` connects to the
game and exposes the Part 1 tools plus:

| Tool | Effect |
|---|---|
| `pause()` / `resume()` | Stop and restart automatic ticking. |
| `step(ticks, commands)` | While paused, advance exactly `ticks` ticks. |
| `screenshot()` | Return the current frame as a PNG. |

Rules for the live bridge:

- **Threading.** A goroutine reads the socket and sends each request through a
  channel to `app`. `app` handles requests in `Update`, and handles screenshots
  in `Draw` by copying the screen image. No other goroutine touches the world.
- **Commands.** Bridge commands enter the same command path as UI input, under
  the host's player slot. Determinism and the command log are unchanged, and
  the bridge never mutates state directly.
- **No cheats.** The bridge adds no debug-only commands such as spawning
  enemies or granting items. Test setups use seeds, command logs and
  `restore`. Adding debug commands needs a new ADR.
- **Focus.** With the bridge enabled, the game calls
  `ebiten.SetRunnableOnUnfocused(true)` so it keeps updating while the agent's
  terminal has focus.

### Placement

- A new package, `internal/bridge`, holds the tool logic: creating and stepping
  a world, the JSON state views, and the request/response types shared by
  `app` and `cmd/mdbridge`. It may import `sim` and `balance`. It may not
  import Ebitengine, `render`, `ui`, `netplay` or `app`. JSON view structs live
  here, so `sim` types never carry encoding tags.
- The MCP protocol layer lives only in `cmd/mdbridge`. `internal/bridge` does
  not know about MCP, which keeps the Open question cheap to settle or reverse.
- The live listener lives in `app`, in files with the `debugbridge` build tag.

### Registration

The repository includes a project-scoped `.mcp.json` that registers
`go run ./cmd/mdbridge` as a stdio server, so every agent in the repository can
use it without any manual setup.

## Open question: MCP library or hand-rolled protocol

This ADR does not decide how `cmd/mdbridge` speaks MCP. The options are:

- **Official Go SDK (`github.com/modelcontextprotocol/go-sdk`).** It tracks
  protocol revisions, handles capability negotiation, and generates input
  schemas from Go structs. Costs: a new third-party dependency and its
  transitive dependencies, which need an ADR under `docs/architecture.md`
  rules. Its API may still change between versions.
- **Hand-rolled JSON-RPC over stdio, using `encoding/json` only.** Estimated
  at about 200 lines for `initialize`, `tools/list` and `tools/call`. No new
  dependency. Costs: protocol upgrades are manual, and input schemas are
  written by hand.

Settle this before the MCP layer is written. Both options affect only
`cmd/mdbridge`, and the dependency would be confined to that binary. When
settled, amend this section with the choice and the reason. If the SDK is chosen, also add it to the dependency list in
`docs/architecture.md` and to the rules in `internal/archtest`.

## Alternatives considered

- **`go test` only.** It stays the primary way to test. It does not allow
  exploratory play or visual checks.
- **A plain JSON command-line tool driven through the shell.** It needs no MCP
  and no dependency, and is a reasonable stopgap. It has to restart the world on
  every call unless it keeps state on disk, and it cannot attach to a running
  game.
- **Desktop computer-use (screenshots and simulated clicks).** It needs no game
  changes. It is slow, imprecise and cannot pause between ticks or read exact
  state.
- **A `netplay` client as the bridge.** It would reuse the real protocol, but
  clients cannot pause or step the host, and `netplay` does not exist yet.
  Revisit if `netplay` grows a spectator mode.

## Consequences

- When Part 1 is implemented, `docs/architecture.md` gains rows for
  `internal/bridge` and `cmd/mdbridge`, and `internal/archtest` gains rules
  for `internal/bridge`, in the same change.
- `sim` must provide read-only accessors detailed enough to build JSON views.
  Snapshot and restore tools wait for the snapshot work in ADR 0002.
- CI builds with and without the `debugbridge` tag, so the tagged code cannot
  rot.
- The live listener binds only to `127.0.0.1` and has no authentication.
  Anyone on the same machine who can reach the port can drive the game. This is
  acceptable for a development-only build.
- Screenshots cost agent context. Agents should prefer `get_state` and use
  `screenshot` only for visual judgements.
- The bridge cannot tell whether the game is fun. It supports playtesting and
  does not replace human playtests.
