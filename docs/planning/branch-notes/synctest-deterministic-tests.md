---
branch: task/synctest-deterministic-tests
---

# Make the headless tests deterministic with `testing/synctest`

## The goal

Concurrency bugs in the node network currently reproduce *sometimes*. A test that fails
one run in fifty costs a whole session to chase and can't gate anything — it has to be
re-run to be believed, which is the opposite of what a guard is for. `testing/synctest`
runs a test's goroutines in an isolated "bubble" with a **fake clock** and a scheduler
that only advances when every goroutine in the bubble is blocked. Timing-dependent
flakes become deterministic failures: same input, same result, every run.

This is the TigerBeetle VOPR idea (deterministic simulation, replay a failure exactly)
at the granularity Go gives us for free.

## Doctrine constraint — read this before converting anything

`CLAUDE.md` § "Testing shape" and `docs/testing-shape.md` prohibit testing that two or
more goroutines communicate properly — not delivery, not ordering, not
absence-of-deadlock, not absence-of-race. That correctness is guaranteed **by
construction** (per-mover ownership, dedicated per-pair channels, no locks/atomics).
`task/remove-multigoroutine-tests` deleted tests of exactly that shape.

synctest's *headline* use case is the prohibited one. **This branch is not that.** The
value being claimed here is narrower and compatible:

- **The fake clock.** A test of what *one* goroutine decided, emitted, or persisted can
  still be timing-sensitive — simtime anchoring, pulse speed, pacing decisions. Those
  currently depend on a real clock. The bubble's fake clock makes them exact without
  asserting anything about inter-goroutine communication.
- **Scheduler determinism as a stability property, not a subject.** Where a
  single-owner assertion happens to run while other goroutines exist, the bubble stops
  the *scheduler* from being the variable. The assertion stays "what this one goroutine
  did."

What this branch must **not** produce: a test whose subject is delivery, ordering, or
deadlock-freedom between movers. If a conversion starts to need that, the answer is to
delete the test, not to stabilize it. Read `docs/testing-shape.md` (decision procedure
and named anti-patterns) before adding or converting anything.

## Why this branch is worth doing first

The other two open branches (`task/runtime-invariant-assertions`,
`task/explicit-upper-bounds`) both add checks to the node/buffer hot path. Verifying
those under a nondeterministic scheduler means "I ran it and it passed" rather than
"it cannot fail." Landing synctest first makes both of them actually verifiable, and
avoids re-validating their assertions later under a new scheduler.

## Prerequisite: bump `go.mod`

`go.mod` currently declares `go 1.23.0`. `testing/synctest` is not available at that
language version. The locally installed toolchain is `go1.26.3` and **does** have it
(`go doc testing/synctest` resolves).

The minimum is **`go 1.25`**, determined empirically rather than assumed — throwaway
modules declaring `go 1.23` and `go 1.24` both fail to vet with
`synctest.Test requires go1.25 or later (module is go1.2x)`, and `go 1.25` passes. Bump
to 1.25, not to the toolchain's 1.26.

Expect a trivial conflict on this line when the other two branches merge — they don't
need the bump. Resolve in favor of the higher version.

## What to convert, in order

Target only invariants that are **timing-sensitive AND single-owner**. Filter every
candidate through the doctrine constraint above before touching it:

| Memory entry | Subject | Verdict |
|---|---|---|
| `feedback_per_emit_simtime_anchoring.md` | what one goroutine anchored its emit to | **good fit** — clock-sensitive, single owner |
| `feedback_uniform_pulse_speed.md` | speed value used by one mover | **good fit** — clock-sensitive, single owner |
| `feedback_paced_tryrecv_blocks.md` | pacing decision vs. channel handoff | **check carefully** — fine if the assertion is what the paced goroutine *decided*; prohibited once it becomes "the send arrived" |
| `project_two_goroutine_node_split.md` | two goroutines per node coordinating | **likely prohibited** — if the subject is the two halves communicating, delete rather than stabilize |

### The root `headless_*` tests are OUT of scope — do not try

