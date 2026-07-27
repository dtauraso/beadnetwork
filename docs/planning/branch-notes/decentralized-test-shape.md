---
branch: task/test-teardown-waitgroup
---

# Testing a decentralized goroutine system

What shape a test may take when every node is an independent goroutine. This is
methodology — the MEDIUM, not the substance — so industry practice applies directly
here, unlike the execution model itself (CLAUDE.md, "Medium vs. substance").

## The rule

A test may assert what a goroutine **itself** decided, emitted, or persisted. It may
**not** assert what a *different* goroutine received and did with a bead.

A node calls `PlaceDrivenAt` and returns. The `PacedWire` goroutine owns the bead and
times its own traversal. So "node A fired, therefore the bead arrived at node B's
input" asserts a delivery guarantee the model deliberately does not have — no ack, no
nack, no handshake, no send-gating. Writing that test re-couples two independent
goroutines and re-imposes exactly what decentralizing removed.

## The dividing line: which channel, and does it guarantee anything

| Channel | Guarantee | May a test assert arrival? |
|---|---|---|
| Bead delivery over a `PacedWire` | None. No ack, no retry, no back-pressure. | **No.** Test the sender's own emit decision; test pacing inside `nodes/wire/` where it is that wire's own local property. |
| The `neighborSetC` layout cascade | Never-dropped. A full channel means the message is RETAINED on the sender's own pending queue and retried next cycle (MODEL.md). | **Yes.** Asserting it took effect verifies a documented property, not an invented one. |

So "drag X, wait for neighbor Y's own breadcrumb, read Y's state" is legitimate, while
"place a bead at A, assert B received it" is not. Same structural shape; different
channel; only one of them claims a guarantee.

If that dividing line is ever rejected, roughly six of the eight tests this branch
touches get recategorized and become removals instead of fixes.

## Legitimate shapes

- A node's own fire decision, given its inputs and timing
- A goroutine's own loop behavior (does it poll its own `speedCh`? does a channel
  coalesce to latest?)
- The emitter's own re-published geometry / stream frames
- Bytes read back off disk, ideally through a fresh reload
- Load-time validation (a bad topology is rejected)
- `-race` regressions on a node's own field access

## How industry frames the same problem

- **Actor-model unit testing** (Akka TestKit, Erlang/Elixir). Test one actor by
  sending it messages and inspecting its own replies and state. You do not assert that
  a downstream actor got something — that is the *next* actor's test. This is exactly
  the per-node rule above.
- **Consumer-driven contract testing** (Pact). Where two components must agree, pin
  the CONTRACT at each side independently rather than standing both up and asserting
  end-to-end. Each side's test stays local; the contract is the shared artifact.
- **"Integration tests are a scam" / test-pyramid critique** (J.B. Rainsberger). A
  test that spans components tends to assert a conjunction of many behaviors, so it
  fails for many unrelated reasons and localizes none of them. Our deleted traversal
  tests were textbook: chronically flaky, and when they failed they said nothing about
  where the fault was.
- **Deterministic simulation testing** (FoundationDB, TigerBeetle, Antithesis). When
  you genuinely must exercise the whole system, drive it from a single logical clock
  with controlled scheduling and no wall-clock sleeps, so a failure replays exactly.
  This project already has the precondition — Go owns the one clock — so a real
  whole-system test would go here rather than toward sleep-and-poll.
- **Jepsen.** Verify the guarantees a system *claims*; do not invent stronger ones.
  This is the dividing line above, stated generally: `neighborSetC` claims never-drop,
  so verify it. Bead delivery claims nothing, so there is nothing to verify.
- **The quiescence/barrier problem.** Asserting on an async effect needs a
  happens-before edge, not a sleep. Where a test legitimately reads another
  goroutine's state, it waits on that goroutine's own emitted signal (its breadcrumb)
  — which is also why the teardown fix on this branch is `wg.Wait()` and not a
  `time.Sleep`.

## History

`TestGateFireAndOutputTraversal` and `TestInputToTimeTraversal` both asserted on the
DESTINATION's recv events and were deleted (merged `359a84df`, 2026-07-27), along with
their `live_event_poll_test.go` helper.

Both were also chronically flaky — the fd-teardown race this branch fixes. Flakiness
was the symptom; the wrong shape was the cause. Note that the flake had a blast radius
of eight files, so deleting those two tests removed instances, not the cause.

Two traps worth recording:

- The tests were twice defended as protecting a real invariant — "the gate must stay
  alive stepping its output during traversal." That was false. `gate.go` contains no
  `StepOnce`; the claim came from a stale header comment describing an era when nodes
  stepped their own wires.
- Both files' headers justified themselves by a finished migration ("after X was
  converted to Y", citing `task/non-blocking-update`). The API one of them named,
  `EmitOneDriven`, no longer exists anywhere in the repo. **Distrust a test header
  that explains itself by history; grep for the API it names.**
