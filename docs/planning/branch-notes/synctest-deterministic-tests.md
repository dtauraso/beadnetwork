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

## Why this branch is worth doing first

The other two open branches (`task/runtime-invariant-assertions`,
`task/explicit-upper-bounds`) both add checks to the node/buffer hot path. Verifying
those under a nondeterministic scheduler means "I ran it and it passed" rather than
"it cannot fail." Landing synctest first makes both of them actually verifiable, and
avoids re-validating their assertions later under a new scheduler.

## Prerequisite: bump `go.mod`

`go.mod` currently declares `go 1.23.0`. `testing/synctest` is not available at that
language version. The locally installed toolchain is `go1.26.3` and **does** have it
(`go doc testing/synctest` resolves). Bump `go.mod` to the minimum version that
provides it, not blindly to the toolchain version.

Expect a trivial conflict on this line when the other two branches merge — they don't
need the bump. Resolve in favor of the higher version.

## What to convert, in order

Start with the tests covering invariants that are *already known to be timing-sensitive*
— these are where the payoff is, and the memory files name them:

- `memory/feedback_paced_tryrecv_blocks.md` — paced `TryRecv` blocking behavior
- `memory/feedback_per_emit_simtime_anchoring.md` — per-emit simtime anchoring
- `memory/project_two_goroutine_node_split.md` — the two-goroutine-per-node split
- `memory/feedback_uniform_pulse_speed.md` — uniform pulse speed

The candidate test files at the repo root:

| File | Why it's a candidate |
|---|---|
| `headless_stream_helpers_test.go` | shared helpers — convert first, everything else depends on the shape chosen here |
| `headless_node_fd_test.go` | per-node stream ownership under concurrent emit |
| `headless_edge_fd_test.go` | edge stream, the highest-volume path |
| `headless_view_fd_test.go` | VIEW-stream fallback for non-per-node emits |
| `headless_first_frame_geometry_test.go` | first-frame ordering — ordering is exactly what a real scheduler randomizes |
| `headless_node_row_order_test.go` | row order, same reason |
| `kind_registry_parity_test.go` | probably pure//static — check before touching; may need nothing |

Convert the helpers first and let the rest follow the pattern. Do not convert all seven
mechanically before checking that the first one genuinely removes nondeterminism.

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
