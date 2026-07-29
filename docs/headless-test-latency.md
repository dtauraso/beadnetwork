# Headless test latency — why the Go leg costs ~68s, and the fix

Five headless tests account for ~68 of the ~80 seconds `scripts/stop-checks.sh` spends on
the Go leg. Every Go commit pays it, and so does every `git push` (the pre-push hook runs
`scripts/verify.sh`). Measured with `go test -count=1 -v .`:

```
TestHeadlessViewFdDedicatedStream        20.26s
TestHeadlessFirstFrameHasRealGeometry    20.25s
TestHeadlessNodeRowOrderIsDeterministic  17.42s
TestHeadlessNodeFdDedicatedStream         8.21s
TestHeadlessEdgeFdDedicatedStream         2.45s
                                         ------
                                        ~68.6s   (package total 68.9s)
```

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

## Steps

Step 0 gates the rest: decide per-test whether it should exist before making it fast (see
"Assumption: a test may be removed rather than sped up" below).

0. **Triage the five.** Drop `TestHeadlessNodeRowOrderIsDeterministic`'s runs 2-5, keeping
   its spec-order assertion. Re-measure before touching anything else — that alone is
   expected to be the single largest win, and it is a deletion, not an optimisation.
1. **Verify `SetReadDeadline` works on these pipes** before building on it. It is documented
   for pollable descriptors and `os.Pipe` qualifies, but this is the load-bearing assumption
   of the whole change — a five-line scratch program either shows a timeout error on a quiet
   pipe or it does not. If it does not, fall back to a reader goroutine with a
   `select` on `time.After`, which costs one goroutine per row and no API assumption.
2. **Change `readLastFrames`** to take an idle timeout instead of `maxFrames`, keeping the
   "fatal if not even one frame arrived" behaviour — that assertion is load-bearing (it is
   what catches a stream that streams nothing, per the `task/edges-not-visible` work).
3. **Update the five call sites**, dropping the 120/200/400 counts.
4. **Rename `TestHeadlessFirstFrameHasRealGeometry`** to match what it asserts (settled,
   not first) — the stale name came from the frame count being removed in step 3.
5. **Re-measure.** Record the new per-test times in this file. Expected: the Go leg drops
   from ~80s to roughly 15s, but that is a prediction, not a result — replace it with what
   is actually measured.
6. **Leave `runCtx` at 20s.** It should now never be reached; if a test still takes 20s
   after this, that is a real hang and the deadline has become a genuine signal instead of
   routine flow control.

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
