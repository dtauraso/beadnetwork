# Headless test latency — why the Go leg cost ~68s, and the fix

**Status: fixed.** Original cost, measured with `go test -count=1 -v .`:

```
TestHeadlessViewFdDedicatedStream        20.26s
TestHeadlessFirstFrameHasRealGeometry    20.25s
TestHeadlessNodeRowOrderIsDeterministic  17.42s
TestHeadlessNodeFdDedicatedStream         8.21s
TestHeadlessEdgeFdDedicatedStream         2.45s
                                         ------
                                        ~68.6s   (package total 68.9s)
```

After the fix (same command, same five tests, one renamed — see "Step 0" and "Step 4"
below):

```
TestHeadlessNodeFdDedicatedStream           4.79s
TestHeadlessSettledFramesHaveRealGeometry   5.02s
TestHeadlessNodeRowOrderMatchesSpecOrder    2.50s
TestHeadlessEdgeFdDedicatedStream           2.76s
TestHeadlessViewFdDedicatedStream           0.51s
                                            -----
                                           ~15.6s   (package total 15.7-15.8s across 3 runs)
```

~69s → ~15.7s. That matches this doc's original ~15s prediction, but the mechanism that
got there is NOT the one predicted — see "The idle-timeout premise was wrong for three of
the five streams" below, which the original plan did not anticipate.

Every other package in the repo together is under 25s, and most are under 2s.

## The cause is not the 20s deadline

The first read of this was "they run until `runCtx`'s 20-second `CommandContext` deadline
kills them" (`headless_stream_helpers_test.go:188`). That is the visible symptom for two of
the five, but it is not the mechanism, and fixing the deadline alone would fix nothing.

The mechanism is `readLastFrames` (`headless_stream_helpers_test.go:294`). It takes a
`maxFrames` count and does a **blocking** `readOneRawFrame` that many times, keeping the
last one:

```go
for i := 0; i < maxFrames; i++ {
    buf, err := readOneRawFrame(r)
    if err != nil { if i == 0 { t.Fatalf(...) }; break }
    last = buf
}
```

There is no early exit on "the stream has told me everything it is going to." The loop ends
one of two ways: it counts to `maxFrames`, or a read errors — and the only thing that makes
a read error is the child process dying, which is the 20s deadline. So the cost of each test
is `maxFrames ÷ that stream's emit rate`, capped at 20s.

That predicts the measurements, and it does:

| test | stream | maxFrames | time |
|---|---|---|---|
| `TestHeadlessEdgeFdDedicatedStream` | edge | 120 | 2.45s |
| `TestHeadlessNodeFdDedicatedStream` | node | 120 | 8.21s |
| `TestHeadlessNodeRowOrderIsDeterministic` | node | 200 | 17.42s |
| `TestHeadlessFirstFrameHasRealGeometry` | node + edge | 400 | 20.25s (capped) |
| `TestHeadlessViewFdDedicatedStream` | view | 200 | 20.26s (capped) |

