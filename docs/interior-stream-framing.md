# Interior-stream framing corruption — investigation and reproduction

## Symptom

Recurring, in live editor sessions only:

```
handleInteriorFd(row=2): bad frame length 16777216 (max 1048576); stopping stream
handleInteriorFd(row=4): bad frame length 3758096384 (max 1048576); stopping stream
handleInteriorFd(row=2): bad frame length 4235264  (max 1048576); stopping stream
```

(`.probe/go-errors.jsonl`, written by `tools/topology-vscode/src/runCommand.ts`), paired
with a repeating `RangeError: Offset is outside the bounds of the DataView` from
`getFloat32` in the webview. `16777216 = 0x01000000`, `3758096384 = 0xE0000000` — both look
like a `[len:u32]` field read from the middle of payload bytes rather than a real frame
boundary: the reader has lost frame sync on that node row's dedicated interior fd.

## What was already known (not redone here)

- `check-buffer-layout-parity.sh` / `check-frame-bytes-parity.sh` pass — the generated
  layout is not stale.
- `TestHeadlessInteriorFdSustainedFraming` decodes every frame off every node's interior
  fd over a 3s headless run against the real binary and passes.
- A standalone Node harness using production's exact stdio array + `splitFrames` ran 15s
  clean.

None of that reproduces from a clean start under steady state. This investigation's job was
to find what a *live* editor session does differently.

## 1. Who writes the interior fd — REPRODUCED VIOLATION

