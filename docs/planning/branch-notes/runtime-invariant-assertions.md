---
branch: task/runtime-invariant-assertions
---

# Assert the invariants at runtime, not only structurally

## The gap this closes

The repo enforces invariants in two places today:

1. **`tools/check-*.sh`** — structural guards. They read *source shape*: is the code
   arranged legally. Run at the Stop hook, cover code that never executes.
2. **`memory/`** — recorded doctrine. Advisory; depends on the model weighting it.

Nothing checks **state while it runs**. A guard confirms TypeScript contains no geometry
math; it cannot confirm a node's simtime went backwards on tick 400. Those are different
failure classes, and only the second one produces a subtly wrong frame that still
renders.

TigerStyle's position: assert arguments, return values, preconditions, postconditions,
and invariants — averaging 2+ per function — and **keep them on in normal runs**. Where
types check structure, assertions check logic and state. This branch applies that to the
Go network.

Directly downstream of `memory/feedback_invariants_drive_design.md` and
`memory/feedback_make_bug_class_unrepresentable.md`: an assertion is the cheapest form of
"make it unrepresentable" available when the type system can't express the constraint.

## Scope rule: derive, don't invent

**Only assert invariants that already exist somewhere in the repo.** Every assertion
should trace to a `tools/check-*.sh`, a `memory/` entry, MODEL.md, or CLAUDE.md. This is
not a design round — inventing new rules here means the assertion and the doctrine can
disagree, and there is no way to tell which is wrong.

Sweep these for candidates:

- `tools/check-no-fan-in.sh` — fan-in prohibition
- `tools/check-uniform-pulse-speed.sh` — pulse speed uniformity
- `tools/check-no-network-locks.sh` — ownership replaces locking
- `tools/check-no-state-cache.sh` — no cached state
- `tools/check-send-rule-parity.sh` — send rules
- `tools/check-message-kind-parity.sh` — Go/TS kind agreement
- `memory/feedback_no_atomics_are_defects.md`
- `memory/feedback_no_single_writer_bridge.md`
- `memory/feedback_reflect_dont_create_store.md`
- `memory/feedback_per_emit_simtime_anchoring.md`

## Candidate assertions (verify each against the source before adding)

Stated as claims to confirm, not as settled facts — check the code says this before
asserting it:

- simtime is **monotonic per node** — never decreases across emits on one node's stream
- a node **never writes another node's** buffer column
- an edge has **exactly two endpoints**, and they match the two names encoded in its
  channel name (the naming convention is load-bearing; assert it holds)
- an emit lands on its **owner's** stream — or, for a node with no per-node stream, on
  the VIEW stream, and never on some third stream
- buffer layout parity holds at the point of write, not only at the point of check
  (`tools/check-buffer-layout-parity.sh` covers the static half)

## Guidance

- Assert at the boundary where the invariant becomes true or must hold — not scattered
  through the body. A precondition at function entry and a postcondition at exit beat
  five mid-body checks.
- Failure should be **loud and immediate**: panic with the violated invariant named. A
  silently-wrong frame is the expensive outcome; a panic pointing at the invariant is
  the cheap one.
- Assertions are not a debug channel. `tr.Breadcrumb(...)` is the debug tool (see
  CLAUDE.md, "Debugging the Go layer"). An assertion says "this cannot happen"; a
  breadcrumb says "here is what happened."
- Keep them cheap enough to leave on. If one is too expensive for the hot path, that is
  a signal the invariant belongs in a structural guard instead.

## Ordering note

`task/synctest-deterministic-tests` should probably land first. Under a nondeterministic
scheduler, an assertion that fires once in fifty runs is indistinguishable from one that
never fires — so a passing suite proves less than it looks like it does.

## Done when

- Every added assertion traces to an existing guard or memory entry (list the mapping in
  the commit message).
- Each one has been made to fail once deliberately —
  `memory/feedback_check_the_signal_the_check_emits.md`.
- `bash scripts/verify.sh` is clean.
