# 001: First dig

## Goal

Open the game, see an underground map around your home pocket, mark blocks to dig,
and watch minions tunnel out while the ground they open is shown as exposed.

## In scope

Solo play only (one player, slot 0). Every item below must be in the build that QA
plays with `go run ./cmd/minedefence`.

Items are listed in the order they must happen. An item may start only when the
items it depends on are done.

### Prerequisites (before any engineering)

1. **Minions spec, slice-minimal part.** Owner: game-designer.
   Write `docs/design/mechanics/minions-and-work-queue.md`. For this iteration it
   must at least define, with an Acceptance checklist: how many minions exist at
   tick 0 and where they start, how a minion chooses and claims an `available`
   mining job, how it moves through open cells to a cell beside the target, how it
   works the job, and when it releases its claim, satisfying
   `map-and-mining.md` 3.7 items 1 to 4. Any balance keys it needs go in its
   Proposed balance keys table. The rest of the spec (needs, death, priority,
   other job kinds) may be written now but is not built in 001.
2. **Rulings.** Owner: architect. Runs in parallel with item 1.
   - Project-wide rule for converting balance seconds to ticks
     (`map-and-mining.md` open question 6). Spec assumes round-half-up, minimum 1.
   - Confirm the `sim.Command` shape (player slot, sequence number) and the
     change from `Step()` to `Step(cmds []Command)` described in
     `docs/architecture.md`.
   - Confirm that `ui` reads cell state through the same `sim` read-only
     accessors as `render`, for the drag preview (`map-and-mining.md` 4,
     Exposure preview).
   An ADR only if one of these is a new decision.

### Simulation

3. **Balance fields.** Owner: gameplay-engineer. After item 2.
   `balance.Config` fields for the 23 `[map_and_mining]` keys in
   `map-and-mining.md` Proposed balance keys, plus the minion keys from item 1,
   converted to ticks per the item 2 ruling.
4. **Balance values.** Owner: game-designer. After item 3.
   Add those keys to `assets/balance.toml` with `What`/`Why`/`Range` comments.
   `TestShippedConfigParses` passes.
5. **Map generation.** Owner: gameplay-engineer. Spec: `map-and-mining.md` 3.1
   to 3.3.
6. **Commands, mining jobs and mining work.** Owner: gameplay-engineer. Spec:
   `map-and-mining.md` 3.4 to 3.6 and 3.8. Includes the first command types
   (DesignateMine, CancelMine), `World.Step(cmds)`, and the loose stack of
   raw items left on a mined cell (count and item only; no catalogue).
7. **Minimal minion.** Owner: gameplay-engineer. Spec: the item 1 part of
   `minions-and-work-queue.md`, and `map-and-mining.md` 3.7.
8. **Open regions and events.** Owner: gameplay-engineer. Spec:
   `map-and-mining.md` 3.9, plus the `BlockMined` event from 3.6.
9. **Read-only state and state hash.** Owner: gameplay-engineer. Spec:
   `map-and-mining.md` 3.10, minion position and current job, and a state hash
   for the Determinism check.

### Presentation

Items 10 to 12 may start as soon as item 5 and the accessors for it exist.

10. **Map rendering.** Owner: rendering-ux-engineer. Spec: `map-and-mining.md` 4,
    Blocks and Home pocket and open regions. Distinct fill and pattern per block
    type, ore visible at the default zoom, warm home back-wall, cold exposed
    back-wall, frontier line on every mouth-to-home edge. Camera pan so the whole
    map can be reached.
11. **Marking input.** Owner: rendering-ux-engineer. Spec: `map-and-mining.md` 2
    and 4, Marks. Drag to mark, drag to cancel, issued as commands. Marks with
    outline and pick icon (slot 0 colour), waiting marks dashed and dimmed, crack
    overlays in four stages that remain after cancel. Exposure preview while
    dragging (block count, yield breakdown, low-opacity exposed tint).
12. **Hover and minions.** Owner: rendering-ux-engineer. Spec:
    `map-and-mining.md` 4. Block tooltip (type, yield, dig time in seconds),
    region outline and size on hovering an exposed cell, loose items visible on
    the mined cell, minions drawn at their position and visibly facing the block
    they are working.