The model requires one writer per fd (`memory/feedback_no_single_writer_bridge.md`,
`Buffer/stream_fds.go`'s `StreamKindInterior` doc comment: "written by that node's OWN
Update goroutine"). That invariant is **violated by construction** for every node kind that
uses `gatecommon.DriveHeld` — `Pulse`, `PulseLeft`, `PulseRight`, `holdflip`, `Time`,
`TimeStart` (`nodes/gatecommon/drive.go`).

`DriveHeld` spawns its **own goroutine** (`nodes/gatecommon/drive.go:86`, `go func() {
... }()`) that calls `Out.PlaceDrivenAt` → `flushSendEvent` → `s.WriteEvents(...)` on the
node's shared `*interiorStream` (`nodes/wire/ports.go:516-531`). Meanwhile the node's own
`Update` goroutine (its main loop, e.g. `nodes/pulse/node.go`'s `consume()`) calls
`EmitHeldBead(v)` → the SAME shared `*interiorStream` (`nodes/Wiring/port_wiring.go`'s
`newInteriorStreamGetter` lazily builds **one** `*interiorStream` per node and hands the
same instance to every closure/port belonging to that node, by design — see its doc
comment, "every closure/port... shares the SAME instance").

So a `Pulse` node (for example) has **two independently-scheduled goroutines** — its
`Update` loop and its `DriveHeld` drive goroutine (a third with `OutFanout` wired) — both
calling into the same `*interiorStream`, hence both writing to the same OS pipe fd. The
code comments on `flushSendEvent`/`PollRecv`/`WriteEvents` assert "this node's own Update
goroutine (the SAME goroutine driving the send) is the sole owner" — that assertion is
**false** for any `DriveHeld`-backed kind: the drive goroutine is not the Update goroutine,
it is a second one `DriveHeld` itself spawns.

## 2. Write atomicity — the mechanism that turns the violation into corruption

`writeInteriorStreamFrame` (`nodes/Wiring/interior_stream.go:95-106`) writes a frame in
**two separate `io.Writer.Write` calls**: the 4-byte `[len:u32]` header, then the frame
payload. There is no lock held across the two. Each individual `Write` call is atomic (Go's
`os.File.Write` is safe for concurrent callers and won't itself split a call's bytes), but
nothing prevents another goroutine's `Write` call from landing **between** this goroutine's
header write and its payload write. With two goroutines both calling
`writeInteriorStreamFrame` on the same `out`, the four writes can land in the order
`A-header, B-header, B-payload, A-payload` (or any other interleaving) — the reader then
reads `A-header`'s declared length, but the bytes that follow are `B-header ++ B-payload`
(or a `B-payload` fragment), which decode as garbage exactly like the reported values.

This is the sufficient cause: no fix to frame *content* (layout, endianness, event
encoding) touches it, because the corruption happens **between** two well-formed frames,
not inside one.

## 3. The events section — checked, not the cause

`BuildEventsSection` (`Buffer/stream_events.go`) allocates
`4 + len(events)*BufEventStride + textLen` bytes and returns exactly that many — the
declared `[count:u32]` plus per-row `TextOff/TextLen` columns are computed from the SAME
`textBytes` slice that gets appended, so the returned buffer's actual length always equals
what the header will declare (`writeInteriorStreamFrame` writes `len(frame)`, where `frame`
is that exact returned/appended byte slice — `Buffer/node_stream_frame.go:140-163`). The
prior note that the *headless test's* length formula ignores event Text bytes describes a
gap in that ONE test's arithmetic, not in the production writer — the writer never computes
a length formula at all, it measures the real slice. `main.go`'s `toStreamEvents` also
carries `Text` through unchanged from `wire.RowEvent` to `Buffer.StreamEvent`. Ruled out.

## 4. Live-session-only conditions

Given #1/#2 fully explain the symptom by construction, #4 (drag-rate raw-input, reload,
hot-restart, backpressure) was not pursued as an independent cause — but it is exactly the
kind of condition (concurrent activity on a `DriveHeld` node under load) that would make
this race land *more often* in a live session than in a short/synthetic headless run:
`TestHeadlessInteriorFdSustainedFraming` and the Node harness were not confirmed to include
a topology with an actively-firing `Pulse`/`Time`/`HoldFlip`-family node (recv events
racing its own drive goroutine's send events) for long enough to hit this timing window —
which is exactly why they passed while the live editor didn't.

## Reproduction (original, pre-fix)

`nodes/Wiring/interior_stream_concurrent_write_test.go`'s original
`TestInteriorStreamConcurrentWritersDesyncFraming` (since renamed and reshaped — see
"Fix" below for its current form): constructed ONE shared `*interiorStream` over a real
`os.Pipe`, spawned two goroutines against it (one calling `write` the way a node's Update
loop does via `EmitHeldBead`, one calling `WriteEvents` the way a `DriveHeld` drive
goroutine's `flushSendEvent` does) for 20000 iterations each, and a reader goroutine
applying the SAME `[len:u32]` framing + `MAX_FRAME_BYTES` bound `runCommand.ts`'s
`splitFrames`/`handleInteriorFd` use, plus a stricter "is this one of the two legitimate
frame shapes" check (since a corrupted-but-small length can slip under `MAX_FRAME_BYTES`
and silently misdecode instead of erroring loudly — a quieter version of the same bug). It
failed reliably (every run observed, in well under 1s) with a decoded frame of a length
that was neither of the two legitimate shapes — the corrupted-length symptom, reproduced
from the writer side alone, no editor or extension host involved. This was the reproduction
this doc's original version was the deliverable of — the fix landed since (see "Fix" below,
which also covers what the test looks like now that production no longer constructs this
sharing at all).

## Guard verdict

No existing guard would have caught this. `check-no-network-locks.sh` guards AGAINST
locks/atomics (the model's actual fix here — see below — legitimately needs neither if
done as ownership: `DriveHeld` should route its send-event notification back through the
node's own Update-loop goroutine instead of writing the shared stream directly, which is
the "no shared memory, message-passing only" shape this codebase otherwise uses
everywhere). There is no guard that asserts "every `*interiorStream` is written by exactly
one goroutine" — the invariant is stated only in doc comments (`interior_stream.go`,
`ports.go`, `stream_fds.go`), never checked. A `go vet`/`-race` run does NOT catch this
either: each individual `Write()` call is internally synchronized by `os.File`, so there is
no *data race* in the Go race-detector's sense, only a *logical* framing race across two
separate `Write()` calls — worth noting for whoever writes the guard, since "run with
`-race`" is not sufficient.

## Fix

**Chosen: each emitting goroutine gets its own fd** — the stated bridge invariant
(CLAUDE.md: "one dedicated inherited-stdio pipe per emitting goroutine",
`memory/feedback_no_single_writer_bridge.md`), not a lock (`tools/check-no-network-locks.sh`
has an empty allowlist — a mutex was never on the table) and not channel-routing back
through the node's own Update goroutine (that would still be a real option under the
model, but the per-fd shape was chosen as the more direct fix and matches how `interior`
itself was already split out from the old fd-3 accumulator).

