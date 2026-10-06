# Map and mining

Balance table: `[map_and_mining]` in `assets/balance.toml`. Names in `code font`
below are keys in that table. None of them exist in the file yet. They are listed
under "Proposed balance keys" at the end, and an engineer adds the matching
`Config` fields before the designer adds the keys.

## 1. Purpose

This system serves **Every dig is a bet**. Mining is the only source of raw
material in the game, and every block mined becomes exposed ground that enemies can
later spawn from. The player should feel a pull in two directions on every
designation: the ore is right there, and taking it opens one more tile of frontier.
It also serves **Direct the work, don't do it**, because players never mine
directly. They mark blocks and minions dig them. It serves **Legible at a glance**
as well, because exposed ground is always drawn distinctly from safe ground.

## 2. Player-facing rules

- The world is one rectangular side-on grid of blocks. The whole map is
  underground. Its outer edge is unbreakable bedrock.
- The colony starts in a small pre-dug **home pocket** near the top centre of the
  map. The home pocket is safe ground. It never grows, and enemies never spawn
  in it.
- Every other cell starts as a solid block. The slice has five block types:

  | Block | Mineable | Yields | Notes |
  |---|---|---|---|
  | Bedrock | No | Nothing | Map border only. |
  | Dirt | Yes, fast | Nothing | The upper band of the map. Digging it costs you exposure and gives no material. |
  | Stone | Yes, slow | Stone | Everything below the dirt band. Building material. |
  | Coal | Yes | Coal | In veins, shallower. |
  | Iron ore | Yes, slow | Iron ore | In veins, deeper. |

- No ore lies directly against the home pocket. At least one coal vein and one
  iron vein are always a short dig away, so every map can feed the pipeline.
- The player marks blocks for mining by dragging a rectangle. Every mineable
  solid block in the rectangle gets a mining job. Bedrock, already-open cells
  and already-marked blocks in the rectangle are ignored.
- The player can drag a rectangle to cancel marks. A cancelled block keeps any
  digging progress it already has. Marking it again continues from that point.
- A marked block can only be dug once it touches open ground on one of its four
  sides (left, right, up or down, not diagonally). A marked block buried deeper
  than that **waits**. It becomes diggable as soon as a neighbouring block is
  dug out. Marking a long tunnel therefore makes minions dig it in order, from
  the open end inwards.
- One minion digs one block at a time, standing in an open cell beside it. Each
  block type takes a fixed time to dig.
- When a block is dug out, its yield appears as loose items on that cell, and the
  cell becomes **exposed ground**.
- Exposed ground stays exposed for the rest of the session. The slice has no way
  to fill it back in or seal it.
- Connected exposed ground forms an **open region**. Two tunnels that leave the
  home pocket separately are two regions. They become one region when they are
  dug into each other. The home pocket does not join regions together.
- All players share all marks. Any player can cancel any mark.

## 3. Simulation rules

### 3.1 Coordinates and grid

- The map is `map_width_blocks` by `map_height_blocks` cells. W and H below
  refer to these values.
- Cells use integer coordinates (x, y). x runs from 0 at the left to W-1. y runs
  from 0 at the top to H-1, and increases downward. **Depth** means y.
- The **tile index** of (x, y) is `y*W + x`. It is used wherever a deterministic
  order or a canonical ID is needed.
- The four **neighbours** of a cell are (x-1,y), (x+1,y), (x,y-1) and (x,y+1).
  This spec never uses diagonal adjacency.

### 3.2 Cell state

Each cell is exactly one of:

- **Solid(type)**, where type is one of `bedrock`, `dirt`, `stone`, `coal` or
  `iron_ore`. A solid cell also holds `mine_progress_ticks`, a non-negative
  integer that starts at 0.
- **Open(origin)**, where origin is `home` or `exposed`.

The only transition is Solid(type) to Open(exposed), through a completed mining
job (3.6). No rule changes an open cell back to solid, changes `home` to `exposed`
or the reverse, or changes a solid cell's type.

### 3.3 Map generation

The map is built once in `sim.New` from the balance config and the world's seeded
RNG. The same seed and config must produce the same map. The engineer chooses the
exact order of RNG calls but must keep it stable, because a change alters every
replay hash.

1. **Border.** Every cell with x = 0, x = W-1, y = 0 or y = H-1 is
   Solid(bedrock). No other cell is bedrock.
