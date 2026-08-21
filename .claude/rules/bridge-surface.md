---
paths:
  - "src/Node/Wiring/**/*.go"
  - "src/Input/messages.ts"
  - "src/extension/handle-message.ts"
  - "src/extension/runCommand.ts"
  - "src/Input/input-layout-gen.ts"
---

# Bridge surface — TS → Go vocabulary detail

The Go → TS invariant (per-goroutine streams, no single writer) is in CLAUDE.md. This file
carries the TS → Go vocabulary.

**TS → Go** is framed binary records on stdin. Two shapes, and the distinction is the model:

- **Addressed edits** — a single geometry-CRUD `edit` message whose sole op is `update`
  (see `src/Input/dispatch/dispatch_edit.go` `applyEdit`, fenced by `EDIT_OPS_START`/
  `EDIT_OPS_END`, and `src/Input/messages.ts` `EditMsg`): **`update` sets
  an ATTRIBUTE on a typed entity** (`kind` = node / edge / camera / overlays / panels /
  scene) — there is no per-feature op. New *addressed* capability is a new entity kind or
  attribute, NOT a new op. `panels` is its OWN entity kind, deliberately separate from
  `overlays`: it addresses the overlays popover's disclosure open/closed state, not overlay
  visibility, and is its own hand-written `viewstate.PanelState` (not generated, unlike
  `OverlayState`), streamed in its own Panel buffer block and persisted to its own file
  under `view/` (`.claude/rules/persistence-ownership.md`).
- **Bare commands** — `save` is the only bare command. It is defined end-to-end (kind byte,
  Go decode + persist) but currently has **no live TS sender** — no UI affordance posts it
  yet; it stays in the vocabulary because Go's decode and the `INPUT_LAYOUT_FINGERPRINT`
  both carry it. It carries **no entity id on purpose**: it acts on state **Go already
  owns** (the current selection / scene), so there is nothing for TS to address. There is no
  `resend` command: the ext host caches the last frame per dedicated stream (view, plus one
  per edge/node/interior row) and replays all of them to a remounted webview on `ready`
  instead (`BuildAndRunRunner.getLastViewFrame`/`getLastEdgeFrames`/`getLastNodeFrames`/
  `getLastInteriorFrames` in `src/extension/runCommand.ts`) — Go only ever emits
  a frame when something changes, and that stays true.

  (Several ops/commands were removed end-to-end with no live TS sender — `edit-create`/
  `edit-delete`, `play`/`pause`, `run`/`stop`, `fade-toggle` — and their kind bytes are left
  as GAPS in `input_codec.go`, never renumbered.)
- **`raw-input`** — raw pointer/wheel + stateless raycast hit → Go's gesture FSM. Camera
  orbit, node moves, and port-anchor moves are produced **in-process** by the FSM
  from raw-input; they do not cross this seam as edits.

## No sidecar

There is **no id/label/kind sidecar**: node identity is the buffer's ROW INDEX, and the human
label rides the buffer's Label section via off/len columns. The loader/mover enforce
`NodeId == row + 1` by construction (`ROW ID = NODE ID - 1`,
`.claude/rules/persistence-ownership.md`), so the row is the identity.

A `NodeId` COLUMN briefly existed to close a hole the row alone cannot: a bare row can never be
CONTRADICTED by the frame that carried it, so a misrouted or permuted stream would render
silently against the wrong node. **That column is gone, and the hole is open.** This paragraph
used to claim `decodeNodeStreamFrame` compared the stated id against the arrival row and
reported a mismatch loudly; it never did — that function reads `COL_STREAM_NODE_LABEL` and
compares nothing, so Go packed the column every frame and nothing read it. Writing the check
where the column was read is not possible either: `decodeNodeStreamFrame`'s only caller is
`probe-append.ts` in the EXTENSION HOST, and `column-values.ts` fills its `latest` map only in
the WEBVIEW, so a comparison there reads the zero fallback and can never fire. Closing this
properly means checking the fd-to-row mapping where frames are demuxed
(`src/extension/runner/stream-demux.ts`), not decoding an id out of a column. (A test asserts the
removed id/label/kind SIDECAR MESSAGE is rejected; a column inside the Node block is the
same shape as `KindId`, not a second channel, and does not reintroduce one.)

Full architecture (frame shape, stream inventory, synthetic frame tags) is canonical in
MODEL.md's "Editor surface (TS)" section; see also `memory/feedback/architecture/bridge/feedback_no_single_writer_bridge.md`
and `memory/feedback/architecture/bridge/feedback_per_goroutine_bridge.md`.

## Parity

Keep all of it in parity across `messages.ts`, the `src/Node/Wiring` stdin reader/dispatch
(`stdin_reader.go`'s `MSG_TYPES` fence, `dispatch_edit.go`'s edit tables), and `handle-message.ts`
(guards: `src/Input/dispatch/check-edit-op-parity.sh`, `src/Input/dispatch/check-message-kind-parity.sh`, and the
`INPUT_LAYOUT_FINGERPRINT` in `input_codec.go` /
`src/Input/input-layout-gen.ts`).
