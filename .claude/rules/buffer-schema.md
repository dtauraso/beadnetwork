---
paths:
  - "src/schema/buffer-layout/**/*.go"
  - "src/schema/buffer-layout/buffer-layout.ts"
  - "src/webview/**/*.tsx"
  - "src/webview/**/*.ts"
---

# Content buffer schema — adding or changing a column

**A block's columns live in the directory of the thing they describe.** The rules panel's
columns are `src/Chrome/PolarRulesPanel/buffer_block.go`, the overlays popover's are
`src/Chrome/OverlaysDropdown/buffer_block.go`, and so on — a column is part of its concern, not part of
a schema directory. What remains in `Buffer/bufschema/` is what has no owner elsewhere: the
model blocks (Node, Edge, Camera, Scene…) and the trace events.

The generator does not need telling where they are. It walks the repo for `bufLayout*`
structs, so a block is found by BEING one — a list of locations would be a second
description of the layout, and this generator exists because two descriptions of the wire
format is exactly how the Go and TS halves drift apart.

**Two things stay central, and only two.** `BufLayoutVersion` and `BufInteriorSlotsPerNode`
in `src/schema/buffer-layout/layout_version.go`, and `bufBlockOrder` in
`src/schema/buffer-layout/gen/buflayout/buf_layout_parse.go` — that order IS the wire format, so it
belongs in one place. Where a block's file sits is not part of the wire format.

To add a column: add the field with its `buf:"…"` tag to that block's `buffer_block.go`, and
regenerate in the SAME commit. `buffer_layout_gen.go` and `buffer-layout.ts` are both
generated, and `check-generated.sh` fails if either is stale (it regenerates and diffs).

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
