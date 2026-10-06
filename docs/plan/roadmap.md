# Roadmap

Owner: producer. Milestones in order. Each milestone ends in a build that can be
run and played. A milestone may take more than one iteration. Iteration plans live
in `docs/plan/iterations/`.

The first target is the vertical slice in `docs/design/vision.md`: one map, one
minion type, a short pipeline (ore to a single tower component), one tower, one
enemy type spawning from mined regions, and two players sharing the queue. It
exists to prove that the mining-versus-exposure tension is felt.

## Vertical slice

### M1. Dig (iteration 001, open)

Solo. A generated map in a window. The player marks and cancels blocks, minions
dig them out, and exposed ground and open regions are drawn distinctly from the
home pocket. Yields land as loose items on the mined cell and stay there.

Specs: `map-and-mining`, minimal part of `minions-and-work-queue`.

### M2. Debug bridge and haul

- MCP debug bridge Part 1 (`cmd/mdbridge`, headless), per ADR 0003. Needs the
  ADR 0003 open question (official MCP Go SDK or hand-rolled JSON-RPC) settled
  by the architect first. Depends on `Step(cmds)`, commands and read-only
  accessors from M1.
- `factory-pipelines-and-items` spec, including whether the slice hauls loose
  items and whether buildings cost stone.
- Item catalogue for the raw items, hauling of loose stacks, and the first
  building that accepts them.
- Minion work-queue priority, if the minions spec defines it.

### M3. Make

The short pipeline from ore to the single tower component, with placement and
connection of the slice buildings. Minion needs, if the minions spec puts them in
the slice.

### M4. Threat

`enemies-and-spawn-regions` spec and implementation: one enemy type spawning from
open regions, including the minimum spawn distance from mouth cells. MCP debug
bridge Part 2 (live game attach, pause, step, screenshot) lands here or earlier,
before visual checks of enemy legibility are needed.

### M5. Defend

`towers-and-defence`: one tower, built by minions from the tower component,
supplied and repaired from the pipeline. First complete solo loop: dig, make,
defend. First real playtest of the mining-versus-exposure tension and of the
`[map_and_mining]` defaults.

### M6. Together (vertical slice complete)

`multiplayer-cooperation` in its minimal form: host-authoritative listen server
(ADR 0002), two players sharing the map, marks and queue. Includes pending
(ghosted) marks for unconfirmed client commands, player-tinted marks for more
than one slot, snapshots and the balance-hash handshake.

## Later

Order to be set after the slice playtest.

- Sealing or reclaiming exposed ground (map-and-mining open question 1).
- Natural caves (map-and-mining open question 4).
- Session persistence (vision open question).
- Win condition (vision open question; needed before `progression-tree`).
- `external-economy-and-blueprints`.
- `progression-tree`.
- Full social layer for multiplayer.
- Mining audio and juice not taken in M1, if not picked up earlier.

## Shipped

Nothing yet.
