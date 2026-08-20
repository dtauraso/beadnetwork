# Input is sampled, not delivered

## The target

TS → Go carries the CURRENT input state in files Go samples, the way Go → TS
now carries the camera. No pipe, no inbox, no backlog — a file cannot lag,
because there is nothing to drain.

## The invariant this reverses

CLAUDE.md pins the seam:

> **TS → Go** is framed binary records on stdin … The TS → Go send is
> **fire-and-forget** — no `await`, no Promise chain, no request/response, no
> delivery signal (guard: `src/check-no-await-on-bridge.sh`).

Fire-and-forget stays true and becomes MORE true: writing a current value is
the purest form of it. What changes is the medium and, with it, the semantics.

## Why

Two queues sit between a wheel event and the camera moving:

1. **the pipe** — TS writes records fire-and-forget, so it can outrun Go by
   construction, and the records accumulate in the OS pipe buffer.
2. **the gesture inbox** — `gestureInboxDepth = 64`, `sendGestureMsgBlocking`,
   drained one at a time by the actor goroutine.

Both are delivery semantics: every event, in order, eventually. Applied to
input that means replaying history — sixty-four queued wheel deltas are sixty
four zoom steps still to apply after the fingers stop. That is the observed
lag, and it is the queue doing its job, not failing at it.

The repo already rejects this shape twice. `neighborSlot` is depth 1 with a
stated merge rule — "a WHERE collapses by keeping the newest, a HOW FAR by
summing" — and `memory/.../feedback_node_model_not_networking_handshake.md`
says no delivery guarantees. A 64-deep inbox is a delivery guarantee.

## The distinction that decides the design

Not all input is state.

| input | shape | as a file |
|---|---|---|
| pointer x/y, rect, fov | a WHERE | current value; sampling is exact |
| wheel delta | a HOW FAR | NOT a current value — see below |
| pointerdown / pointerup | a TRANSITION | must not be sampled away |

**A delta is not a value.** "How much did you scroll since I last looked" has no
current reading. TS writes a running TOTAL instead, and Go takes the difference
from the total it last saw; then it is a value, it samples exactly, and no
scroll is ever lost or replayed.

**A transition is not a value either.** A pointerdown followed by a pointerup
between two samples would vanish, and a vanished gesture is worse than a lagged
one. Same answer: a monotonically increasing counter per button, so Go sees
that a transition happened even if it missed the instant.

Getting this wrong drops gestures rather than lagging them, which is why the
merge rule is stated per kind before any code.

## Who writes the file

The webview has no filesystem — it can READ by path but not write. So the write
is: webview → `postMessage` → extension host → host writes the file → Go
samples. The host relay stays; the pipe and both queues go.

That asymmetry is the platform's, not a design choice: reads cross the sandbox,
writes do not.

## Order

1. The wheel path alone, since it is the one with observable lag: TS writes a
   scroll TOTAL, Go samples and differences it.
2. Pointer position and rect — plain current values.
3. Button transitions with their counters.
4. Then the inbox and the stdin record path come out, not before.

## Verification

Zoom continuously for several seconds and stop. The screen must stop when the
fingers stop — no catch-up. Then: press and release fast enough to fall between
two samples and confirm the gesture still registers, which is the failure this
design can introduce and the stream cannot.

`bash scripts/stop-checks.sh` empty.

## Risk

`check-no-await-on-bridge.sh` guards fire-and-forget on the TS side; a file
write through the host must not grow an await or a completion signal.
