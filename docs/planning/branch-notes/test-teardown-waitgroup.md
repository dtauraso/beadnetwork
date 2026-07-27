---
branch: task/test-teardown-waitgroup
---

# Tests must wait for goroutines before TempDir cleanup

## The bug

`MoveDispatch.Start` (`nodes/Wiring/node_move.go:327`) returns a `*sync.WaitGroup`
covering every launched goroutine. Every test discards it.

The resulting race:

1. `dir := t.TempDir()` registers a `RemoveAll` via `t.Cleanup`.
2. `ctx, cancel := context.WithCancel(...)`, `defer cancel()`.
3. `md.Start(ctx)` / `go n.Update(ctx)` launch goroutines that write persist files and
   hold stream fds under that dir.
4. The test returns. The deferred `cancel()` fires — but `cancel()` only SIGNALS; the
   goroutines have not stopped.
5. `t.Cleanup` then runs `RemoveAll` concurrently with those goroutines still closing
   fds → `unlinkat: bad file descriptor`, intermittently, only in the parallel suite.

`defer cancel()` looks like teardown but is only a request for it. Nothing waits.

## The fix

Exploit `t.Cleanup`'s LIFO order — registered AFTER `t.TempDir()`, so it runs BEFORE
the `RemoveAll`:

```go
dir := t.TempDir()
wg := md.Start(ctx)
t.Cleanup(func() { cancel(); wg.Wait() })
```

Eight files: `rotating_pole`, `pre_split_round_trip`, `drag_persist_e2e`,
`per_edge_travel_time`, `drag_anchor`, `abc_drag_count_target_node`, `node_move`,
`time_node_abc_drag_breadcrumb`.

**Assertions stay unchanged. This is teardown hygiene only.**

## The guard

`tools/check-start-waitgroup-used.sh`: flag any `_test.go` calling `.Start(ctx)` as a
bare statement. This is a syntactic property, not a heuristic "does it wait somewhere"
— no allowlist needed. Wire into `scripts/stop-checks.sh`.

Fix order is code-first (CLAUDE.md drift checklist #6): a guard, not more prose. The
earlier attempt at this was a session-log entry on a branch, which is prose and was
lost when the branch went.

## Why these 8 tests stay

See `decentralized-test-shape.md` in this directory. Short version: none of them
assert bead delivery. The two tests that did were deleted (merged `359a84df`).

## Loose thread

`drag_anchor_test.go` reads deltas keyed by `"dst"`. Confirm whether the SENDER or the
RECIPIENT logs that breadcrumb. Does not change the classification.