2. **Base layer.** An interior cell with y < `dirt_band_bottom_depth_blocks` is
   Solid(dirt). Every other interior cell is Solid(stone).
3. **Home pocket.** The pocket is a rectangle `pocket_width_blocks` wide and
   `pocket_height_blocks` high. Its left column is
   `px = (W - pocket_width_blocks) / 2` (integer division) and its top row is
   `py = pocket_top_depth_blocks`. Every cell in the rectangle is Open(home).
4. **Ore-free margin.** The **pocket distance** of a cell is its Chebyshev
   distance to the nearest pocket cell. A cell with pocket distance
   ≤ `pocket_ore_free_margin_blocks` is never turned into ore.
5. **Guaranteed near veins.** Place one coal vein, then one iron-ore vein. Each
   vein's start cell is chosen uniformly from the cells that are currently
   Solid(dirt) or Solid(stone), whose pocket distance is greater than
   `pocket_ore_free_margin_blocks` and at most
   `near_vein_max_distance_blocks`. The ore's depth band (step 6) does not apply
   to these two veins.
6. **Remaining veins.** Place `coal_vein_count - 1` further coal veins and
   `iron_ore_vein_count - 1` further iron-ore veins. Each start cell is chosen
   uniformly from the cells that are currently Solid(dirt) or Solid(stone), with
   `<ore>_vein_min_depth_blocks ≤ y ≤ <ore>_vein_max_depth_blocks` and pocket
   distance greater than the margin. If that set is empty, skip the vein. Place
   all coal veins before any iron veins.
7. **Growing a vein.** Start at the vein's start cell and take
   `<ore>_vein_size_blocks` steps in total, counting the start cell as the first
   step. At each step, if the current cell is Solid(dirt) or Solid(stone) and is
   outside the ore-free margin, it becomes Solid(<ore>). Then move to one of the
   four neighbours, chosen uniformly. If the chosen neighbour is on the border,
   stay in place for that step. The walk may revisit cells and may cross other
   veins. A later vein overwrites only dirt and stone, so it never overwrites
   earlier ore. A vein can therefore end up smaller than its step count, but its
   start cell is always ore.
8. No cell outside the pocket is open at tick 0. There are no natural caves in
   the slice.

### 3.4 Mining jobs

A **mining job** targets one cell. It records:

- `job_id`: a world-unique integer, assigned in increasing order of creation and
  never reused.
- `target`: the cell (x, y).
- `designated_by`: the player slot that created it.
- `created_tick`: the tick it was created.
- `state`: one of `waiting`, `available` or `claimed`. Who claims a job and how
  is defined by the minions spec (see 3.7).

A cell has at most one mining job.

A job is **available** when its target has at least one open neighbour and no
minion has claimed it. It is **waiting** when the target has no open neighbour.
In the slice, every open cell connects to the home pocket, because the map has no
caves and no sealing. An open neighbour therefore means a minion can reach the
block. If a later system adds obstacles that block movement, the minions spec
must tighten "available" to "has an open neighbour a minion can reach".

### 3.5 Commands

Both commands carry the player slot and sequence number required by ADR 0002.
The host applies them in (slot, sequence) order at the start of the tick.

- **DesignateMine{x0, y0, x1, y1}.** The rectangle is inclusive. It is
  normalised so that x0 ≤ x1 and y0 ≤ y1, then clipped to the map. Cells are
  visited in tile-index order. A job is created for each cell that is solid, is
  not bedrock and has no job. Every other cell is skipped without an error. If
  two commands in the same tick cover the same cell, the first in
  (slot, sequence) order creates the job and the later one skips the cell. The
  minions spec will add a priority field to this command. Until then, every job
  has the default priority.
- **CancelMine{x0, y0, x1, y1}.** The rectangle is normalised and clipped the
  same way. Every mining job whose target lies in the rectangle is removed,
  whichever player created it. A minion that held a removed job releases it, as
  the minions spec defines. The target cell keeps its `mine_progress_ticks`.

### 3.6 Mining work and completion

- `mine_ticks(type)` = max(1, round_half_up(`<type>_mine_seconds` ×
  `sim.TicksPerSecond`)). It is computed once, at load time.
- A minion **works** a job on a tick when it holds the claim and stands on an
  open cell that is a neighbour of the target. Each tick of work adds exactly 1 to
  the target's `mine_progress_ticks`. At most one minion can work a given job.
