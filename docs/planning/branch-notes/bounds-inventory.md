---
branch: task/explicit-upper-bounds
---

# Inventory: what can grow without a declared, asserted maximum

Read-only sweep of the whole worktree, 2026-07-28, against `main` at `6784aafe`.
Companion to [explicit-upper-bounds.md](explicit-upper-bounds.md), which states the rule
and the design questions. This file is the *findings* — no fixes proposed here.

Each site is classified:

- **STRUCTURAL** — the true maximum is derivable from static facts already in the repo
  (9 nodes, 10 edges, declared ports; strides from `Buffer/layout.go`; fd ranges). The
  bound can be computed, not guessed.
- **DYNAMIC** — the maximum depends on runtime behavior (beads in flight, firing rate,
  gesture rate). No saved file answers it; it needs a measured value plus a decision
  about what happens at the limit.
- **BOUNDED** — already has a declared constant. Separately noted whether it is
  *asserted* (fails loudly) or merely *sized*.

"Loud vs silent" is what happens today when the ceiling is reached.

## Headline: two findings that change the branch's priorities

**1. `pw.pending` can grow forever, and it is not on anyone's list.**

`writeStreamFrame` returns early when the per-edge stream is inactive
(`nodes/Wiring/edge_mover.go:236`, `if m.streamOut == nil || m.buildFrame == nil`), but
the drain — `m.dest.DrainPendingEvents()` — sits *after* that guard at `:273`. Meanwhile
`emitArrive` appends one event per arrival (`nodes/wire/paced_wire.go:559-563`) gated on
`ai.emit`, which traces back to `beadPlacement.streams()` = **`bp.Node != ""`**
(`:80-82`) — i.e. it keys off whether the *placement names a node*, **not** whether a
stream fd is wired.

So the two conditions are independent: beads keep appending while nothing drains.
`pw.pending` then grows monotonically for the life of the process, unbounded, silently —
pure memory growth with no counter and no error.

It is **latent, not active today**: with 10 edges the per-edge streams are wired, so the
drain runs. It triggers when `streamOut == nil` while beads still carry node names —
notably the `edgeCount > MAX_EDGE_STREAMS` path (see finding 2), a fallback launch, or a
long headless run. Worth confirming against a headless run before fixing, since that is
the cheapest place to observe it.

**2. The one place with a declared maximum responds to exceeding it by silently
disabling the feature.**

`MAX_EDGE_STREAMS = 256` / `MAX_NODE_STREAMS = 256`
(`tools/topology-vscode/src/runCommand.ts:35,49`) are real declared bounds — but at
`:458`/`:464` exceeding them sets the count to **0**, disabling all dedicated streams
with no message. That is the quietest possible signal for the loudest consequence, and
it is the same path that strands finding 1. Current usage is 10 edges / 9 nodes against
a ceiling of 256, so this is head-room, not a live problem.

## Channel buffers (Go)

Only `wireChanBufferSize` is a named constant. **Every other channel capacity in the
network is a bare literal.**

| Site | Connects | Cap | Class | On overflow |
|---|---|---|---|---|
| `nodes/wire/paced_wire.go:263` | source node → wire (`inCh`) | 4096 (const) | BOUNDED, sized not asserted | `BufferFull`, caller retries next cycle — **silent** |
| `nodes/wire/paced_wire.go:264` | wire → dest node (`outCh`) | 4096 (const) | BOUNDED, sized not asserted | bead stays in `inflight`, retries — **silent** |
| `nodes/wire/paced_wire.go:265` | breadcrumbs | **4** (magic) | DYNAMIC | **drops the diagnostic silently** |
| `nodes/Wiring/edge_mover.go:111-113` | `extIn`/`srcIn`/`dstIn`, per edge | **8** (magic) ×3 ×10 | DYNAMIC | retained in sender's `pending` — silent |
| `nodes/Wiring/node_mover.go:314` | per-node `extIn` | **8** (magic) ×9 | DYNAMIC | retained — silent |
| `nodes/Wiring/node_move.go:270,273` | `neighborIn[peer]` | **8** (magic) | STRUCTURAL count (≤2×edges=20), DYNAMIC depth | retained — silent |
| `nodes/Wiring/stdin_reader.go:185` | frame reader → dispatch | **8** (magic) | BOUNDED by real backpressure | blocking send, **no loss** |
| `ports.go:649`, `node_mover.go:316`, `node_move.go:210,258`, `port_wiring.go:47`, `holdflip/pulse/PulseLeft/PulseRight` | latest-wins slots | 1 | STRUCTURAL | drop-oldest **by design** |

