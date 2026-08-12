# Interior-stream framing corruption — the fix

[← interior-stream-framing.md](interior-stream-framing.md)

## Fix

**Chosen: each emitting goroutine gets its own fd** — the stated bridge invariant
(CLAUDE.md: "one dedicated inherited-stdio pipe per emitting goroutine",
`memory/feedback/architecture/bridge/feedback_no_single_writer_bridge.md`), not a lock (`tools/network/concurrency/check-no-network-locks.sh`
has an empty allowlist — a mutex was never on the table) and not channel-routing back
through the node's own Update goroutine (that would still be a real option under the
model, but the per-fd shape was chosen as the more direct fix and matches how `interior`
itself was already split out from the old fd-3 accumulator).

**New stream kind: `drive`** (`Buffer.StreamKindDrive`, `Buffer/streamframe/stream_fds.go`). One
dedicated fd per **(node row, drive slot)** — `fd = baseFd["drive"] + nodeRow*
DriveSlotsPerNode + slot`, `DriveSlotsPerNode = 2` (the current max: `Pulse` drives both
`Out` and `OutFanout`; every other `DriveHeld`-driving kind — `PulseLeft`, `PulseRight`,
`holdflip` — drives one). Allocated unconditionally per node row, kind-agnostic, mirroring
how `node`/`interior` are already allocated regardless of a row's actual kind — a slot a
kind doesn't use is simply never written (nil-safe, same fallback as every other unused
stream). `Time`/`TimeStart` were investigated and found NOT to need this: they call
`wire.Broadcast.PlaceDrivenAllAt` synchronously from their own Update loop, never spawning
a second goroutine — the original per-kind list in this doc's earlier sections was
imprecise on that point; only `Pulse`/`PulseLeft`/`PulseRight`/`holdflip` actually call
`gatecommon.DriveHeld`.

**Go-side wiring**: `nodes/Wiring/streamwire/stream_wiring.go`'s `StreamWiring.driveOuts` holds
`[DriveSlotsPerNode]io.Writer` per node id, populated by `SetNodeStreams` alongside
`interiorOuts` (same "wire before any goroutine launches" ordering).
`nodes/Wiring/portwiring/port_wiring.go`'s `NewDriveStreamGetter(name, slot, pb)` is
`NewInteriorStreamGetter`'s counterpart, reading `driveOuts[name][slot]` instead of
`interiorOuts[name]` — a DIFFERENT lazy-cache-once closure, so it can never alias the
node's own interior stream by construction. `nodes/Wiring/kindapi/build_args.go`'s
`BuildArgs.DriveOut(portName, slot)` is the kind-facing entry point: `Pulse`/`PulseLeft`/
`PulseRight`/`holdflip` now build their `DriveHeld`-driven `Out`/`OutFanout` fields via
`a.DriveOut(...)` instead of `a.Out(...)`.

**fd count known before spawn — how**: `runCommand.ts` derives the "drive" range's size
from `nodeCount * DRIVE_SLOTS_PER_NODE`, where `nodeCount` is the SAME `counts.json`-stored
number (`.claude/rules/persistence-ownership.md` "Counts are stored, never re-derived")
already used to size the `node`/`interior` ranges — no new counting mechanism, no
per-kind lookup at spawn time (the ext host does not need to know which node kinds use
`DriveHeld`; every row gets `DriveSlotsPerNode` fds regardless, same as `node`/`interior`).
`DRIVE_SLOTS_PER_NODE` is a hand-kept mirrored constant on both sides (matching
`MAX_EDGE_STREAMS`/`MAX_NODE_STREAMS`'s existing "small bound, kept in parity by hand"
shape) — raise both together if a kind ever needs a third `DriveHeld` output.

