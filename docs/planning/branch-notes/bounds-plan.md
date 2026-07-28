---
branch: task/explicit-upper-bounds
---

# Fix plan

Derived from [bounds-inventory.md](bounds-inventory.md) (the sweep) and constrained by
[explicit-upper-bounds.md](explicit-upper-bounds.md) (the rule). Ordered so each step's
decision is settled before anything depends on it.

Status as of 2026-07-28: the inventory is done, the `pending` leak is **confirmed and
fixed** (`93d2e9b6` — gated accumulation on a consumer existing rather than on
`bp.Node != ""`). Everything below is what remains.

## Step 1 — bound + assert `PacedWire.pending`  *(next)*

**The reframing that decides this one.** With a consumer wired, `pending` is drained
*every cycle* by `writeStreamFrame`. So its true maximum is **one cycle's production** —
the beads on this wire that emit, plus that cycle's arrivals. Anything beyond that does
not mean "busy"; it means **the drain stopped running**.

That makes this an **invariant check, not a capacity policy**, which answers the
at-the-bound question the rule insists on deciding per site:

| Option | Verdict |
|---|---|
| **Fail loudly** | **Chosen.** Exceeding the bound is a bug by definition |
| drop-oldest / drop-newest | Wrong — hides a broken drain, the exact failure just fixed |
| backpressure the producer | Wrong — couples the wire's pacing to its owner's frame rate |

The number does not need a magic constant: `wireChanBufferSize` already ceilings how many
placements can be outstanding, so the bound derives from it. The message should name its
own cause — "pending exceeded N; the per-cycle drain is not running" — not just "limit
exceeded".

Per the rule's "Done when", exceed it once deliberately and confirm it fires
(`memory/feedback_check_the_signal_the_check_emits.md`).

## Step 2 — the structural batch

Every site whose maximum is computable from facts already in the repo: the mover inboxes
(`8`), `neighborIn`, the per-edge/per-node triples. Bounds are **derived** (≤ edges = 10,
≤ nodes = 9, ≤ 2×edges = 20 — see the inventory's derivations), not chosen.

Lands as one commit because deriving is mechanical and low-risk, and it retires the
largest un-named group in the repo: **every channel capacity in the network except
`wireChanBufferSize` is currently a bare literal.**

## Step 3 — the dynamic sites, one decision each

These are **not** a blanket policy — the rule says so explicitly, and they genuinely
differ:

- **`inflight`** — head-of-line delivery means a stalled destination blocks everything
  behind it, so the bound and the at-the-bound behaviour are a real design question, not
  a number. Also note its backing array only ever grows (`inflight[1:]` never re-slices).
- **`nm.pending`** — a blocked peer inbox retains that item *and every later item to that
  destination*, FIFO-preserved. Same question, different shape.
- **`breadcrumbCh` (cap 4)** — the tightest queue in the system, dropping **silently**, on
  the diagnostic path that has already bitten twice in one session. Argument to make: this
  deserves a bound *and* a loud signal when it drops. A silently-lost breadcrumb is the
  failure mode we keep paying for, and the whole point of the channel is to be trusted
  when something is wrong.
- **the drain-until-empty loops** — bounded only by the producer; a saturated producer
  makes one "cycle" arbitrarily long. Iteration cap vs. accepting producer-bound.

`task/synctest-deterministic-tests` is now merged (`b2f41e1c`), so a bound-exceeded case
here can be exercised deterministically rather than occasionally — the ordering caveat in
the original note is resolved. **This branch does not make synctest changes**; that is a
constraint on how the work is done here, not a reason to avoid the deterministic tests
that already exist.

## Step 4 — two defects the sweep surfaced

Fixes, not bounds. They are on this branch only because the sweep found them:

- **`splitFrames` has no maximum frame length** (`runCommand.ts:214-228`) while Go
  enforces `maxFrameBytes` on the same protocol (`stdin_reader.go:162`). A corrupt length
  grows `rest` until the payload "arrives". The bound exists on one side of a two-sided
  protocol.
- **`MAX_EDGE_STREAMS` overflow silently sets the count to 0** (`runCommand.ts:458,464`),
  disabling all dedicated streams with no message — quietest signal, loudest consequence,
  and the same path that used to strand the `pending` leak.

## Scope note

**Step 3 is where the cost is** — roughly four separate design conversations, each
needing a decision written down. Steps 1, 2 and 4 are self-contained and could land
without it; step 3 would then be its own branch. Worth choosing deliberately rather than
discovering halfway through.

## Standing constraints

- **No synctest changes on this branch.**
- Every bound lands **with** its assertion — "a limit nobody checks is a comment".
- Every bound gets exceeded once deliberately before it is trusted.
- Breadcrumbs stay unconditional. `tools/check-breadcrumb-not-gated.sh` covers both
  `edgeBeadTraceEnabled` and `StreamsActive`, and catches a **removed** gate as well as a
  widened one. Any new flag near event emission belongs in that guard too.
- No locks, mutexes, or atomics in the network. Startup-set, read-only-after bools are
  the established shape (`edgeBeadTraceEnabled`, `StreamsActive`).
- Every accumulator here is owned by exactly one goroutine, so a per-site counter and
  assertion needs no shared state — and stays inside the testing-shape rule that a test
  asserts what one goroutine itself decided.
