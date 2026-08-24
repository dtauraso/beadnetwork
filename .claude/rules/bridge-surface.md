---
paths:
  - "Categories/Scene/Drag/**/*.go"
  - "Start/extension/messages.ts"
  - "Start/extension/handle-message.ts"
  - "Start/extension/runCommand.ts"
  - "Categories/Scene/Drag/input-defs.ts"
  - "Categories/Node/update-attrs.ts"
---

# Bridge surface — TS → Go vocabulary detail

The Go → TS invariant (per-goroutine streams, no single writer) is in CLAUDE.md. This file
carries the TS → Go vocabulary.

**TS → Go** is framed binary records on stdin. Two shapes, and the distinction is the model:

- **Addressed edits** — a single geometry-CRUD `edit` message whose sole op is `update`
  (see `Categories/Scene/Dispatch/dispatch_edit.go` `applyEdit`, fenced by `EDIT_OPS_START`/
  `EDIT_OPS_END`, and `Start/extension/messages.ts` `EditMsg`): **`update` sets
  an ATTRIBUTE on a typed entity** (`kind` = node / edge / camera / overlays / panels /
  scene) — there is no per-feature op. New *addressed* capability is a new entity kind or
  attribute, NOT a new op. `panels` is its OWN entity kind, deliberately separate from
  `overlays`: it addresses the overlays popover's disclosure open/closed state, not overlay
  visibility, and is its own hand-written `Panel.PanelState` (not generated, unlike
  `OverlayState`), persisted to its own file
  under `view/` (`.claude/rules/persistence-ownership.md`).
- **Bare commands** — `save` is the only bare command. It is defined end-to-end (kind byte,
  Go decode + persist) but currently has **no live TS sender** — no UI affordance posts it
  yet; it stays in the vocabulary because Go's decode and
  `Categories/Scene/Drag/stdin_reader.go` both carry it. It carries **no entity id on purpose**: it acts on state **Go already
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

Keep all of it in parity across `messages.ts`, the `Categories/Scene/Drag` stdin reader and `Categories/Scene/Dispatch`
(`stdin_reader.go`'s `MSG_TYPES` fence, `dispatch_edit.go`'s edit tables), and `handle-message.ts`
(guards: `Categories/Scene/Dispatch/check-edit-op-parity.sh`, `Categories/Scene/Dispatch/check-message-kind-parity.sh`).

There is no single wire-layout string any more. Each vocabulary is declared in Go by the
concern that uses it, and generated into TS beside it, so the TS copy cannot drift from
the Go one. Two kinds of vocabulary, two kinds of generated file:

- **What an edit can address** is Drag's, because Drag declares it: the record kinds in
  `kinds.go` and `stdin_reader.go`. `Categories/Scene/Drag/inputdefs` writes it once to
  `Categories/Scene/Drag/input-defs.ts` — `IN_KIND_RAW_INPUT`, `IN_KIND_EDIT_UPDATE`,
  `IN_EVENT_KINDS`, `IN_HIT_KINDS`, `IN_UPDATE_KINDS` — and every sender imports it from
  there. It is one number and one entity list for the whole bridge; a per-concern copy of
  either could disagree with the reader that decodes it.
- **Which attributes that concern will accept** is the concern's own, declared in its
  `update_attrs.go`. Its `updateattrs` generator writes only that list, to its own
  `update-attrs.ts`. These are deliberately per-concern: Node's attrs need agree with
  nothing but Node.