**Mismatch handling — loud, not crash-loop**: `main.go` now requires `"node"`, `"interior"`,
AND `"drive"` present together in `WIREFOLD_STREAM_FDS` (extending the prior `node`/
`interior` pairing check to three). A partial set is reported via `fmt.Fprintf(os.Stderr,
...)` — which the ext host already captures into `.probe/go-errors.jsonl`
(`appendGoError`) and its output channel — and leaves ALL THREE per-node streams unwired
for that run (same degrade-and-report shape the existing `node`/`interior`/`edge`
mismatches already use), never a panic. The editor's runner respawns Go on exit, so a
startup panic would become an unreadable flicker; this instead runs degraded (no node
geometry/interior beads/drive-goroutine sends drawn) with the cause visible in the error
log, exactly like every other stream-fd mismatch in this file already does.

**Read side**: `runCommand.ts`'s `handleDriveFd(row, slot, chunk)` keeps its OWN
partial-frame carry buffer and dead-stream key per `(row, slot)` — reusing `interiorBufs`'
carry state across two physically distinct pipes would reintroduce the exact desync this
whole change removes, just moved from Go's write side to the read side. Once reassembled,
a drive-slot frame IS an interior-shaped frame for that node row
(`Buffer.BuildInteriorStreamFrame` on both), so DECODE and PROBE-LOG
(`processInteriorLikeFrames`) is shared with `handleInteriorFd` — but the shared tail takes
an `assertsSlots` flag (`handleInteriorFd` passes `true`, `handleDriveFd` passes `false`)
and a drive frame with `assertsSlots=false` is never written into `lastInteriorFrames` and
never relayed to the webview as interior state (`BUF_BLOCK_TAG_INTERIOR_STREAM`). This was
tightened after a follow-on bug (see below): a drive stream's own `lastPresent` is never set
by anything, so relaying its frame as though it stated the node's slots painted an
all-absent snapshot over a held bead the node's own interior stream had just emitted. Only
the node's own interior stream — its Update loop's `emitHeldBead`/`emitNodeBeads`/
`emitInputBeads` — is the one writer of slot state; a drive frame's EVENTS are still decoded
and probe-logged, they just never reach the webview's interior cell or its replay cache.

**Single-Write framing, done alongside**: `writeInteriorStreamFrame`
(`nodes/Wiring/interior/interior_stream.go`) now issues ONE `io.Writer.Write` call per frame
(header+payload in one buffer) instead of two — closes the short-write/signal hazard for a
genuinely single writer, on top of (not instead of) the per-fd fix. It does NOT make
sharing a stream between two goroutines safe on its own (see the reproduction test's
current form, below) — the per-fd split is the structural fix; the single-Write is cheap
insurance layered on it, at the cost of one extra byte-slice allocation per frame (buildFrame's
own backing array can't be mutated in place, since callers may still hold it).

**Verification**:

- **`tools/check-driveheld-uses-driveout.sh` REMOVED (2026-08), superseded by a compile-time
  fix.** It used to express the SAME invariant at the source-text level — every
  `nodes/<Kind>/node.go` that calls `gatecommon.DriveHeld` must also call `a.DriveOut(...)`,
  and must not resolve `"Out"`/`"OutFanout"` via bare `a.Out(...)` — by grepping for the
  mistake after the fact. `nodes/Wiring/kindapi/driven_out.go`'s `Wiring.DrivenOut` now makes that
  mistake **unrepresentable**: `gatecommon.DriveHeld`'s signature accepts ONLY
  `Wiring.DrivenOut`, never a bare `*wire.Out`, and the only way to construct one is the
  unexported `newDrivenOut` constructor, called from exactly one place
  (`BuildArgs.DriveOut`). `Wiring.DrivenOut`'s only field is unexported, so no package
  outside `nodes/Wiring` can construct one via a struct literal — patching a kind back to
  `n.Out = a.Out("Out")` (with `Out` typed `Wiring.DrivenOut`) is now a `go build` failure,
  confirmed by reverting `nodes/holdflip/node.go` to that exact patch. The shell guard's
  grep-the-source-text check is strictly weaker than a compile error the mistake can no
  longer even type-check its way past, so it was deleted rather than kept alongside the
  structural fix.
- `TestHeadlessInteriorFdSustainedFraming` (the existing sustained-framing test) now passes
  cleanly against the real spawned binary — it previously would have hit this bug under
  enough load; the interior fd it reads now carries ONLY the node's own Update-loop
  writes.
