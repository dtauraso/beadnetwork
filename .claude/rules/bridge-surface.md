---
paths:
  - "src/Node/Wiring/**/*.go"
  - "src/extension/messages.ts"
  - "src/extension/handle-message.ts"
  - "src/extension/runCommand.ts"
  - "src/Node/wire-gen.ts"
---

# Bridge surface — TS → Go vocabulary detail

The Go → TS invariant (per-goroutine streams, no single writer) is in CLAUDE.md. This file
carries the TS → Go vocabulary.

**TS → Go** is framed binary records on stdin. Two shapes, and the distinction is the model:

- **Addressed edits** — a single geometry-CRUD `edit` message whose sole op is `update`
  (see `src/Scene/scenerun/dispatch_edit.go` `applyEdit`, fenced by `EDIT_OPS_START`/
  `EDIT_OPS_END`, and `src/extension/messages.ts` `EditMsg`): **`update` sets
  an ATTRIBUTE on a typed entity** (`kind` = node / edge / camera / overlays / panels /
  scene) — there is no per-feature op. New *addressed* capability is a new entity kind or
  attribute, NOT a new op. `panels` is its OWN entity kind, deliberately separate from
  `overlays`: it addresses the overlays popover's disclosure open/closed state, not overlay
  visibility, and is its own hand-written `viewstate.PanelState` (not generated, unlike
  `OverlayState`), persisted to its own file
  under `view/` (`.claude/rules/persistence-ownership.md`).
- **Bare commands** — `save` is the only bare command. It is defined end-to-end (kind byte,
  Go decode + persist) but currently has **no live TS sender** — no UI affordance posts it
  yet; it stays in the vocabulary because Go's decode and the `INPUT_LAYOUT_FINGERPRINT`
  both carry it. It carries **no entity id on purpose**: it acts on state **Go already
  owns** (the current selection / scene), so there is nothing for TS to address. There is no
  `resend` command, and now nothing to resend: a remounted webview re-reads the files, which
  are the current state by definition. The ext host used to cache the last frame per stream
  and replay them all on `ready`; that cache is gone with the frames.

  (Several ops/commands were removed end-to-end with no live TS sender — `edit-create`/
  `edit-delete`, `play`/`pause`, `run`/`stop`, `fade-toggle` — and their kind bytes are left
  as GAPS in `input_codec.go`, never renumbered.)
- **`raw-input`** — raw pointer/wheel + stateless raycast hit → Go's gesture FSM. Camera
  orbit, node moves, and port-anchor moves are produced **in-process** by the FSM
  from raw-input; they do not cross this seam as edits.

## No sidecar

There is **no id/label/kind sidecar**: node identity is the ROW, which is the directory its
file sits in, and the human label is the `label` value in that same file. The loader/mover enforce
`NodeId == row + 1` by construction (`ROW ID = NODE ID - 1`,
`.claude/rules/persistence-ownership.md`), so the row is the identity.

A `NodeId` COLUMN briefly existed to close a hole the row alone could not: a bare row could
never be CONTRADICTED by the frame that carried it, so a misrouted or permuted stream would
render silently against the wrong node. **That hole is now closed by construction, and the
column is not coming back.** There is no stream to misroute: node 3's state is the file at
`view/nodes/3/node.bin`, and the renderer reads row 3 by asking for that path. The row is not
a claim travelling beside the data that could disagree with where the data arrived — it IS
the address. Demuxing a frame onto the wrong fd was the failure this worried about, and there
are no fds.

Nothing crosses Go → TS but files. There is no frame, no stream inventory, no frame tag, and
no host→webview message of any kind.

## Parity

Keep all of it in parity across `messages.ts`, the `src/Node/Wiring` stdin reader/dispatch
(`stdin_reader.go`'s `MSG_TYPES` fence, `dispatch_edit.go`'s edit tables), and `handle-message.ts`
(guards: `src/Scene/scenerun/check-edit-op-parity.sh`, `src/Scene/scenerun/check-message-kind-parity.sh`, and the
`INPUT_LAYOUT_FINGERPRINT` in `input_codec.go` /
`src/Node/wire-gen.ts`).