An earlier draft of this note listed the seven root-level `headless_*_test.go` files as
the conversion targets. **That was wrong, and the reason is structural, not a matter of
taste.** Recorded here so it is not rediscovered:

Every one of those tests drives the *built binary as a child process*.
`headless_stream_helpers_test.go` runs `go build -o binPath .` (line ~123) and then
`exec.CommandContext(runCtx, binPath, "-topology", …)` (line ~146); the other six go
through those helpers. Across all seven files there are **zero `time.Sleep` calls, zero
timers, and zero goroutines started in-test** — the only `time.` references are the 60s
build and 20s run timeouts on the subprocess.

A synctest bubble cannot cross a process boundary. Whatever nondeterminism these tests
carry lives in the child process, where the bubble's fake clock and scheduler have no
reach. This is the "avoid the network and external processes" constraint below, applied
to this repo's actual test shape. Converting them is not merely low-value; it is not
possible.

They are also the headless-repro pattern
(`memory/feedback_headless_repro_verifies_persistence.md`) — real binary, real bytes on
disk. That is the point of them, and it is worth keeping exactly as is.

### The real candidates: real-clock waiting AND single-owner

25 test files under `nodes/` do genuine real-clock waiting (`time.Sleep`, or
`for time.Now().Before(deadline)` polling). That is the actual flake surface. But most of
them start 2–3 goroutines in-test, which puts their subject in the territory
`docs/testing-shape.md` prohibits — for those the question is deletion, not stabilization
(see "Out of scope" below).

The intersection — clock-sensitive **and** zero goroutines started in-test — is the
conversion set, in order:

| File | Why it's a candidate | What the fake clock buys |
|---|---|---|
| `nodes/wire/clock_realclock_test.go` | one test, sleeps 2 tick periods, no goroutines | `b > a` becomes exactly `b == a+2` |
| `nodes/wire/clock_copy_test.go` | 4 sleeps, one explicitly "let some real elapsed accumulate" | `Tick() >= 2` becomes exactly `== 2` |
| `nodes/wire/clock_speed_test.go` | speed scaling measured across sleeps | kills the `1.5×` slack and the `after+5` slack — the ratio becomes exactly 2× |
| `nodes/wire/paced_wire_rebase_tolerance_test.go` | first test sleeps 10 tick periods | fixes the elapsed span; the assertion is already exact |

`RealClock` is a pure function of `time.Now()`/`time.Since` (see `nodes/wire/clock.go`),
and `SleepCycle` is `time.After` — all of it is fake-clocked inside a bubble, so no
production change is needed to make these deterministic.

Convert `clock_realclock_test.go` first: smallest, purest, no goroutines. Confirm it
genuinely removes a timing tolerance before propagating the pattern to the other three.

### Out of scope for this branch

`nodes/wire/paced_wire_concurrency_race_test.go` (3 in-test goroutines) is named for
exactly the shape `task/remove-multigoroutine-tests` deleted, and the ~20
`nodes/*/firing_rule_lean_test.go` files sit at 2 goroutines each. Triaging those against
`docs/testing-shape.md` is its own branch — stabilizing a test whose subject is
prohibited would entrench it rather than remove it.

## Constraints from the synctest docs

Within a bubble:

- the `time` package uses a **fake clock**, starting at midnight UTC 2000-01-01
- avoid interacting with goroutines not started inside the bubble
- avoid the network and external processes — use a fake where needed
- avoid leaking goroutines in background tasks

Two of those bite here specifically: any test that reaches a real wall clock for simtime
anchoring will need to move to the bubble clock, and any goroutine started outside the
bubble (a long-lived stream writer set up by a helper) has to move inside it or be
faked.

## Done when

- The converted tests pass repeatedly with no timing tolerance / retry / sleep-based
  waiting left in them.
- At least one previously-flaky behavior fails *deterministically* when deliberately
  broken — per `memory/feedback_check_the_signal_the_check_emits.md`, make the check
  fail once before believing it.
- `bash scripts/verify.sh` is clean (exit 0; nonzero means something failed, reason on
  stderr).
