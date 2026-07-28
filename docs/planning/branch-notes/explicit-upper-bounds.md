---
branch: task/explicit-upper-bounds
---

# Every queue, buffer, and loop carries a declared maximum

## The rule

TigerStyle requires an explicit limit on every loop and every queue — no unbounded
growth anywhere, ever. The limit is *declared* (a named constant, not a magic number at
the use site) and *asserted* (exceeding it fails loudly rather than degrading).

The payoff is diagnostic, not just protective. Unbounded growth fails as "it hung and
ate memory" — no location, no cause, a session to bisect. A declared bound fails as
"queue X exceeded MAX_X at N" — the failure names itself.

## Why this repo needs it

There is already direct evidence of the failure mode. It was measured and closed on
`task/edge-stream-volume`, now merged (`6784aafe`) — that branch's note was branch-local
and stripped at merge, so **the durable record is the 2026-07-28 entry in
[session-log.md](../visual-editor/session-log.md)**, not the file this note used to cite.

`.probe/go-edge.jsonl` reached **1.1 GB** in one session against 34 MB for the next
largest log — ~30x everything else combined, discovered by noticing disk usage rather
than by any check firing. That is the shape this branch exists to prevent.

Two numbers from the original framing were wrong, and the corrections matter here because
they change what a bound would have to be:

- **The rate is not a constant.** The headline 734 KB/s (~258 MB/hr) did not reproduce; a
  later session measured 75 KB/s from the same code. Volume is
  `beads_in_flight x 50/sec x ~186 bytes`, so the high figure was a saturated ring. **A
  fixed byte/sec ceiling would therefore be the wrong bound** — the quantity that varies
  is beads in flight, not the rate.
- **The cadence was never the problem.** Sampling ran at 20.3 ms (~50 Hz), matching the
  documented ~16 ms. The multiplier is bead *lifetime*: ~20 s in flight, so one bead
  alone produces ~1000 rows.

The existing responses are **external**: `tools/run-bounded.sh` wraps a run to cap it
from outside, and the log-flood lesson is recorded doctrine. This branch moves the bound
**inside** — the queue knows its own limit.

## Scope

Sweep for anything that can grow without a stated ceiling:

- **channel buffers** in the node network — each should have a named capacity constant,
  and the choice should be justified in a comment (why *this* number)
- **content buffers / per-owner streams** — node, edge, interior, VIEW
- **retry and wait loops** — a bound on iterations, not only on wall time
- **any accumulating slice or map** on a per-tick path — the classic silent leak
- **the probe/trace sinks** — the 1.1 GB case above

Cross-check against `tools/check-no-dead-buffer-column.sh` and
`tools/check-buffer-layout-parity.sh` for where buffer structure is already pinned.

## Design questions to answer, not assume

These need a decision per site rather than a blanket policy — write down which you chose
and why:

- **What happens at the bound?** Panic (a bug — the bound was wrong or something leaked),
  drop-oldest, drop-newest, or apply backpressure to the producer. For a simulation with
  a paced clock these are *not* equivalent: dropping changes what gets rendered, blocking
  changes timing. See `docs/backpressure-investigation-order.md` before choosing.
- **Is the edge-stream volume inherent or a bug?** The open question on
  `task/edge-stream-volume` is whether something emits per-frame where it should emit
  per-change. A bound doesn't answer that — but don't paper over it with a bound either.
  Coordinate rather than duplicate.
- **Debug vs. production paths.** Breadcrumbs are meant to be sparse (CLAUDE.md); the
  bound on a debug sink can be much tighter than on a live data path.

## Guidance

- Name the constant near what it bounds. A bound defined far from its queue drifts.
- The bound and its assertion land together — a limit nobody checks is a comment.
- Prefer failing at the bound over growing "just this once." The point is that the
  failure is legible.

## Ordering note

Overlaps `task/runtime-invariant-assertions` on the node/buffer hot path, and
`task/synctest-deterministic-tests` rewrites the tests that would exercise both. Landing
synctest first makes a bound-exceeded case reproducible on demand instead of
occasionally.

## Done when

- No queue, buffer, or loop on the hot path grows without a declared, asserted maximum.
- Each bound has been exceeded once deliberately to confirm it fires and names itself —
  `memory/feedback_check_the_signal_the_check_emits.md`.
- `bash scripts/verify.sh` is clean.