- When `mine_progress_ticks` ≥ `mine_ticks(type)` at the end of the work phase,
  the block **completes** in that same tick:
  1. The cell becomes Open(exposed).
  2. A loose stack of `<type>_yield_count` items of the block's yield item is
     created on that cell. Dirt yields nothing. Stone yields `stone`, coal
     yields `coal` and iron ore yields `iron_ore`. The item catalogue and loose
     items belong to the factory-pipelines spec. This spec only names the three
     raw items and fixes the hand-off point as "on the mined cell".
  3. The job is removed and the minion's claim ends.
  4. A `BlockMined{x, y, type, by_slot}` event is recorded in the tick's change
     set. `by_slot` is the job's `designated_by`.
- Blocks that complete in the same tick are processed in ascending tile-index
  order.

### 3.7 What mining needs from the minions spec

The minions spec is not written yet. Mining needs exactly the following from it
and nothing more:

1. A minion can **claim** an `available` mining job. A claimed job cannot be
   claimed by a second minion. The minions spec owns how jobs are chosen and in
   what priority order.
2. A minion with a claim moves to an open cell that neighbours the target, and
   then works it as defined in 3.6.
3. A minion releases its claim when the job is cancelled, or for any reason the
   minions spec defines, such as needs or danger. A released job becomes
   `available` again, or `waiting` if it has no open neighbour. Progress stays on
   the cell.
4. No minion claims a `waiting` job.

### 3.8 Tick order (mining's part)

Within `World.Step`:

1. Apply commands, including DesignateMine and CancelMine.
2. Minion phase, owned by the minions spec: claims, movement and work. Work adds
   progress as defined in 3.6.
3. Complete blocks (3.6), in tile-index order.
4. Recompute open regions (3.9) and the waiting or available state of every job.

Because job states are recomputed at the end of the tick, a waiting job whose
neighbour was mined this tick is available to claim from the next tick onward.

### 3.9 Open regions

- An **open region** is a maximal set of `exposed` cells in which every cell can
  reach every other cell through exposed neighbours. `home` cells never belong to
  a region and never connect two regions.
- **Region ID**: the smallest tile index among the region's cells. IDs are
  recomputed every tick. When two regions merge, the merged region takes the
  smaller ID. Consumers must not assume an ID stays the same across a merge.
- **Region size**: the number of cells in the region.
- **Mouth cells**: the exposed cells in a region that neighbour a home cell. A
  region can have zero mouth cells only if a later system lets regions exist
  without a link to the pocket. In the slice every region has at least one.
- Events recorded in the tick's change set:
  - `RegionOpened{id}` when a completion creates a region that shares no cell
    with any region from the previous tick.
  - `RegionsMerged{into_id, from_ids}` when a completion joins two or more
    existing regions.
- This spec does not decide where enemies spawn. It only defines the regions
  that the enemies spec may spawn from.

### 3.10 Read-only state the simulation must expose

These serve render, ui, netplay and QA. The engineer chooses their form.

- Map width and height, and the home pocket rectangle.
- For each cell: its state (solid type or open origin), `mine_progress_ticks`
  and `mine_ticks` for its type, and its mining job if it has one (`job_id`,
  `designated_by`, `state`).
- The region ID of an exposed cell.
- The list of regions, each with its ID, size, cells and mouth cells.
- The events described in 3.6 and 3.9, in the per-tick change set.

## 4. Feedback

This is the contract with rendering. It must hold for every player, the host and
clients alike.

**Blocks**

- Each block type has a distinct fill and a distinct texture pattern, so the
  types can be told apart without colour (for colour-blind players). Bedrock is
  the darkest, with a hatched pattern that reads as "cannot dig". Dirt is
  lighter and plain. Stone is mid-grey with a blocky texture. Coal has black
  angular flecks. Iron ore has rust-coloured round flecks.
- Ore must be visible from the default zoom without hovering.
- Hovering any block shows a tooltip with its type, its yield and its dig time
  in seconds.

**Marks (mining jobs)**

- A marked block has an outline and a pick icon, tinted in the colour of the
  player who marked it.
- A **waiting** mark is dashed and dimmed. An **available** or claimed mark is
  solid. The player can see at a glance how far into a tunnel the digging can
  currently reach.
- A block with progress greater than 0 shows crack overlays in four stages, at
  0-25%, 25-50%, 50-75% and 75-100% of `mine_ticks`. Cracks remain after a
  cancel, so abandoned progress is visible.
- While a minion is working a block, the minion plays a dig animation facing
  the block, with a short per-type dig sound on a regular beat.