**New stream kind: `drive`** (`Buffer.StreamKindDrive`, `Buffer/stream_fds.go`). One
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

**Go-side wiring**: `nodes/Wiring/stream_wiring.go`'s `streamWiring.driveOuts` holds
`[DriveSlotsPerNode]io.Writer` per node id, populated by `setNodeStreams` alongside
`interiorOuts` (same "wire before any goroutine launches" ordering).
`nodes/Wiring/port_wiring.go`'s `newDriveStreamGetter(name, slot, pb)` is
`newInteriorStreamGetter`'s counterpart, reading `driveOuts[name][slot]` instead of
`interiorOuts[name]` — a DIFFERENT lazy-cache-once closure, so it can never alias the
node's own interior stream by construction. `nodes/Wiring/build_args.go`'s
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
though, a drive-slot frame IS an interior-shaped frame for that node row
(`Buffer.BuildInteriorStreamFrame` on both), so decode/probe-log/cache/relay
(`processInteriorLikeFrames`) is shared with `handleInteriorFd` — same `lastInteriorFrames`
cache (last writer wins), same `BUF_BLOCK_TAG_INTERIOR_STREAM` tag to the webview.

**Single-Write framing, done alongside**: `writeInteriorStreamFrame`
(`nodes/Wiring/interior_stream.go`) now issues ONE `io.Writer.Write` call per frame
(header+payload in one buffer) instead of two — closes the short-write/signal hazard for a
genuinely single writer, on top of (not instead of) the per-fd fix. It does NOT make
sharing a stream between two goroutines safe on its own (see the reproduction test's
current form, below) — the per-fd split is the structural fix; the single-Write is cheap
insurance layered on it, at the cost of one extra byte-slice allocation per frame (buildFrame's
own backing array can't be mutated in place, since callers may still hold it).

**Tests**:

- `nodes/Wiring/interior_stream_concurrent_write_test.go`'s
  `TestInteriorStreamTwoCallSharedWriterMechanismStillDesyncs` pins the ORIGINAL
  mechanism by hand — two goroutines, one pipe, each framing via a frozen copy of the
  PRE-FIX two-Write-calls shape (`writeTwoCallFrame`). It PASSES when the desync it exists
  to pin actually occurs and FAILS when it does not (inverted from the original
  fail-on-desync form, because a permanent `go test ./...` member cannot be "expected to
  fail" — see the test's own doc comment). The single-Write fix alone does not make sharing
  safe: an earlier attempt to prove that by inflating frame size past a guessed OS
  pipe-buffer threshold turned out to pass unreliably (this OS's pipe write behavior
  serialized even the large single-Write frames in practice) — the two-Write-calls shape is
  the deterministic reproduction that doesn't depend on guessing pipe internals.
- `nodes/Wiring/drive_stream_wiring_test.go`'s `TestDriveStreamNeverSharesNodesInteriorStream`
  is the WIRING assertion the original bug actually needed: that
  `newInteriorStreamGetter` and `newDriveStreamGetter` resolve to DIFFERENT
  `*interiorStream` instances (and different underlying `io.Writer`s) for the same node.
  Confirmed to fail against a deliberately-reintroduced pre-fix arrangement (patched
  `newDriveStreamGetter` to alias `newInteriorStreamGetter`) and pass against the real fix.
- `tools/check-driveheld-uses-driveout.sh` is a guard expressing the SAME invariant at the
  source-text level: every `nodes/<Kind>/node.go` that calls `gatecommon.DriveHeld` must
  also call `a.DriveOut(...)`, and must not resolve `"Out"`/`"OutFanout"` via bare
  `a.Out(...)`. Confirmed to fail when `nodes/holdflip/node.go` is patched back to `a.Out("Out")`
  and pass on the real tree.
- `TestHeadlessInteriorFdSustainedFraming` (the existing sustained-framing test) now passes
  cleanly against the real spawned binary — it previously would have hit this bug under
  enough load; the interior fd it reads now carries ONLY the node's own Update-loop
  writes.
