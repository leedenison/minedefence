# 0002: Multiplayer topology, tick model and transport

## Status

Proposed (2026-10-06). Becomes Accepted when the user commits it.

## Context

Co-op is the designed experience, and solo must be a complete session (vision
pillar "Better together, whole alone"). The user decided on peer-to-peer with one
player hosting and holding authority, at most 4 players including the host, and
no Steam or relay service. Players act only through commands such as queueing
jobs, placing buildings and setting priorities. There is no avatar and no twitch
input, so latency of a few hundred milliseconds is acceptable.

## Decision

**Topology.** A listen server. The host's process runs the only authoritative
`sim.World`. Clients hold a replica that they display and never step. Solo play
is a host with the listener turned off and uses the same code path. The host
rejects a fifth connection at handshake. When the host leaves, the session ends.
There is no host migration.

**Tick.** The simulation runs at a fixed **20 ticks per second**
(`sim.TicksPerSecond`). Ebitengine updates at 60 per second, and the host's
session steps the world on every third update. The update count drives
simulation time, never the wall clock.

**Commands.** All player intent, including the host's, becomes a `sim.Command`:
plain data carrying the player slot and a per-player sequence number. The host
assigns each command received to the next tick. Within a tick it applies them in
order of (player slot, sequence). Each tick's command list is logged, so the
seed, the balance file and the command log together reproduce a session. This
log is the basis for replay and regression tests.

**Replication.** A joining client receives a full snapshot. After that, the host
sends one delta per tick containing the blocks and entities that changed, taken
from a per-tick change set that the simulation records. Once a second, a delta
also carries a hash of the full state. If the client's replica hash differs, it
requests a fresh snapshot. Clients do no prediction. The UI shows commands it
has sent but not yet seen confirmed as pending.

**Transport.** One TCP connection per client, with `TCP_NODELAY` set and
length-prefixed frames. Each message is encoded with `encoding/gob`. The
handshake exchanges a protocol version and a SHA-256 of the balance file, and
any mismatch rejects the client. The transport sits behind a small interface
inside `netplay`, so it can be replaced without touching `sim`.

**Threading.** Each connection has one reader goroutine and one writer
goroutine, which talk to the session over channels. Only the Ebitengine update
goroutine reads or writes simulation state. The simulation contains no locks.

**Joining (first slice).** Joining is set by command-line flags (`-host :port`,
`-join addr:port`). A lobby UI comes later.

## Alternatives considered

- **Dedicated server.** It needs hosting infrastructure and does not help solo
  play. The headless simulation keeps a headless host possible later.
- **Deterministic lockstep (sending only commands).** Bandwidth is tiny, but it
  needs bit-identical results on every machine. Each tick waits for the slowest
  peer, a mid-session join still needs a snapshot, and desyncs are hard to
  debug. Snapshots plus deltas tolerate those faults. Lockstep stays possible
  later because the simulation is deterministic and commands are already logged.
- **UDP with a reliability layer (`kcp-go`, `quic-go`).** It avoids TCP
  head-of-line blocking and enables UDP hole punching later. Rejected for the
  first slice because every message here must arrive in order anyway, TCP needs
  no extra dependency, and nobody has measured a stall problem yet. Prefer
  `quic-go` if latency spikes under packet loss are measured or NAT traversal
  is taken on.
- **Custom binary encoding or Protocol Buffers.** These give smaller messages
  and explicit versioning. `gob` is in the standard library, and both ends run
  the same build, which the handshake enforces. Revisit if bandwidth is too high.
- **Steam networking or a relay service.** Excluded by the user.

## Consequences

- **NAT gap.** Remote players can join only if the host forwards a TCP port, or
  if everyone is on the same LAN or an overlay VPN such as Tailscale or
  ZeroTier. Fixing this means hole punching, which requires a UDP transport,
  or a relay server. Both are future options and are not solved here.
- The simulation must provide: plain-data `Command` types,
  `Step(cmds []Command)`, a per-tick change set, snapshot encode/restore of the
  whole `World`, and a state hash. The gameplay engineer builds these as the
  slice needs them.
- Clients see their own actions confirmed one round trip plus up to one tick
  late. The host sees no delay, which is acceptable in co-op.
- Delta size grows with how much changes each tick. The slice must measure it
  against a budget of 32 KB/s per client. If it exceeds that, the fix is
  quantising or batching the deltas, not changing the topology.
