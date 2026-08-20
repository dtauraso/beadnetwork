# Input is a file the gesture goroutine reads on its own clock

## The target

TS → Go carries the current input in ONE file. The gesture goroutine wakes on
its own `Copy()` of the clock, reads that file, and acts — the way every node
kind already wakes, does local work, and sleeps. No pipe, no inbox, no backlog.

## The invariant this reverses

CLAUDE.md pins the seam:

> **TS → Go** is framed binary records on stdin … The TS → Go send is
> **fire-and-forget** — no `await`, no Promise chain, no request/response, no
> delivery signal.

Fire-and-forget stays and becomes more literal: writing a current value is the
purest form of it. The medium changes, and with it the semantics.

## Why

Two queues sit between a wheel event and the camera moving:

1. **the pipe** — TS writes fire-and-forget, so it outruns Go by construction
   and records accumulate in the OS pipe buffer.
2. **the gesture inbox** — `gestureInboxDepth = 64`, drained one at a time.

Both are delivery semantics: every event, in order, eventually. For input that
means replaying history — queued wheel deltas are zoom steps still to apply
after the fingers stop, which is the observed lag. That is the queue working,
not failing.

The queue exists because the gesture actor blocks on a channel instead of
pacing itself. It is the one goroutine in this program that waits to be handed
work rather than waking and looking. Everything else — every node kind — holds
a clock copy and calls `SleepCycle`.

## What wakes the reader

Its own clock, which it already has. This is the same answer the camera gave in
the other direction: the renderer reads the camera's files every frame because
`requestAnimationFrame` already gave it a reason to wake. Neither side needs a
notification, a watch, or a queue — both already had a heartbeat.

That collapses the problem I first wrote this plan around. There is no "between
two reads": there is what the input IS when the goroutine looks, which is the
same relationship a node has to its inputs. No transition counters, no running
totals to difference — those were solving a problem created by framing this as
sampling.

## One file, not one per field

An input event's fields describe ONE moment: `x`, `y`, the rect, the modifiers,
and the raycast hit come from a single browser event. Read `x` from one moment
and `hit` from another and Go acts on an input that never happened — a click
against the wrong node. That is the split-pivot defect, and worse here, since a
wrong hit selects the wrong thing rather than shifting the camera slightly.

Values that change together arrive together: one file, written whole.

## Who writes it

The webview can READ by path but cannot WRITE — that asymmetry is the
platform's. So: webview → `postMessage` → extension host → host writes the file
→ the gesture goroutine reads it. The host relay stays; the pipe and both
queues go.

## Order

1. The file and the read: the gesture goroutine paces on its clock and reads
   the current input instead of blocking on the inbox.
2. The write: the host writes the file instead of the stdin record.
3. The inbox, `sendGestureMsgBlocking`, and the raw-input stdin path come out —
   not before the first two are live.

## Verification

Zoom continuously for several seconds and stop. The screen stops when the
fingers stop; no catch-up. Then drag a node and confirm it tracks the pointer,
which is the path with the most fields riding on one moment.

`bash scripts/stop-checks.sh` empty.

## Risk

`check-no-await-on-bridge.sh` guards fire-and-forget on the TS side; a file
write through the host must not grow an await or a completion signal.
