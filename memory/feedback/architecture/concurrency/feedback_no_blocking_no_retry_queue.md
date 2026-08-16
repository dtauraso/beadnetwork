---
name: no-blocking-no-retry-queue
description: In the node network, never propose a blocking receive, a wake channel, or a retry queue. Non-blocking polls only; a continuous quantity rides a depth-1 coalescing slot, a discrete event rides a small queue whose overflow panics.
type: feedback
---

**Rule:** When a producer outruns a consumer in the network, the fix is a **depth-1 coalescing slot**, never any of: a blocking receive, a "block on a wake channel then drain" loop, a retry/pending queue, or a raised bound. David, 2026-08-15, on a proposal containing two of those: *"no blocking. no blocking wakes. not even a retry queue."*

Sort the traffic by whether the quantity is continuous:

- **Continuous** (an incremental delta — a neighbour's `KindCenter`, the FSM's `KindDrag`): N of them are exactly equivalent to ONE of their sum, so they get a depth-1 slot that merges on deposit. A slow reader skips intermediate values and lands in the same place. Deposits cannot fail, so nothing retries.
- **Discrete** (select, hover, dragStart/dragEnd, tilt): not summable, must not be dropped, so a small queue — and they arrive at human *decision* rate, so a full one means something else is broken. Overflow panics by name; it does not block and does not grow.

**Why:** MODEL.md already says both halves of this — "Blocking couples this goroutine's progress to another's" and a violated bound means the code is wrong rather than something to grow or defer. I proposed a blocking wake channel anyway, having *just read* the section that forbids it. Reading the model is not the same as deriving from it. Also: a retry queue whose sends cannot fail is dead code wearing a safety net's clothes — deleting it is correct, and the invariant worth keeping loud moves into the slot (a non-summable kind routed onto a coalescing slot panics naming that fact), because that is the mistake a future change would actually make.

**How to apply:** Before proposing any inter-goroutine plumbing here, ask "is this quantity continuous or discrete?" and let the answer pick the structure. If the answer is "I need the reader to wait until there's work", that is the wrong frame — the reader has its own clock. Worked example: the geometry/animation goroutine split (2026-08-15), where dragging a node killed it because pointer-rate deltas fed a FIFO drained once per sim cycle. Related: [[feedback_no_atomics_are_defects]], [[feedback_paced_tryrecv_blocks]], [[feedback_node_model_not_networking_handshake]], [[feedback_go_vs_coordinator_bias]].