### Close-out

13. **Review.** Owner: code-reviewer. Blockers resolved by the authoring engineer.
14. **QA.** Owner: qa-playtester. Acceptance tests in
    `internal/sim/acceptance_map_and_mining_test.go` and
    `internal/sim/acceptance_minions_and_work_queue_test.go`, a play session of
    the window build, and a report at `docs/qa/reports/001-first-dig.md`.

## Out of scope

- **Hauling and the item catalogue.** Yields stay on the mined cell; the pipelines
  spec has not decided hauling (`map-and-mining.md` open question 2). Roadmap M2.
- **Any building, recipe or use for stone.** Depends on the pipelines spec
  (open question 3). `stone_yield_count` keeps its default.
- **Minion needs, death, job priority and non-mining jobs.** Not needed to dig;
  the DesignateMine priority field waits for the minions spec to define it. M2 or
  later.
- **Multiplayer and netplay.** Solo only. Pending (ghosted) marks, more than one
  player tint, snapshots and the balance-hash handshake are M6.
- **MCP debug bridge (ADR 0003), both parts.** Its Part 1 depends on commands and
  accessors this iteration creates, and the SDK-or-hand-rolled question is not
  settled. Part 1 is M2; Part 2 is M4 or earlier.
- **Audio and juice.** Dig and break sounds, region rumble, debris burst, region
  pulse and a full dig animation. Legibility first; picked up in a later
  iteration.
- **Camera zoom.** Only the default zoom is needed for the spec's visibility
  rule.
- **Sealing exposed ground and natural caves.** Deferred by the spec (open
  questions 1 and 4).
- **Enemies, including minimum spawn distance from mouth cells.** For the
  enemies spec (open question 5). M4.
- **Session persistence, win condition.** Vision open questions, not needed to
  dig.

## Acceptance criteria

All copied from `docs/design/mechanics/map-and-mining.md` section 7, plus the
build gate from `docs/architecture.md`. With the item 7 minion in place, the
checks marked (M) are in scope.

**Build**

- [ ] `go build ./...`, `gofmt -l .` (empty), `go vet ./...`,
      `go tool staticcheck ./...` and `go test ./...` all pass.
- [ ] `go run ./cmd/minedefence` opens a window showing the map.

**Generation**

- [ ] After `sim.New(cfg, seed)`, the map is `map_width_blocks` ×
      `map_height_blocks`. Every border cell is bedrock and no interior cell is
      bedrock.
- [ ] The pocket rectangle in 3.3 step 3 is entirely Open(home). No other cell
      is open. There are 0 regions.
- [ ] Every interior cell that is not pocket and not ore is dirt when
      y < `dirt_band_bottom_depth_blocks`, and stone otherwise.
- [ ] No coal or iron-ore cell has a pocket distance ≤
      `pocket_ore_free_margin_blocks`.
- [ ] At least one coal cell and at least one iron-ore cell have a pocket
      distance ≤ `near_vein_max_distance_blocks`. Check this for at least 20
      different seeds.
- [ ] Two worlds built from the same seed and config have identical cells. For
      at least one pair among seeds 1 to 5, the maps differ.

**Designation**

- [ ] A DesignateMine rectangle that covers bedrock, home cells, cells outside
      the map and solid blocks creates jobs only for the in-map, non-bedrock
      solid cells. `job_id`s increase in tile-index order.
- [ ] Repeating the same DesignateMine creates no new jobs.
- [ ] Player slots 0 and 1 designate the same cell in the same tick. Exactly
      one job exists, with `designated_by` = 0.
- [ ] A job on a solid cell that neighbours the pocket is `available`. A job one
      cell further out, with no open neighbour, is `waiting`.
- [ ] A diagonal neighbour of the pocket with no orthogonal open neighbour is
      `waiting`.
- [ ] CancelMine removes the jobs in its rectangle that player 1 created when
      player 0 issues it.

**Mining (M)**

