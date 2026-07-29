# Testing a decentralized goroutine system

What shape a test may take when every node is an independent goroutine. This is
methodology — the MEDIUM, not the substance — so industry practice applies directly
here, unlike the execution model itself (CLAUDE.md, "Medium vs. substance").

## The rule

**A test asserts what ONE goroutine itself decided, emitted, or persisted.**

Two corollaries, and the second is the sharper one:

1. **Never assert delivery.** A node calls `PlaceDrivenAt` and returns; the `PacedWire`
   goroutine owns the bead and times its own traversal. "Node A fired, therefore the bead
   arrived at node B's input" asserts a guarantee the model deliberately does not have —
   no ack, no nack, no handshake, no send-gating
   (`memory/feedback_node_model_not_networking_handshake.md`). Split it: A's own emit
   decision, and B's own behavior given an input.

2. **Do not test that two or more goroutines communicate properly at all.** Not delivery,
   not ordering, not absence-of-deadlock, not absence-of-race. In this model that
   correctness is guaranteed BY CONSTRUCTION: each mover owns its own state, writes only
   its own fields, and communicates over dedicated per-pair channels with no locks and no
   atomics — enforced by `tools/check-no-network-locks.sh` with an empty allowlist, and by
   the write-once/mutable type split documented on `nodeMover.geom`. A test asserting two
   movers interact correctly asserts what the structure already guarantees; what it
   actually exercises is Go's runtime.

Corollary 2 subsumes cases that look defensible in isolation — a `-race` regression, a
no-deadlock stress test, a "drag X and check neighbor Y re-quantized" cascade test. All of
them take two goroutines as their subject. If ownership is structural, none of them can
fail for a reason that is about this codebase.

## The one exception: persistence

Tests that assert **bytes on disk through a real reload** stay, even though a goroutine
wrote those bytes. `memory/feedback_headless_repro_verifies_persistence` records that
green unit tests hid live persistence failures three times, and that the fix was driving
the real thing and reading on-disk bytes. That hole is real and empirical; do not close it
on theory.

## Legitimate shapes

- A node's own fire decision, given its inputs and timing
- A goroutine's own loop behavior (does it poll its own `speedCh`? does a channel coalesce
  to latest?)
- A mover's own routing decision — call `forwardDelta` on a bare `nodeMover` and assert
  what it *sent*, with no recipient running
- The emitter's own re-published geometry / stream frames
- Bytes read back off disk through a fresh reload (the exception above)
- Load-time validation (a bad topology is rejected)

## The channel-guarantee distinction, and why it is not enough

It is tempting to permit cross-goroutine assertions when the channel *claims* a guarantee:

| Channel | Guarantee |
|---|---|
| Bead delivery over a `PacedWire` | None. No ack, no retry, no back-pressure. |
| The `neighborSetC` layout cascade | Never-dropped — a full channel means the message is RETAINED on the sender's own pending queue and retried next cycle (MODEL.md). |

That distinction is real and worth knowing. But it does **not** license the test: the
never-drop property is itself a property of one mover's own send path
(`sendMove`/`pending`/`flushPending`), testable locally on one goroutine. Asserting it
end-to-end buys nothing and costs a fixed-sleep race window.

Worse, such tests can fake the very wiring they appear to prove. The removed
`delta_forward_test.go` said so outright: `wireNodeStream` "does NOT wire
`nodeRowFor`/`forwardOnce` (SetNodeStreams' job in production, never called by this
bare-LoadTopology test harness) — so wire them here." It started nine goroutines, rebuilt
the wiring by hand, slept 150ms, and re-asserted a rule a bare-`nodeMover` unit test in
the same package already proved exactly.

## Absence assertions cannot be polled

A routing test usually asserts an *absence* — "node 3 forwarded to nobody", "node 6 stayed
at 0". An absence has no convergence signal, so on a wall clock the only option is "sleep
150ms and hope". Such a test passes if the cascade is merely SLOW rather than ABSENT.

This is a second, independent reason to assert routing locally: `forwardDelta` returning
without sending is an exact, instant fact.

## How industry frames the same problem

These are NOT one idea. Only one of them is "check that the output agrees with a recorded
spec"; the rest are behavioral. Sorted by how much they actually transfer here.

### Transfers directly

- **Actor-model unit testing** (Akka TestKit, Erlang/Elixir). Behavioral, not
  spec-conformance: send an actor a message and inspect its own replies and state. No spec
  artifact, no expected-output schema. You do not assert a downstream actor got something —
  that is the *next* actor's test. This is the rule above.
- **Deterministic simulation testing** (FoundationDB, TigerBeetle, Antithesis). Asserts
  INVARIANTS under adversarial scheduling and injected faults, across many randomized runs;
  the point is not matching a spec but that a failure replays byte-identically from a seed.
  Fits unusually well here, because the hard precondition is already met — Go owns the one
  clock. NOW possible for the pieces that only need `time`/`RealClock`: `testing/synctest`
  (Go 1.25) runs a test inside a bubble with a FAKE clock, where `time.Sleep` advances
  time exactly with no scheduler jitter, so `RealClock` reads deterministic time and
  assertions can be equalities instead of wall-clock inequalities with slack (see
  `nodes/wire/clock_realclock_test.go`, `clock_copy_test.go`, `clock_speed_test.go`,
  `nodes/Wiring/pending_bound_test.go`). This does not cover every test: pieces pacing on
  a real goroutine schedule outside a bubble (e.g. a gate loop's own background ticking in
  `nodes/gatecommon/gate_unwired_speed_test.go`) still sleep on real wall time — that test
  waits up to 2s for a window to open and then sleeps 3.8s real time to prove a
  speed-0 gate does NOT advance.

