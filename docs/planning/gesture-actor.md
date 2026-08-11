---
branch: task/god-objects
---

# Making the gesture FSM its own actor

The last 20 of `MoveDispatch`'s 37 methods are one subsystem — the gesture → view loop — and
it resists every decomposition tool used on this branch because it is a CYCLE, not a
hierarchy: read raw input → mutate UI state → emit a frame, with each step touching the same
state. Extract-type and extract-package pull a leaf off a tree; there is no leaf here.

A cycle is cut with a channel, not a function signature. That is what this change does, and it
is the shape MODEL.md already uses everywhere else.

## What is already true (measured, not assumed)

**The FSM is already actor-shaped with respect to the movers.** It reads mover-owned state
through exactly two accessors:

- `mr.centerOfNode(id)` — reads `centerMirror`, *"the DISPATCH goroutine's OWN mirror of every
  node's last-known world center, kept current by messages from each node's own goroutine"*
  (`mover_registry.go`), drained non-blockingly from each mover's `centerOut` channel.
- `mr.nodeBodyRadius(id)` — a static per-node dimension.

So there is no shared-memory read to remove. The mirror-fed-by-channels pattern this change
would need already exists and is already in use.

**`RunStdinReader` is already framing + routing only.** It takes a `stdinreader.Handlers`
struct of three func values (`ApplyEdit`, `HandleRawInput`, `HandleSave`) and imports nothing
from `Wiring`. The seam a channel would replace is already a seam.

## The one real problem: who owns the VIEW stream

MODEL.md pins one writer per stream. Today the view-owner is `RunStdinReader`'s goroutine, and
BOTH the gesture handlers and the stdin dispatch handlers emit on it — they are the same
goroutine, so that is currently safe. Split the FSM off and there are two writers unless
ownership moves with it.

Emit sites today: `viewpoint_state.go` (5), `gesture_actions.go` (2), `gesture_handlers.go`
(1), `gesture_graph.go` (1) — the FSM; plus `stdin_dispatch.go` (2), `scene_*_persist.go` (4),
`scene_structure.go` (1), `distance_groups.go` (1), `view_stream.go` (1) — the dispatch side.

**Decision: the gesture goroutine becomes the view owner.** It is the dominant emitter
(camera, selection, hover, drag are all gesture), and the alternative — FSM computes, sends a
result, dispatch emits — puts a channel round-trip inside every pointer-move, which is the
tightest loop in the editor.

`RunStdinReader` then does what its name says: frame bytes, decode, route. Raw input goes to
the FSM's channel. Non-raw edits (overlay toggle, clock speed, scene select, structural edit)
ALSO go to the FSM's channel, because they mutate state the frame carries. That is one inbox,
one owner, one writer — the same shape as a `nodeMover`.

## Order

1. **Give the FSM its own inbox and goroutine**, receiving a message union of {raw-input,
   edit, save}. `RunStdinReader`'s three `Handlers` funcs become three channel sends. No
   behaviour change yet — the FSM still calls the same code, just on its own goroutine.
2. **Move VIEW-stream ownership to it.** `SetViewStream` targets the FSM. Every `emitViewFrame`
   call site must then be on the FSM goroutine; any that is not is a bug this step must find,
   not paper over.
3. **Only then** re-measure `MoveDispatch`. The 20 gesture/view methods become the FSM's own,
   and `uiState`/`gestureState` lift with it — the probe that failed does so because
   `gestureState`'s 18 private fields are read at 49 sites by handlers that would now live
   in the same package.

## Verification — the part that cannot be skipped

- **`go test -race ./...`** on every step. This is the only change on this branch that adds a
  goroutine; the existing suite does not exercise concurrency (per `docs/process/testing-shape.md`
  the repo deliberately does not test cross-goroutine behaviour), so the race detector is the
  only automated signal.
- **Exactly one writer per stream** must remain true. `memory/feedback/architecture/feedback_no_single_writer_bridge.md`
  and MODEL.md both pin it. State how it was checked, not that it holds.
- **Drive the editor.** Pointer-move is the tightest loop here and the suite asserts nothing
  about it. A regression shows up as lag, a dropped drag, or a stale frame — none of which any
  test will catch.
- The `.probe` logs (`.claude/rules/go-debugging.md`) are the diagnostic if it misbehaves.

## Risks

- **This adds a goroutine to the interaction path.** Every prior step on this branch was pure
  code motion with `no behaviour change` as the claim. This one changes when work happens.
  It is not reversible by `git revert` once the editor has been used against it.
- **Deadlock**: the FSM sends to movers and movers push to `centerOut`, which the FSM drains.
  If the FSM ever blocks on a send while a mover blocks on a push, both stop. Every send on
  this path must stay non-blocking (`SendLatestNonBlocking` / `select` with `default`), which
  is the existing convention — verify it, do not assume it.
- **If step 2 finds an `emitViewFrame` caller that cannot move to the FSM goroutine**, stop.
  That is a real second writer and the design is wrong as stated.