- [ ] With a single minion working a dirt job, the cell is still solid after
      `mine_ticks(dirt)` - 1 ticks of work and is Open(exposed) after
      `mine_ticks(dirt)` ticks. Repeat for stone, coal and iron ore.
- [ ] Mining iron ore leaves exactly `iron_ore_yield_count` `iron_ore` items on
      that cell. The same holds for coal and stone. Mining dirt creates no
      items.
- [ ] A `BlockMined` event appears in the change set of the completion tick
      only.
- [ ] Cancel a job after k work ticks, where 0 < k < `mine_ticks`. The cell
      stays solid with `mine_progress_ticks` = k. After re-designation it
      completes after exactly `mine_ticks` - k further work ticks.
- [ ] Designate a two-cell tunnel A then B going outward from the pocket. B is
      `waiting` until the tick after A completes, then becomes `available`.
- [ ] Bedrock never gains a job and never changes state.

**Regions (M)**

- [ ] Dig two cells off the pocket wall that are not neighbours of each other.
      There are 2 regions, each of size 1, each with that cell as its only
      mouth cell, and with IDs equal to their tile indices. One
      `RegionOpened` event fires for each.
- [ ] Dig a connecting path between them outside the pocket. There is 1 region,
      its ID is the smaller of the two, its size is the total number of exposed
      cells, and a `RegionsMerged` event fires.
- [ ] Two exposed cells that are linked only through home cells remain two
      regions.
- [ ] No home cell ever becomes exposed, and the number of home cells never
      changes.

**Determinism**

- [ ] The same seed, config and command log give the same state hash at the
      final tick on two runs.

**Minions**

- [ ] The items of the `minions-and-work-queue.md` Acceptance checklist that
      cover the item 1 behaviour (starting minions, claim, movement, work,
      release) pass. Items covering anything listed under Out of scope are not
      checked in 001.

**Feedback (checked by playing the window build)**

From `map-and-mining.md` section 4, the in-scope subset:

- [ ] Each block type has a distinct fill and a distinct texture pattern, so the
      types can be told apart without colour.
- [ ] Ore is visible from the default zoom without hovering.
- [ ] Hovering any block shows a tooltip with its type, its yield and its dig
      time in seconds.
- [ ] A marked block has an outline and a pick icon, tinted in the player's
      colour. A waiting mark is dashed and dimmed. An available or claimed mark
      is solid.
- [ ] A block with progress greater than 0 shows crack overlays in four stages.
      Cracks remain after a cancel.
- [ ] While dragging a designation rectangle, the cursor shows the number of
      blocks to be marked and their yield, and the cells to be exposed are drawn
      in the exposed tint at low opacity.
- [ ] Yield items appear on the cell immediately when a block is mined.
- [ ] Home cells have a warm back-wall; exposed cells a cold, darker back-wall
      with a subtle pattern, always drawn, with no toggle.
- [ ] A frontier line is drawn along every edge between a mouth cell and a home
      cell.
- [ ] Hovering an exposed cell outlines its whole region and shows its size.
- [ ] A minion working a block is visibly at the block and facing it.

## Risks

- **Minions spec not ready, or larger than expected.** Engineering of items 7 and
  12 (minion part) cannot start. Fallback: ship 001 as generation, marking,
  cancel and waiting/available states in the window, with the (M) checks and the
  Minions criteria moved to 002.
- **Rounding ruling differs from the spec's assumption.** The spec says it
  follows the project rule, so no redesign is needed. Fallback: none required;
  QA uses the ruled formula for `mine_ticks`.
- **Region recompute or pathing too slow at 96 × 64 on a 20 Hz tick.** Fallback:
  the gameplay-engineer reports it, and the iteration ships with the map defaults
  unchanged; any change to map size goes to the designer, not the plan.
- **No input binding is specified for mark versus cancel.** The
  rendering-ux-engineer chooses one and records it in the report for the designer
  to review.
- **The spec and vision changes are uncommitted in the working tree.** If they
  change again mid-iteration, engineering builds against a moving target.
  Fallback: work to the version committed before item 3 starts.

## Outcome

