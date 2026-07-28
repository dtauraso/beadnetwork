---
branch: task/edge-stream-volume
---

# Why does the edge stream emit 734 KB/s?

## Measured

Taken live on 2026-07-28, one editor session running the shipped 9-node topology,
no interaction — idle playback only:

| Log | Size after one long session |
|---|---|
| `.probe/go-edge.jsonl` | **1.1 GB** |
| `.probe/go-interior.jsonl` | 34 MB |
| `.probe/ts.jsonl` | 1.8 MB |
| `.probe/go-node.jsonl` | 900 KB |
| `.probe/go.jsonl` (view) | 14 KB |

Rate, measured over 10s with `stat -f%z`: **580914 -> 1332944 bytes = 734 KB/s,
~258 MB/hour.** The edge stream is ~30x everything else combined.

Every line is the same shape — a per-bead position sample:

```
{"ts_ms":1785224605871,"src":"go","kind":"edge-bead","node":"","port":"","value":0,
 "x":979.02,"y":-369.65,"z":-505.05,"f":0.887...}
```

## The question

Is this rate INHERENT (N beads x M frames/sec is genuinely what an edge stream carries,
and the log is simply recording every frame), or is something emitting per-frame where
it should emit per-change?

Worth checking, in order:

1. Who emits `kind:"edge-bead"` — the Go edge mover's own trace call, and at what
   cadence relative to its clock cycle.
2. Whether the ext host writes EVERY decoded event to `.probe` or is supposed to
   sample/filter. `runCommand.ts`'s `appendFileSync` sites are the write path.
3. Whether `node`/`port`/`value` being empty/zero on every line means the row is
   carrying fields it does not need (cheap win on volume, not on rate).
4. Whether the VIEW stream's 14 KB vs the edge stream's 1.1 GB reflects a real
   difference in what changes, or a missing "only emit when something changed" rule that
   the view path has and the edge path does not. CLAUDE.md says Go "only ever emits a
   frame when something changes, and that stays true" — a bead in flight DOES change
   every frame, so this may be legitimate; confirm rather than assume.

## Constraint

CLAUDE.md's breadcrumb guidance already names the log-flood lesson: keep the debug
channel SPARSE, "a debug tool for control events, not a per-tick firehose". The edge
stream is not a breadcrumb channel, but the `.probe` decode of it is subject to the same
rule.

## Answers (2026-07-28)

All four questions resolved by reading the code. **The rate is log-only overhead, not
render traffic.** The renderer never sees a single one of these events.

**1. Who emits, at what cadence.** `advanceBead` sets `emit = true` for every in-flight
bead on every tick whenever the bead is streaming and its arc is nonzero
(`nodes/wire/paced_wire.go:559-593`); `stepAll` appends each as a `KindEdgeBead`
pendingWireEvent (`:405-410`). `edgeMover.run` then calls `writeStreamFrame`
**unconditionally every cycle** (`nodes/Wiring/edge_mover.go:365-368` — "Beads on this
wire may have moved even with no geometry change this cycle"), with no dirty check;
`writeStreamFrame` early-returns only on a nil `streamOut`/`buildFrame`
(`edge_mover.go:235-237`), never on content equality. `Trace/Trace.go:38-41` already
documents the kind as resolving "one every ~16 ms while a bead is in flight".

**2. Nothing filters the write.** Every decoded event is appended:
`handleEdgeFd` decodes the frame's EVENTS section and `appendFileSync`s the whole batch
to `.probe/go-edge.jsonl` (`tools/topology-vscode/src/runCommand.ts:630-646`). No
sampling, no level gate. `runCommand.ts:139-142` already carries a comment measuring the
result at "1.2 GB, 1.1 GB of it go-edge.jsonl" — the cost was known and mitigated only by
rotation.

**3. The event is ~80% redundant with the frame's own Bead block.** Same frame, two
copies of the same bead:

| | fields | bytes/bead/tick |
|---|---|---|
| Bead block (`Buffer/buffer_layout_gen.go:35-37`) | `x, y, z, value` | 16 (`BufBeadStride`) |
| `edge-bead` event (`Buffer/stream_events.go:18`) | `x, y, z, value` **+ `gen`, `t`** | 67 (`BufEventStride`) |

So `x/y/z/value` are duplicated 1:1, and the event's only unique content is `gen` (bead
identity) and `t` (fractional progress) — for ~4x the bytes. Note this answers the
original question 3 differently than it guessed: the empty `node`/`port` columns are a
rounding error next to the whole-row duplication.

**4. The 14 KB vs 1.1 GB gap is real and the edge stream is the outlier.** Every other
stream is change-driven: `KindNodeBead` emits only when the array changes
(`Trace/Trace.go:60-68`, `nodes/Wiring/bead_emit.go:16-17`), `KindNodeGeometry` on load
and on node-move (`Trace/Trace.go:49-52`). `KindEdgeBead` is the only continuous
per-tick per-item sample in the system.

**The decisive fact — no production consumer.** `decodeEventLine`'s `"edge-bead"` case
(`buffer-log.ts:241-244`) builds a `Line` object whose only path is the `.probe` append
above. Grepping the entire `webview/` tree for `"edge-bead"` returns **zero hits**; the
on-wire bead renderer reads the Bead block instead, via `edgeStream.beads(row)`
(`webview/three/BeadInstances.tsx:42`). The event is decoded, written to a log file, and
discarded.

So CLAUDE.md's "Go only ever emits a frame when something changes" is NOT violated — a
bead in flight really does change every frame, and the *frame* is legitimate. What is
unjustified is the second, larger, per-bead copy riding along in the EVENTS section that
only a log file reads.

## Proposed next step (one step, needs sign-off — this is Go network code)

Stop appending the `KindEdgeBead` pendingWireEvent in `stepAll`, keeping the
side-effect-free `LiveBeadRow`/Bead-block path that actually feeds the renderer. That
removes the ~734 KB/s at the source rather than rotating it away.

Blast radius, checked:

- Renderer: unaffected (no consumer — see above).
- `Trace/Trace.go:165` `TraceEventKinds` and `src/schema/trace-kinds.ts:4` both list the
  kind; they are parity-checked against each other, so both change together.
- `test/contracts/trace-event-fields.test.ts:33` asserts the exact kind SET and `:76-86`
  asserts the edge-bead row shape. Both fail on removal and must be updated in the same
  commit — that is the guard doing its job, not an obstacle.
- Cost: per-bead `gen` and fractional `t` stop being observable during transit. Nothing
  reads them today. If a future debugging need wants them, they belong as columns on the
  Bead block (widening `BufBeadStride`), not as a duplicate event row.

The alternative — keep emitting and let rotation bound it — leaves a per-tick firehose
running permanently to feed a log nobody reads, which is the log-flood lesson CLAUDE.md
already names.

## Related

`task/probe-log-rotation` (merged separately) bounds the SYMPTOM — logs now rotate per
run and VS Code no longer watches them. It deliberately does not touch the emission
rate. Rotation turns out NOT to be the whole fix: the rate is not legitimate, because
the bytes feed no consumer.