The view case is the worst kind of capped: the **VIEW stream is event-driven**, unlike
NODE/EDGE/INTERIOR (`distance_groups.go:141` and `view_stream.go`'s header). A headless run
with no pointer input emits a handful of view frames and then goes quiet *forever*. That
test can never reach 200 frames, so it is structurally guaranteed to sit until the deadline.
Raising or lowering the deadline just changes how long it sits.

## What the tests actually need

Each of these asserts on the **settled** state — the last complete frame per row, not the
first (possibly still-degenerate) one. `readLastFrames`' own doc comment says so. The frame
*count* was never the thing under test; it is a stand-in for "wait long enough that what I
read is settled," and its comment's justification —

> bounded, non-blocking-in-effect since production keeps emitting on a change/tick cadence

— is false for the event-driven VIEW stream, which is exactly the slowest case.

## The fix: read until quiet, not until N

Replace the frame count with an **idle timeout**. Drain frames until the stream produces
nothing for a short interval, then return the last complete one:

- `os.File` supports `SetReadDeadline` on pipes (pollable fds), so each `readOneRawFrame`
  can carry a deadline of ~250ms.
- On `os.ErrDeadlineExceeded`, stop and return `last` — that is "the stream has gone quiet,"
  which is the settled condition every one of these tests actually wants.
- Any other error keeps today's behaviour (fatal if it happened on frame 0, else break).

Cost becomes *time-to-settle* rather than `maxFrames ÷ rate`. The event-driven VIEW case
stops being pathological — it is the FASTEST case under this rule, because it goes quiet
almost immediately.

This also lets the 20s `runCtx` deadline go back to being what it should be: a backstop for
a genuinely hung child, never the thing that ends a passing test.

## The idle-timeout premise was wrong for three of the five streams — what actually shipped

The plan above (per-frame idle timeout: reset the read deadline after every frame, stop on
the first timeout) was implemented and measured exactly as written. Result: VIEW dropped to
~0.7s, but NODE, EDGE, and NODE-ROW-ORDER (which reads NODE) all landed back at the 20.2xs
cap — no better than the `maxFrames` version, and EDGE regressed from 2.45s to 20.26s.

The cause: NODE/EDGE/INTERIOR do not go idle, ever, in a headless run. Probing the real
binary (reading raw frames and diffing them) showed `nodeMover.run` calls
`writeStreamFrame` **every clock cycle unconditionally** — not on change — at roughly a
17ms cadence, forever, and `edgeMover` does the same plus every edge frame carries its
wire's live in-flight bead positions, which change every single frame for as long as a bead
is on the wire. This repo's topology has a self-feeding ring (`memory/feedback_edge_seed_
required_for_rings.md`), so in a headless run with no pointer input that is *forever*. A
per-frame idle timeout on a stream that never goes idle degenerates back to exactly the
problem being fixed: wait for the child to die at the 20s `runCtx` deadline.

What the probe also showed: NODE payload (frame bytes minus the tick prefix) is
byte-**identical** from the second frame onward — the structural state (ports, geometry,
ids) these tests actually assert on settles almost immediately; only the tick counter and,
on EDGE, the moving bead position keep changing after that, and none of these tests read
bead position.

**The shipped fix is a wall-clock BUDGET per row, not idle-since-last-frame**, and that
budget only STARTS after a row's first frame arrives. `readLastFrames` reads each row's
first frame with NO deadline at all, then sets one `SetReadDeadline` at `now + settleWindow`
(250ms, not renewed per read) and keeps whatever frame arrived last when that expires.

The "no deadline on the first frame" part was itself a second correction, found the same
way as the first: measured, not assumed. A version that put the 250ms deadline around the
first frame too passed standalone but failed under `scripts/stop-checks.sh`, which runs
every package's `go build`+`go test` concurrently — under that load, the time from the
child process's `Start()` to its first write occasionally exceeded 250ms for reasons that
have nothing to do with the stream (scheduling/compiler contention), and the "fatal if not
even one frame arrived" guard (correctly) fired. Un-deadlining the first read fixes that:
it relies on the same backstop the old `maxFrames` version always relied on for a genuinely
silent stream — the caller's 20s `runCtx` — and only starts the fixed 250ms budget once a
row has proven it is actually producing.

This still uses the same `SetReadDeadline`/`os.ErrDeadlineExceeded` mechanism proven in
Step 1, still fatals if the first frame never arrives (now via the undeadlined read plus
`runCtx`, rather than a budgeted read), and still bounds every row that DID start to a
fixed, small post-startup cost regardless of whether that stream's production goroutine
ever actually falls silent. VIEW (which may go genuinely quiet after its first frame) and
NODE/EDGE/INTERIOR (which never do) both cost about the same now: startup time plus one
`settleWindow` per row, not `maxFrames ÷ rate` and not "however long until the child is
killed."

This is the "the idle interval is a knob... if it proves flaky, find a definite
end-of-settling signal" case named in this doc's own Risk section below — the failure
showed up as a flake under concurrent-package load (`stop-checks.sh`) rather than in
isolation, and was fixed before landing, not after; `settleWindow` remains a knob worth
staying suspicious of if it flakes again.

## Steps (done — this section is now a record, not a plan)

Step 0 gated the rest: decide per-test whether it should exist before making it fast (see
"Assumption: a test may be removed rather than sped up" below).

0. **Triage the five.** Dropped `TestHeadlessNodeRowOrderIsDeterministic`'s runs 2-5,
   keeping its spec-order assertion (renamed to `TestHeadlessNodeRowOrderMatchesSpecOrder`
   since it no longer tests determinism across runs). Measured before touching anything
   else: 17.42s → 3.86s standalone.
1. **Verified `SetReadDeadline` works on these pipes** with a throwaway scratch program
   (`os.Pipe`, `SetReadDeadline(+250ms)`, blocking `Read` on the quiet end): it returned
   after ~251ms with `errors.Is(err, os.ErrDeadlineExceeded) == true` (note: the raw error
   value is `"i/o timeout"`, not `== os.ErrDeadlineExceeded` by identity — `errors.Is` is
   required). The documented API path was used; the goroutine+`select` fallback was not
   needed.
2. **Changed `readLastFrames`** — but not to a per-frame idle timeout as originally
   written, and not in one pass. See "The idle-timeout premise was wrong for three of the
   five streams" above for the first correction (per-frame idle-reset → per-row
   `settleWindow` budget: NODE/EDGE never go idle, so idle-since-last-frame degenerated
   back to waiting for the child to die). A second correction followed once run under
   `scripts/stop-checks.sh`'s concurrent-package load: budgeting the FIRST frame too made
   the test flake there (child startup, not stream behaviour, occasionally ate the whole
   250ms). Shipped: the first frame per row is read with no deadline at all (same backstop
   the old `maxFrames` version always had — the caller's 20s `runCtx`); `settleWindow` only
   starts once a row has proven it is producing.
3. **Updated the five call sites**, dropping the 120/200/400 counts.
4. **Renamed `TestHeadlessFirstFrameHasRealGeometry`** → `TestHeadlessSettledFramesHaveRealGeometry`
   (file: `headless_first_frame_geometry_test.go` → `headless_settled_geometry_test.go`) to
   match what it asserts (settled, not first).
5. **Re-measured**, standalone runs, consistent within run-to-run system noise:
   ```
   TestHeadlessNodeFdDedicatedStream           4.8-5.0s
   TestHeadlessSettledFramesHaveRealGeometry   5.0-5.2s
   TestHeadlessNodeRowOrderMatchesSpecOrder    2.5-2.7s
   TestHeadlessEdgeFdDedicatedStream           2.7-2.9s
   TestHeadlessViewFdDedicatedStream           0.5-0.7s
                                               ------
   package total                              ~15.7-16.6s
   ```
   Also verified clean (not just green) under the actual `scripts/stop-checks.sh` full run,
   where every package's tests execute concurrently — this is what caught the second
   correction above; a standalone `go test -count=1 -v .` pass alone did not.
   The ~80s→~15s prediction held, but only after correcting the mechanism twice (see
   above) —
   the first honest measurement of the plan AS WRITTEN was a near-total miss (only VIEW
   improved).
6. **Left `runCtx` at 20s.** No test in this package reaches it anymore; if one does in the
   future, that is a real hang, not routine flow control.

## Assumption: a test may be removed rather than sped up

The steps above assume all five tests survive and only get faster. That assumption should
be tested per-test BEFORE optimising, because making a test that should not exist run
faster is the more expensive mistake — it locks the test in by making it cheap.

The criterion is `docs/testing-shape.md`'s: a test asserts what **one goroutine itself**
decided, emitted, or persisted. A test whose cost buys a property the structure already
guarantees is not slow, it is unnecessary.

**`TestHeadlessNodeRowOrderIsDeterministic` is the clearest candidate.** It spawns the whole
binary **5 times** (`const runs = 5`, `headless_node_row_order_test.go:48`) — that is where
its 17.42s comes from, not the frame count alone. It asserts two different things:

1. row order equals spec order — a real assertion about what the loader emitted, and it
   needs exactly **one** run;
2. row order is IDENTICAL across the 5 runs — a determinism claim about a
   directory-sorted id list.

The second is the anti-pattern doctrine names: node row order is directory-sorted
(`.claude/rules/persistence-ownership.md`, "Node row order is directory-sorted"), and a sort
is deterministic by construction. Re-running the binary 4 more times to observe that
`sort` sorted the same way exercises Go's runtime, not this codebase. Dropping runs 2-5
keeps assertion (1) intact and removes ~14s — more than the idle-timeout change would win
on this test, and it removes the reason to tune it at all.

**The three per-owner fd tests (view / node / edge) are NOT redundant with each other** even
though their headers read almost identically. Each proves a different stream kind reaches
its own dedicated fd, which is the invariant in
`memory/feedback_no_single_writer_bridge.md`, and no other test drives real fds. Keep all
three.

**`TestHeadlessFirstFrameHasRealGeometry` has a stale name, not a redundant assertion.** It
is called "first frame" but its own doc says SETTLED, and it reads 400 frames to take the
last. The assertion (no degenerate segments, ports seeded) is real and uniquely covered;
the name should be corrected to match, and that rename belongs in the same commit as the
`maxFrames` removal, since the count is what made the name wrong.

So the honest expected outcome is a mix: one test loses four spawns, one gets renamed, and
all five lose their frame counts. Re-measure after each, not once at the end — otherwise a
removal and a speedup get credited to whichever landed last
(`memory/feedback_ease_of_fix_is_confounded.md`).

## What this does not change

No test's assertions change. This is purely how long the harness waits before reading what
it already asserts on. If any assertion has been silently depending on "400 frames' worth of
simulation has elapsed" rather than on settling, this will surface it as a failure — which
is information worth having, not a reason to keep the counts.

## Risk worth naming

"Quiet for 250ms" is a timing heuristic, and this repo's testing doctrine
(`docs/testing-shape.md`) is hostile to tests that assert across goroutines by waiting. This
change does not add such an assertion — the tests already wait; it makes the wait
proportional to the thing being waited for instead of to an unrelated constant. But the
idle interval is a knob, and per `memory/feedback_go_vs_coordinator_bias.md` a knob is worth
being suspicious of. If 250ms proves flaky under load, the answer is not to raise it: it is
that the harness should read a definite end-of-settling signal, and finding that signal is
the better fix.
