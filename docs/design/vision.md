# Minedefence: Vision

## Pitch

Minedefence is a cooperative multiplayer side-scrolling game that blends factory building with tower defence. Players work together to design factory pipelines that turn raw materials into the items needed to grow the factory, keep their minions alive and hold off enemies. The work itself is done by autonomous minions whose work queue the players manage. Raw materials come from mining blocks out of the map, and every region mined open becomes a place enemies can spawn from. Among the factory's outputs are the components for defensive towers, which the minions build. Progression comes from manufacturing goods and selling them through an external economy in exchange for blueprints that unlock new recipes, buildings and towers.

## Pillars

**Every dig is a bet.** Mining is the only source of material and the only source of danger: the player must take exposed ground on purpose to gain anything. This rules out safe, infinite resource sources and enemy waves that arrive regardless of what the player did.

**Direct the work, don't do it.** Players plan, prioritise and lay out; minions execute. This rules out direct player-controlled avatars that mine, build or fight by hand.

**The factory is the defence.** Towers, ammunition, repairs and minion upkeep are all outputs of the production chain, so a well-designed pipeline is what holds the line. This rules out defences that are bought with abstract currency or that need no supply once placed.

**Legible at a glance.** Everything that matters is readable from the side-on view: what a minion is doing, where a pipeline is starved, where the ground is open and enemies can come from. This rules out hidden state, hidden spawn rules and information that only lives in menus.

**Better together, whole alone.** Co-op is the designed experience and splitting attention between mining, logistics and defence is the point, but one player must be able to play a complete session as a one-person team. This rules out roles, mechanics or content that only function with a minimum player count.

## Core loop

Moment to moment:

1. Queue mining jobs on the map to extract the blocks the factory needs next.
2. Minions dig; material flows in and the newly opened region becomes enemy-spawnable.
3. Route the material through the pipeline, placing and connecting buildings to make the items the factory, the minions and the defences need.
4. Spend part of the output on towers and supplies at the exposed edges; spend the rest on expanding the factory.
5. Enemies emerge from open regions and press the defences; the players reprioritise the work queue to repair, resupply and recover.

The central tension: every block mined is both material gained and ground opened. Players who dig too little starve the factory; players who dig too much cannot defend the frontier they created.

Session level:

1. Early: a small protected pocket, hand-sized production, a few towers.
2. Middle: pipelines widen, the mined-out frontier grows and defence becomes a logistics problem as much as a placement one.
3. Late: surplus goods are sold through the external economy for blueprints that open new tiers of recipes and towers, which demand deeper and more dangerous mining to feed.
4. The session ends when the players lose the factory or reach a progression goal the current design set as the finish line.

## Player fantasy

The players are the overseers of a subterranean industrial colony: part site foreman, part logistics engineer, part garrison commander. The best moment is watching a frontier that was about to be overrun get held because the pipeline you redesigned a minute ago is now delivering shells to the towers exactly as fast as they fire, while your minions, following the priorities you set, are already digging the next seam behind the line without being told. The feeling is of a plan coming together through others' hands.

## Systems map

Each line is expected to become `docs/design/mechanics/<system>.md`. **Slice** marks the smallest playable vertical slice; **Later** marks systems that follow once the slice is proven.

- `map-and-mining` (Slice): the side-scrolling block map, block types and yields, mining jobs, and how mined-out space becomes an open region.
- `minions-and-work-queue` (Slice): autonomous workers, the shared prioritised job queue, minion needs and what keeps them alive.
- `factory-pipelines-and-items` (Slice): buildings, recipes, item transport and the item catalogue, including the components that feed defence.
- `enemies-and-spawn-regions` (Slice): what counts as a spawnable region, spawn rules and pacing, enemy types and what they attack.
- `towers-and-defence` (Slice): tower types, placement, construction by minions, supply and repair.
- `multiplayer-cooperation` (Slice, minimal form): shared world, shared work queue and shared resources for two or more players; the full social layer comes later.
- `external-economy-and-blueprints` (Later): selling goods to an off-map market and the blueprints received in return.
- `progression-tree` (Later): the ordering and gating of blueprints into tiers of recipes, buildings and towers.

The vertical slice is one map, one minion type, a short pipeline (ore to a single tower component), one tower, one enemy type spawning from mined regions, and two players sharing the queue. It exists to prove the mining-versus-exposure tension is felt.

## Non-goals

- **Not competitive.** No PvP, no versus modes, no competitive leaderboards between players in a session.
- **Not a story game.** No procedural or authored narrative campaign; the fiction is a setting, not a plot.
- **Not 3D.** The world is a 2D side-scrolling block map; depth is not a gameplay axis.
- **Not a separate single-player campaign.** Solo play is a one-player co-op session with the same rules, not a first-class mode with its own content.
- **Not an avatar game.** Players have no in-world body to control; they act through the work queue and placement.
- **Not a real-money economy.** The external economy is an in-fiction market for blueprints, not monetisation.

## Open questions

- Resolved: target platform and distribution. Windows x64 only, with no storefront or browser build. See `docs/adr/0001-language-engine-and-project-layout.md`.
- Resolved: multiplayer topology. A host-authoritative listen server over TCP, with no relay and no dedicated server. See `docs/adr/0002-multiplayer-topology.md`.
- Resolved: player count per session. 1 to 4 including the host. See `docs/adr/0002-multiplayer-topology.md`. The designer confirms this range fits "Better together, whole alone".
- Session persistence: are worlds saved and resumed across sessions, or is a run a single sitting? Decides: designer, after producer confirms it fits the first iteration scope.
- Win condition for a session: a progression goal, survival for a duration, or open-ended? Decides: designer, to be settled before the progression-tree spec.
- Whether minions can die permanently and be replaced through the factory, or only be incapacitated. Decides: designer, in the minions spec.
- Art and audio direction for legibility (how enemies, open regions and starved pipelines are shown). Decides: designer in the feedback sections of each spec, with whoever owns rendering. Blocks, mining marks, exposed ground and open regions are settled in the Feedback section of `mechanics/map-and-mining.md`. Enemies and starved pipelines remain open.
