---
paths:
  - "nodes/Wiring/**/*.go"
  - "tools/topology-vscode/src/messages.ts"
  - "tools/topology-vscode/src/extension/handle-message.ts"
  - "tools/topology-vscode/src/runCommand.ts"
  - "tools/topology-vscode/src/schema/input-layout.ts"
---

# Bridge surface — TS → Go vocabulary detail

The Go → TS invariant (per-goroutine streams, no single writer) is in CLAUDE.md. This file
carries the TS → Go vocabulary.

**TS → Go** is framed binary records on stdin. Two shapes, and the distinction is the model:

- **Addressed edits** — a single geometry-CRUD `edit` message whose sole op is `update`
  (see `nodes/Wiring/stdin_reader.go` `applyEdit`, fenced by `EDIT_OPS_START`/
  `EDIT_OPS_END`, and `tools/topology-vscode/src/messages.ts` `EditMsg`): **`update` sets
  an ATTRIBUTE on a typed entity** (`kind` = node / edge / camera / overlays / scene) —
  there is no per-feature op. New *addressed* capability is a new entity kind or
  attribute, NOT a new op.
- **Bare commands** — `save` is the only bare command. It is defined end-to-end (kind byte,
  Go decode + persist) but currently has **no live TS sender** — no UI affordance posts it
  yet; it stays in the vocabulary because Go's decode and the `INPUT_LAYOUT_FINGERPRINT`
  both carry it. It carries **no entity id on purpose**: it acts on state **Go already
  owns** (the current selection / scene), so there is nothing for TS to address. There is no
  `resend` command: the ext host caches the last frame per dedicated stream (view, plus one
  per edge/node/interior row) and replays all of them to a remounted webview on `ready`
  instead (`BuildAndRunRunner.getLastViewFrame`/`getLastEdgeFrames`/`getLastNodeFrames`/
  `getLastInteriorFrames` in `tools/topology-vscode/src/runCommand.ts`) — Go only ever emits
  a frame when something changes, and that stays true.

  (Several ops/commands were removed end-to-end with no live TS sender — `edit-create`/
  `edit-delete`, `play`/`pause`, `run`/`stop`, `fade-toggle` — and their kind bytes are left
  as GAPS in `input_codec.go`, never renumbered.)
- **`raw-input`** — raw pointer/wheel + stateless raycast hit → Go's gesture FSM. Camera
  orbit, node moves, and port-anchor moves are produced **in-process** by the FSM
  from raw-input; they do not cross this seam as edits.

## No sidecar

There is **no id/label/kind sidecar**: node identity is the buffer's **row index**, kind is
a numeric column, and the human label rides the buffer's Label section via off/len columns
on the Node block. (A test asserts the removed sidecar message is rejected; do not
reintroduce one.)

Full architecture (frame shape, stream inventory, synthetic frame tags) is canonical in
MODEL.md's "Editor surface (TS)" section; see also `memory/feedback_no_single_writer_bridge.md`
and `memory/feedback_per_goroutine_bridge.md`.

## Parity

Keep all of it in parity across `messages.ts`, `stdin_reader.go`, and `handle-message.ts`
(guards: `tools/check-edit-op-parity.sh`, `tools/check-message-kind-parity.sh`, and the
`INPUT_LAYOUT_FINGERPRINT` in `input_codec.go` /
`tools/topology-vscode/src/schema/input-layout.ts`).