- When a client has sent a designation or cancel that the host has not yet
  confirmed, it draws the marks as ghosted (ADR 0002 "pending").
- **Exposure preview.** While the player drags a designation rectangle, the
  cursor shows the number of blocks that will be marked and the yield they will
  produce, for example "12 blocks: 3 iron ore, 0 coal, 9 nothing". The cells
  that will become exposed are drawn in the exposed tint at low opacity. This is
  the moment the bet is made, so the cost must be visible before the player
  commits.

**Completion**

- When a block is mined, its sprite breaks into a short debris burst with a
  per-type break sound. The yield items appear on the cell immediately.

**Home pocket and open regions**

- Open cells show a back-wall. Home cells have a warm, lit back-wall. Exposed
  cells have a cold, darker back-wall with a subtle pattern. Exposed cells are
  always drawn this way, with no toggle, because exposure must never be hidden
  information.
- A bright **frontier line** is drawn along every edge between a mouth cell and a
  home cell. These lines are the doors into the base.
- When a `RegionOpened` event occurs, the new region's cells pulse once and a low
  rumble sound plays. When regions merge, the merged region pulses once.
- Hovering an exposed cell outlines its whole region and shows the region's
  size.

## 5. Interactions

- **minions-and-work-queue** (`minions-and-work-queue.md`, not yet written).
  Mining jobs go into the shared queue. That spec owns priority, claiming,
  movement, and how a minion behaves when its claim is cancelled. Section 3.7
  lists what mining requires from it.
- **factory-pipelines-and-items** (`factory-pipelines-and-items.md`, not yet
  written). It receives the raw items `stone`, `coal` and `iron_ore` as loose
  stacks on mined cells, and owns how they reach buildings. Buildings occupy
  open cells. Placing a building on an exposed cell does not change the cell's
  exposure.
- **enemies-and-spawn-regions** (`enemies-and-spawn-regions.md`, not yet
  written). It consumes open regions (3.9): IDs, sizes, cells and mouth cells,
  plus the `RegionOpened` and `RegionsMerged` events. It decides spawn rules,
  including whether spawns must keep a minimum distance from mouth cells. That
  decision matters a great deal: if enemies can spawn on the very first cell dug
  next to the pocket, the first dig is punished instantly.
- **towers-and-defence** (`towers-and-defence.md`, not yet written). Towers
  occupy open cells. Mouth cells and the frontier line show where towers are
  needed.
- **multiplayer-cooperation** (`multiplayer-cooperation.md`, not yet written).
  Marks are shared, and any player can cancel any mark. The `designated_by` field
  exists only to tint marks, not to give ownership.

## 6. Open questions

1. **Can exposed ground be sealed or reclaimed?** For example, back-filling or
   building a wall that turns exposed cells back into safe ground. The slice says
   no, so the cost of digging is permanent. Long sessions may need a way to
   recover from over-digging. Decides: designer, after the first slice
   playtest. A seal mechanic needs its own spec.
2. **Does the slice need hauling?** This spec drops yields on the mined cell.
   If the slice does not carry them to buildings, raw material never reaches the
   pipeline. Decides: designer in factory-pipelines-and-items, with the producer
   for scope.
3. **Is stone needed in the slice?** It is included as construction material.
   If the pipelines spec makes buildings free in the slice, set `stone_yield_count`
   to 0 rather than removing the block. Decides: designer in
   factory-pipelines-and-items.
4. **Natural caves.** Pre-open pockets that are not linked to the base and
   become regions when dug into would make digging into the unknown a sharper
   bet. Deferred past the slice. Decides: designer.
5. **Minimum spawn distance from the pocket.** See the enemies interaction in
   section 5. Decides: designer in enemies-and-spawn-regions.
6. **Resolved: rounding when converting seconds to ticks.** The project-wide
   rule is round half up with a minimum of 1 tick, as this spec assumed, applied
   once at load by `balance.SecondsToTicks`. See Balance configuration in
   `docs/architecture.md`.

## 7. Acceptance

All of these run against a headless `sim.World`. Load `assets.BalanceTOML` for
any check that depends on a value (see `docs/architecture.md`, Testing). Write
`mine_ticks(type)` for the rounded tick count from 3.6. Checks marked (M) need a
minion as defined by the minions spec. Until that spec exists, they cannot be
run.

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

## Proposed balance keys