### Transfers as a framing

- **Jepsen.** Verify the guarantees a system *claims*; do not invent stronger ones. This IS
  checking against a formal spec, but a spec about legal ORDERINGS of events, not I/O
  shapes.
- **The quiescence/barrier problem.** Asserting on an async effect needs a happens-before
  edge, not a sleep. A sleep is a guess about timing; a barrier is a fact about ordering.

### Does NOT apply here

- **Consumer-driven contract testing** (Pact). The one that really is "make sure the output
  agrees with an expected spec": the consumer records "when I send X I expect Y", and that
  recording becomes an artifact the provider is replayed against. Neither side ever runs
  against the other. Its known failure mode follows — you can end up verifying the recorded
  expectation rather than the system, and both sides pass while the real interaction is
  broken. It exists for independently-deployed, separately-versioned services. Our nodes
  are one process, one binary: the "contract" is Go types the compiler already checks, and
  there is no version skew. Listed only so it is not mistaken for applicable.

### Diagnoses what was removed

- **"Integration tests are a scam" / test-pyramid critique** (J.B. Rainsberger). A test
  spanning components asserts a conjunction of many behaviors, so it fails for many
  unrelated reasons and localizes none of them.
- **Flaky-as-design-signal.** A test failing intermittently for reasons unrelated to its
  subject is usually reporting a coupling it should not have had. Read chronic flake as
  evidence about the test's shape before reaching for a retry or a longer timeout.

## Deciding a new test, in order

1. **What is the subject?** Name the single goroutine whose behavior is on trial.
2. **Does the assertion read only that goroutine's own output?** Its state, its emissions,
   its persisted bytes. If yes — write it.
3. **If it needs a second goroutine running, stop.** Either the property is guaranteed by
   ownership construction (then there is nothing to test), or it decomposes into two local
   tests (then write those), or it is persistence (then it is the exception).
4. **How does it wait?** A happens-before edge, never a sleep.
5. **How does it tear down?** Stop the goroutines and WAIT for them before any filesystem
   cleanup runs — `MoveDispatch.Start` returns a `*sync.WaitGroup` for exactly this.
6. **Does the header justify the test by history?** If it says "after X was converted to
   Y", grep for X. It is often already deleted, and the test's stated reason with it.

## Anti-patterns, named

- **Asserting delivery.** "A placed a bead, therefore B received it."
- **Two-goroutine subjects.** Race regressions, deadlock stress tests, cross-node cascade
  assertions — the structure already guarantees these.
- **Sleep-as-barrier.** `time.Sleep(150ms)` standing in for a happens-before edge. Passes
  on a fast machine, flakes in the parallel suite, and silently passes an absence assertion
  when the effect is merely slow.
- **Signal-without-wait teardown.** `defer cancel()` reads as teardown but only requests
  it; goroutines are still running when `t.TempDir`'s `RemoveAll` starts, which produces
  intermittent "bad file descriptor".
- **Faking the wiring.** A harness that hand-assigns the hooks production sets up, then
  claims to test integration.
- **Testing the recording.** Asserting against a fixture that encodes the same assumption
  the code does, so both drift together and both stay green.
- **Stale self-justification.** A test explaining its existence by a finished migration,
  naming an API that no longer exists.

## History

Three passes, 2026-07-27/28:

1. `TestGateFireAndOutputTraversal` and `TestInputToTimeTraversal` deleted (merged
   `359a84df`) — both asserted on the DESTINATION's recv events, i.e. bead delivery. Both
   were also chronically flaky; flakiness was the symptom, the wrong shape was the cause.
2. ~25 further tests removed under corollary 2 — all routing/tap drag tests, the cross-node
   layout-cascade tests, and the race/deadlock/goroutine-count tests, plus their orphaned
   helpers. `Start(ctx)` sites went 33 → 5 (the five persistence tests). `nodes/Wiring`
   went from ~40s and the suite's flakiest package to ~2s. 11 of 12 fixed-sleep barriers
   went with them.
3. What survives pins the same rules deterministically: the cascade routing rules are fully
   covered by bare-`nodeMover` unit tests that call `forwardDelta`/`handle` directly.

Two traps worth recording, both of which cost real time:

- The deleted traversal tests were twice defended as protecting a real invariant — "the
  gate must stay alive stepping its output during traversal". That was false: `gate.go`
  contains no `StepOnce`. The claim came from a stale header comment describing an era when
  nodes stepped their own wires.
- Both files justified themselves by a finished migration, citing `EmitOneDriven` — an API
  that no longer exists anywhere in the repo. **Distrust a test header that explains itself
  by history; grep for the API it names.**
