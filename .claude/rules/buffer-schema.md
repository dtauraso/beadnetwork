---
paths:
  - "src/Buffer/**/*.go"
  - "src/Buffer/buffer-layout.ts"
  - "src/webview/**/*.tsx"
  - "src/webview/**/*.ts"
---

# Content buffer schema — adding or changing a column

**A block's columns live in the directory of the thing they describe** — a column is part of
its concern, not part of a schema directory.

**Chrome has left the buffer entirely**, and so has the geometry that was per-row. Each of
these crosses as a block FILE whose layout is its generated value list, found through that
piece's `paths/block.bin`: the rules panel, both dropdowns, the tab strip, the fit chip, the
tilt panel, the overlays popover, the pointer target, the speed slider, the interior beads
(one file per node) and the edges (one file per edge). The row sits in the PATH, not in the
file, so the goroutine that owned the stream is still the only writer.

**The read cadence is the reason a file can serve any of them.** `makeLeafValues` and
`makeRowLeafValues` take `"interval"` (a hundred milliseconds, for what changes at human
speed) or `"frame"` (requestAnimationFrame, for what follows the cursor, a drag, or a tick).
What remains in the buffer is the per-tick geometry it exists for: the model blocks (Node,
Camera, Scene…) and the trace events.

The generator does not need telling where they are. It walks the repo for `bufLayout*`
structs, so a block is found by BEING one — a list of locations would be a second
description of the layout, and this generator exists because two descriptions of the wire
format is exactly how the Go and TS halves drift apart.

**Two things stay central, and only two.** `BufLayoutVersion` in
`src/Buffer/layout_version.go`, and `bufBlockOrder` in
`src/Buffer/gen/buflayout/buf_layout_parse.go` — that order IS the wire format, so it
belongs in one place. Where a block's file sits is not part of the wire format.

`bufBlockOrder` now names only the trace events (Recv, Fire, Send, Arrive, Breadcrumb).
Everything else that once had a `buffer_block.go` has a `*_values.go` instead: a Go name
list, a generated TS name list beside it, and a `paths/block.bin` naming the file. To add a
value, add its name to that list and regenerate in the SAME commit — `check-generated.sh`
fails if the generated half is stale, and `check-no-dead-buffer-column.sh` fails if nothing
reads the new name.

A new BLOCK also needs its name in `bufBlockOrder`, and a `var _ = bufLayoutX{}` beside it
so it is not dead code in a package nothing imports it from.

## Which guard catches what

`check-buffer-layout-parity.sh` does **NOT** catch an unregenerated column. It only asserts
the two GENERATED files carry the same `BUF_LAYOUT_FINGERPRINT` — it never reads
`layout.go`. So it catches a botched/partial regen or a deleted generated file, not a column
you added to `layout.go` and forgot to regenerate. `check-generated.sh` is the one that
catches that.

## Every column needs a production consumer

A generated READER exists for every column automatically; nothing forces a production
CONSUMER, but `check-no-dead-buffer-column.sh` enforces one — a generated `read*` helper
with no non-test `src/` caller fails that guard. The allowlist is empty. A dead column like
the former `Bead.Live` gets deleted from the schema, not allowlisted.

So an unconsumed column is a guard failure, not silent dead surface — the last of the
grep-discoverable edits that the four-step node-kind rule doesn't cover is machine-checked.
