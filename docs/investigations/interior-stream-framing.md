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

The model requires one writer per fd (`memory/feedback/architecture/bridge/feedback_no_single_writer_bridge.md`,
`Buffer/streamframe/stream_fds.go`'s `StreamKindInterior` doc comment: "written by that node's OWN
Update goroutine"). That invariant WAS **violated by construction** for every node kind that
used `gatecommon.DriveHeld` — `Pulse`, `PulseLeft`, `PulseRight`, `holdflip`, `Time`,
`TimeStart`.

`DriveHeld` spawned its **own goroutine** (`go func() { ... }()`) that called
`Out.PlaceDrivenAt` → `flushSendEvent` → `s.WriteEvents(...)` on the
node's shared `*interiorStream` (`nodes/wire/outport/out_port_send.go`'s `flushSendEvent`). Meanwhile the node's own
`Update` goroutine (its main loop, e.g. `nodes/pulse/node.go`'s `consume()`) calls
`EmitHeldBead(v)` → the SAME shared `*interiorStream` (`nodes/Wiring/portwiring/port_wiring.go`'s
`NewInteriorStreamGetter` lazily builds **one** `*interiorStream` per node and hands the
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

`writeInteriorStreamFrame` (`nodes/Wiring/interior/interior_stream.go:95-106`) writes a frame in
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

`BuildEventsSection` (`Buffer/streamframe/stream_events.go`) allocates
`4 + len(events)*BufEventStride + textLen` bytes and returns exactly that many — the
declared `[count:u32]` plus per-row `TextOff/TextLen` columns are computed from the SAME
`textBytes` slice that gets appended, so the returned buffer's actual length always equals
what the header will declare (`writeInteriorStreamFrame` writes `len(frame)`, where `frame`
is that exact returned/appended byte slice — `Buffer/streamframe/node_stream_frame.go:140-163`). The
prior note that the *headless test's* length formula ignores event Text bytes describes a
gap in that ONE test's arithmetic, not in the production writer — the writer never computes
a length formula at all, it measures the real slice. `main.go`'s `toStreamEvents` also
carries `Text` through unchanged from `wire.RowEvent` to `Buffer.StreamEvent`. Ruled out.

## 4. Live-session-only conditions

Given #1/#2 fully explain the symptom by construction, #4 (drag-rate raw-input, reload,
hot-restart, backpressure) was not pursued as an independent cause — but it is exactly the
kind of condition (concurrent activity on a `DriveHeld` node under load) that would make
this race land *more often* in a live session than in a short/synthetic headless run: a
topology with an actively-firing `Pulse`/`Time`/`HoldFlip`-family node (recv events racing
its own drive goroutine's send events) needs to run for long enough to hit this timing
window — which is exactly why a short headless run can pass while the live editor didn't.

## Reproduction (original, pre-fix)

Constructing ONE shared `*interiorStream` over a real `os.Pipe`, and spawning two
goroutines against it (one calling `write` the way a node's Update loop does via
`EmitHeldBead`, one calling `WriteEvents` the way a `DriveHeld` drive goroutine's
`flushSendEvent` does) for 20000 iterations each, with a reader goroutine applying the SAME
`[len:u32]` framing + `MAX_FRAME_BYTES` bound `runCommand.ts`'s `splitFrames`/
`handleInteriorFd` use, plus a stricter "is this one of the two legitimate frame shapes"
check (since a corrupted-but-small length can slip under `MAX_FRAME_BYTES` and silently
misdecode instead of erroring loudly — a quieter version of the same bug), failed reliably
(every run observed, in well under 1s) with a decoded frame of a length that was neither of
the two legitimate shapes — the corrupted-length symptom, reproduced from the writer side
alone, no editor or extension host involved. This was the reproduction this doc's original
version was the deliverable of — the fix landed since (see "Fix" below).

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

The second goroutine no longer exists. `DriveHeld` was replaced by
`helddrive.HeldDriver` (`nodes/Wiring/helddrive/held_driver.go`), which is a plain
stepper the node's OWN loop calls once per tick — the ownership fix the Guard verdict above
names, arrived at from the other direction (every node kind became one goroutine). Two
goroutines can no longer reach one node's streams, so the race in §2 is now unrepresentable
rather than avoided. The per-fd `drive` stream kind below has since been REMOVED with it:
it existed only to keep the two writers on separate fds, and there is no second writer to
separate. A driven out's events now ride the node's own interior stream.

See [interior-stream-framing-fix.md](interior-stream-framing-fix.md) for the earlier fix: the
per-fd `drive` stream kind, the Go-side wiring, how the fd count is known before spawn,
mismatch handling, the read side, the single-Write framing done alongside, and verification
(including the compile-time fix that replaced `check-driveheld-uses-driveout.sh`).
