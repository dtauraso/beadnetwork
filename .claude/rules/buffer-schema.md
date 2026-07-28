---
paths:
  - "Buffer/**/*.go"
  - "tools/topology-vscode/src/schema/buffer-layout.ts"
  - "tools/topology-vscode/src/webview/three/**/*.tsx"
  - "tools/topology-vscode/src/webview/three/**/*.ts"
---

# Content buffer schema — adding or changing a column

If a change needs a **new column in the content buffer**, add it to the hand-authored
schema (`Buffer/layout.go`, the `buf:"…"` struct tags) and regenerate in the same commit.
`Buffer/buffer_layout_gen.go` and `tools/topology-vscode/src/schema/buffer-layout.ts` are
BOTH generated from `layout.go`, and `check-generated.sh` fails if either is stale relative
to it (it regenerates and diffs).

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
