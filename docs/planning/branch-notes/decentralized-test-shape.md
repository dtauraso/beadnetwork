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

These are NOT one idea. Only one of them is "check that the output agrees with a
recorded spec"; the rest are behavioral. Sorted by how much they actually transfer
here.

### Transfers directly

- **Actor-model unit testing** (Akka TestKit, Erlang/Elixir). Behavioral, not
  spec-conformance: send an actor a message and inspect its own replies and state. No
  spec artifact, no expected-output schema. You do not assert that a downstream actor
  got something — that is the *next* actor's test. This is exactly the per-node rule
  above.
- **Deterministic simulation testing** (FoundationDB, TigerBeetle, Antithesis).
  Asserts INVARIANTS under adversarial scheduling and injected faults ("no committed
  write is ever lost"), across many randomized runs. The point is not matching a spec;
  it is that a failure replays byte-identically from a seed. Fits unusually well here,
  because the hard precondition is already met — Go owns the one clock — so a genuine
  whole-system test goes this way rather than toward sleep-and-poll.

### Transfers as a framing

- **Jepsen.** Verify the guarantees a system *claims*; do not invent stronger ones.
  This IS checking against a formal spec, but a spec about legal ORDERINGS of events,
  not about I/O shapes. Stated generally it is the dividing line above: `neighborSetC`
  claims never-drop, so verify it. Bead delivery claims nothing, so there is nothing
  to verify.
- **The quiescence/barrier problem.** Asserting on an async effect needs a
  happens-before edge, not a sleep. Where a test legitimately reads another
  goroutine's state, it waits on that goroutine's own emitted signal (its breadcrumb)
  — which is also why the teardown fix on this branch is `wg.Wait()` and not a
  `time.Sleep`. A sleep is a guess about timing; a barrier is a fact about ordering.

### Does NOT apply here

- **Consumer-driven contract testing** (Pact). This is the one that really is "make
  sure the output agrees with an expected spec": the consumer records "when I send X I
  expect Y," and that recording becomes an artifact the provider is replayed against.
  Neither side ever runs against the other. Its known failure mode follows from that —
  you can end up verifying the recorded expectation rather than the system, and both
  sides pass while the real interaction is broken.

  It exists to solve independently-deployed, separately-versioned services that cannot
  be tested together. Our nodes are one process, one binary, one compilation: the
  "contract" between them is Go types the compiler already checks, and there is no
  version skew to defend against. Nothing to buy. Listed here only so it is not
  mistaken for applicable.

### Diagnoses the tests we deleted

- **"Integration tests are a scam" / test-pyramid critique** (J.B. Rainsberger). A
  test that spans components tends to assert a conjunction of many behaviors, so it
  fails for many unrelated reasons and localizes none of them. Our deleted traversal
  tests were textbook: chronically flaky, and when they failed they said nothing about
  where the fault was.
- **Flaky-as-design-signal.** A test that fails intermittently for reasons unrelated
  to its subject is usually reporting a coupling it should not have had. Treat chronic
  flake as evidence about the test's shape before reaching for a retry or a longer
  timeout. Both tests deleted here were chronically flaky AND wrong-shaped; fixing the
  flake alone would have preserved the wrong test.

## Deciding a new test, in order

1. **What is the subject?** Name the single goroutine whose behavior is on trial.
2. **Does the assertion read only that goroutine's own output?** Its state, its
   emissions, its persisted bytes, its stream frames. If yes, done — write it.
3. **If it reads another goroutine's state, which channel carried the effect?**
   Bead → wrong shape, split into two local tests. Documented never-drop layout
   message → allowed, continue.
4. **How does it wait?** A happens-before edge (the target's own emitted signal, a
   `WaitGroup`), never a sleep.
5. **How does it tear down?** Stop the goroutines and WAIT for them before any
   filesystem cleanup runs. See `test-teardown-waitgroup.md`.
6. **Does the header justify the test by history?** If it says "after X was converted
   to Y," grep for X. It is often already deleted, and the test's stated reason with
   it.

## Anti-patterns, named

- **Asserting delivery.** "A placed a bead, therefore B received it." No such
  guarantee exists. Split it: A's emit decision, and B's behavior given an input.
- **Sleep-as-barrier.** `time.Sleep(150ms)` standing in for a happens-before edge.
  Passes on a fast machine, flakes in the parallel suite.
- **Signal-without-wait teardown.** `defer cancel()` reads as teardown but only
  requests it; goroutines are still running when cleanup starts.
- **Testing the recording.** Asserting against a fixture that encodes the same
  assumption the code does, so both drift together and both stay green.
- **Stale self-justification.** A test explaining its existence by a finished
  migration, naming an API that no longer exists.

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
