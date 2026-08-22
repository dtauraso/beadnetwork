---
branch: task/buffer-goes-away
---

# The buffer goes away

## Target

The Go → TS seam is **files only**. Go writes binary files; TS reads them. No inherited-stdio
frame streams, no `WIREFOLD_STREAM_FDS`, no framing/demux/last-frame cache, and no
`src/Buffer`.

## Why now

The trace events were the buffer's last payload. With them written per-item as binary files,
a stream frame carries a 4-byte tick and a layout fingerprint, and neither is read:

- Nothing anywhere reads the decoded `tick`.
- The fingerprint refuses a frame whose columns would be decoded at wrong offsets. The layout
  declares **no columns** — there is nothing left to mis-decode.
- `decodeNodeStreamFrame` and `decodeInteriorStreamFrame` are dead: their importers use
  `nodeLabel` (a file reader) and the `INTERIOR_SLOTS_PER_NODE` constant, never the decoders.

The one surviving consumer is `NavGuides.tsx:41`, `!viewFrameReady() || ownerCounts().nodes <= 0`.

## The liveness question, settled

`viewFrameReady()` proves Go is alive NOW; a file proves only that Go ran at some point. That
difference is real but it is **not a property this system has** — `viewFrameReady` has exactly
one caller, and every other reader (nodes, edges, camera, overlays, owner counts) reads
persisted files with no liveness gate. The render tree already draws from the previous run's
files before Go starts, everywhere. NavGuides is the lone exception, not a guarantee being
given up. Removing it makes NavGuides behave as the rest of the tree already does.

So: no liveness signal replaces the frame. If one is ever wanted it is a binary file like
everything else, not a stream.

## Ripple, in order

1. **TS render tree** — drop `viewFrameReady()` from NavGuides, leaving the file-based
   `ownerCounts().nodes <= 0` gate. Delete `view-frame-ready.ts`, `buffer-decode-view.ts`,
   `buffer-decode-node.ts`'s frame half (keep `nodeLabel`), `buffer-decode-interior.ts`'s
   frame half (keep the `INTERIOR_SLOTS_PER_NODE` re-export), `snapshot-buffer.ts`'s frame
   cache.
2. **Extension host** — delete `runner/`'s framing, demux, dispatchers and last-frame cache;
   the spawn keeps stdio for stderr and the error log only.
3. **Go** — delete `runtopology/streamwire/`, the four `*_stream.go` wirings, the frame
   builders, and every `WireStream`/`WireBeadStream`/`WireInteriorStream` binding.
4. **`src/Buffer` is deleted.** `src/Buffer/gen`'s three jobs: `generateShadingParams` is the
   only survivor and its output already lands in `src/Node/nodegeom/`, so the generator moves
   to `src/Node/nodegeom/gen`. `generateBufferLayout` and `generateFrameTags` go.
5. **Guards** — `check-buffer-layout-parity.sh` and `check-overlay-row-struct.sh` guard the
   layout; both go with it. `check-no-dead-buffer-column.sh` keeps its `*_values.go` half.
6. **Docs** — `.claude/rules/buffer-schema.md` describes a buffer that no longer exists;
   delete it. CLAUDE.md's `src/Buffer/` bullet and bridge-surface wording, MODEL.md, and
   `.claude/rules/bridge-surface.md` all say Go→TS is block files **plus trace-event
   streams**; the second half goes.

## Verification

`scripts/stop-checks.sh` empty is necessary, not sufficient — it cannot see a scene that
stopped updating. The real check is driving the editor: nodes and edges draw and move, beads
animate, overlays toggle, nav guides appear, and `.probe` breadcrumbs still arrive via
`scripts/probe-merge.sh --debug`.

## Risks

- **Silent stall.** If some render path depended on a frame arriving to schedule a redraw
  rather than on a file changing, the scene freezes while every check stays green. Grep for
  `buffer-snapshot` consumers before deleting the dispatchers.
- **Startup order.** Files persist, so the first paint may use the previous run's values for a
  frame or two. That is already true everywhere else.
- **Reverses a stated invariant.** The bridge surface is written down in three places as
  "block files plus trace-event streams"; all three change in this commit or the docs lie.