Two observations worth more than the table:

- **The `8`s are the largest un-named group** — six-plus declarations, all magic, all
  with retain-and-retry overflow. Whatever bound they should have, they should share a
  named constant.
- **`breadcrumbCh` at 4 is the tightest queue in the system and drops silently.** The
  diagnostic path is the least reliable path in the repo. Given the session that just
  ended — where a breadcrumb channel silently going quiet was the whole near-miss — this
  deserves its own decision rather than a blanket policy.

## Accumulators on a per-tick / per-event path

| Site | Drained by | Class | Notes |
|---|---|---|---|
| `paced_wire.go:165` `pw.pending` | `drainPendingEvents` — **only via `writeStreamFrame`** | DYNAMIC | **See headline 1.** Unbounded when `streamOut == nil` |
| `paced_wire.go:143` `pw.inflight` | FIFO-head delivery only (`:443`) | DYNAMIC | Head-of-line: if the head cannot hand off, nothing behind it drains either. Backing array only ever grows (`inflight[1:]` never re-slices) |
| `node_mover.go:175` `nm.pending` | `flushPending` (`:823-841`) | DYNAMIC | A blocked peer inbox (cap 8) retains that item *and every later item to it*, FIFO-preserved |
| `layout_holder.go:156` `lh.localPolars` | upsert by `To` | STRUCTURAL | ≤ distinct neighbours ≤ 2×edges = 20. Not a leak |
| `bead_emit.go:46-53`, `node_mover.go:791`, `paced_wire.go:491` | rebuilt per frame | STRUCTURAL | locals; size tracks the dynamic sources above |
| load-time appends (`build.go`, `loader_*.go`, `topo_spec.go`, `validate.go`, `row_tables.go`) | n/a | STRUCTURAL | one pass over 9 nodes / 10 edges |

## Loops

Main `select` loops (one `SleepCycle` per iteration, ctx-cancellable) are fine and not
listed — that is the correct shape for a node goroutine.

| Site | Shape | Class | On overrun |
|---|---|---|---|
| `distance_groups.go:150-155` `waitForCenterSettle` | `for time.Now().Before(deadline)` + 1 ms sleep, 200 ms cap | DYNAMIC | **Bounded by time, not iterations, and exits silently on timeout with no signal it never settled.** The only `time.Sleep` on a non-clock path |
| `paced_wire.go:383` `drainPlacements` | drain-until-empty | DYNAMIC | can pull up to 4096 in one cycle — silent stall/jitter |
| `paced_wire.go:185`, `gate.go:96`, `TimeStart:100`, `Time:98`, `edge_mover.go:333`, `node_mover.go:883` | drain-until-empty | DYNAMIC | bounded only by the producer — silent |
| `paced_wire.go:414` `stepAll` | `i` not advanced on delivery | correct | length is the unbounded `inflight` |

**The model to copy already exists in-tree:** `runCommand.ts:271`
`BUILD_BINARY_MAX_ATTEMPTS = 50`, declared *and* loud — `:286` returns
`binary … not built after 50 attempts`. This is the only site already matching the
branch's target shape, and the bound reads as a sentence when it fires.

## Sinks and frames