Table `[map_and_mining]`. These keys are not yet in `assets/balance.toml`. Unknown
keys are rejected, so an engineer must add the `Config` fields first. The designer
then adds the keys with full `What`/`Why`/`Range` comments.

| Key | Default | Safe range | What / why | Depends on |
|---|---|---|---|---|
| `map_width_blocks` | 96 | 48 to 256 | Map width. About two screens wide: enough room for 2 to 4 players to dig in different directions without the slice map turning into a long trek. | Vein counts. |
| `map_height_blocks` | 64 | 32 to 128 | Map depth. Leaves room for the deep iron band below the dirt. | `iron_ore_vein_max_depth_blocks` must be < this - 1. |
| `pocket_width_blocks` | 12 | 6 to 24 | Width of the safe home pocket. Fits a handful of slice buildings, so the factory must grow into exposed ground. | `pocket_ore_free_margin_blocks` |
| `pocket_height_blocks` | 5 | 3 to 10 | Height of the home pocket. | `dirt_band_bottom_depth_blocks` |
| `pocket_top_depth_blocks` | 6 | 2 to `dirt_band_bottom_depth_blocks` - `pocket_height_blocks` | Depth of the pocket's top row. Leaves dirt above, so digging upward is also an option. | `dirt_band_bottom_depth_blocks` |
| `pocket_ore_free_margin_blocks` | 3 | 1 to 8 | Ring around the pocket that holds no ore. The first material always costs some exposure. | `near_vein_max_distance_blocks` |
| `near_vein_max_distance_blocks` | 10 | margin + 2 to 20 | The first coal and iron veins start within this distance, so every seed can feed the pipeline early. | `pocket_ore_free_margin_blocks` |
| `dirt_band_bottom_depth_blocks` | 20 | `pocket_top_depth_blocks` + `pocket_height_blocks` + 2 to `map_height_blocks` / 2 | First row of stone. Fast, worthless dirt near the base, then slow, useful stone deeper down. | Pocket keys. |
| `dirt_mine_seconds` | 1.5 | 0.5 to 5 | Dig time for dirt. Short enough that early digging feels responsive (matches the file header example). | `stone_mine_seconds` |
| `stone_mine_seconds` | 4 | 1 to 10 | Dig time for stone. About 2.5x dirt, so going deep feels like a commitment. | `dirt_mine_seconds` |
| `coal_mine_seconds` | 3 | 1 to 10 | Dig time for coal. | Smelting rate (pipelines spec). |
| `iron_ore_mine_seconds` | 5 | 1 to 12 | Dig time for iron ore. The slowest block, because it is the scarce input to the tower component. | Tower component recipe (pipelines spec). |
| `stone_yield_count` | 1 | 0 to 3 | Stone items per block. Set it to 0 if the slice does not use stone. | Building costs (pipelines spec). |
| `coal_yield_count` | 1 | 1 to 4 | Coal items per block. | Smelter recipe. |
| `iron_ore_yield_count` | 1 | 1 to 4 | Iron ore items per block. | Smelter recipe. |
| `coal_vein_count` | 10 | 1 to 30 | Number of coal veins, including the guaranteed near vein. | Map size. |
| `coal_vein_size_blocks` | 8 | 3 to 20 | Random-walk steps per coal vein, which caps the vein's size. | |
| `coal_vein_min_depth_blocks` | 8 | 1 to `coal_vein_max_depth_blocks` | Shallowest start depth of a non-guaranteed coal vein. | |
| `coal_vein_max_depth_blocks` | 40 | min to `map_height_blocks` - 2 | Deepest start depth of a non-guaranteed coal vein. | `map_height_blocks` |
| `iron_ore_vein_count` | 8 | 1 to 30 | Number of iron veins, including the guaranteed near vein. Scarcer than coal. | Map size. |
| `iron_ore_vein_size_blocks` | 6 | 3 to 20 | Random-walk steps per iron vein. | |
| `iron_ore_vein_min_depth_blocks` | 14 | 1 to `iron_ore_vein_max_depth_blocks` | Shallowest start depth of a non-guaranteed iron vein. Deeper than coal, so the better material means more exposure. | |
| `iron_ore_vein_max_depth_blocks` | 62 | min to `map_height_blocks` - 2 | Deepest start depth of a non-guaranteed iron vein. | `map_height_blocks` |

These defaults are first guesses. The first playtest must check two things:
whether the near veins make the opening too safe or too slow, and how many cells
are exposed by the time the first tower component is made.