| Site | Class | Notes |
|---|---|---|
| `.probe` appends — `runCommand.ts:631,665,698,729` | DYNAMIC | **No byte cap, no row cap.** Truncated only at run start. The 1.1 GB case. Now gated by `wirefold.probe.trace` (default off), but the gate is a switch, not a bound |
| `runCommand.ts:116`, `extension.ts:140` `go-errors.jsonl` | DYNAMIC | appended per stderr line, unbounded — and it is always-on by design |
| `Buffer/stream_events.go:60-80` `BuildEventsSection` | DYNAMIC | frame size is a direct function of drained-pending count; no cap on `len(events)` |
| `stdin_reader.go:162` `maxFrameBytes = 1<<20` | **BOUNDED and asserted** | rejects and stops the reader — the only asserted limit in Go besides nothing else |
| `Trace/Trace.go:268,284` | bounded by construction | direct write, no accumulation; nil sink in production |

**Protocol asymmetry worth its own line:** Go enforces `maxFrameBytes` on inbound frames;
the TS counterpart `splitFrames` (`runCommand.ts:214-228`) reads `frameLen` off the wire
with **no maximum**, so a corrupt length grows `rest` until the payload "arrives". The
same protocol is bounded on one side only.

## TS side

| Site | Class | Notes |
|---|---|---|
| `runCommand.ts:365,371,381,382` last-frame caches | STRUCTURAL cardinality | `Map.set` by row, ≤ edge/node count, cleared on respawn (`:477-480`). **Cardinality bounded, per-entry bytes not** — each holds a whole frame |
| `webview/snapshot-buffer.ts:63,102,132` | STRUCTURAL cardinality | same shape but **never cleared** — stale rows persist across a Go restart |
| `webview/three/buffer-decode.ts:178-179,336-337,402-403` | STRUCTURAL cardinality | six module-level Maps, `.set()` only, **no `delete`/`clear` in the file** — renumbered rows leak for the session |
| `runCommand.ts:796,811 pendingStdin` | DYNAMIC | **no cap**; if the process never spawns (build failure returns rather than spawning) every later gesture appends forever — silent |
| `runCommand.ts:194-202 splitJsonlLines` | DYNAMIC | `rest` grows until a newline arrives |

## What this means for the plan

The branch note asks for "a declared maximum on every queue, buffer, and loop." The sweep
says that is the right goal but the wrong *first* move, for three reasons:

1. **Most sites are STRUCTURAL**, and their bound should be *derived* from topology and
   layout (≤ edges, ≤ nodes, ≤ ports), not chosen. Deriving is mechanical and low-risk;
   picking numbers is neither. These can land as a batch.
2. **A handful are DYNAMIC** — `inflight`, `pending`, the mover inboxes, the drain loops.
   Each needs a measured value *and* an at-the-bound decision (panic / drop-oldest /
   drop-newest / backpressure), and for a paced-clock simulation those are not
   interchangeable: dropping changes what renders, blocking changes timing. These are one
   decision each, not a policy.
3. **Two sites are bugs, not missing bounds** — headline 1 (`pending` never drained) and
   the `splitFrames` asymmetry. A bound would paper over the first rather than fix it.
   The branch note already warns against exactly this ("don't paper over it with a
   bound").

Suggested order: fix the two bugs, derive the structural bounds as a batch, then take the
dynamic ones one at a time with their at-the-bound decision written down.

Per `memory/feedback_check_the_signal_the_check_emits.md`, every bound that lands must be
exceeded once deliberately to confirm it fires and names itself. A limit nobody has seen
fail is a comment.

**Ownership note:** every accumulator above is owned by exactly one goroutine
(`pending`/`inflight` = the wire's; `nm.pending` = that node's; TS maps = ext-host or
webview thread), so a per-site counter and assertion needs no shared state — consistent
with MODEL.md and with the testing-shape rule that a test asserts what one goroutine
itself decided.
