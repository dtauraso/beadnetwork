# The column is a channel

## The model

Every buffer column has **exactly one writer and exactly one reader**, and is carried on
its own fd. A column is not a field at an offset inside a stride; it is a channel from
the goroutine that owns a value to the one thing that draws it.

The row goes away. A node does not emit "everything about me" -- it emits the values
that changed, each on its own channel, and each channel is last-write-wins exactly as
the row is today. There is no accumulated state anywhere, no diff against what a
consumer has seen, and no message that means "a delta"; a channel simply holds its
latest value.

**The position column carries the polar index**, `polarindex.Index{Phi, Theta, R int}`,
not a derived cartesian triple. The index is the authority: it is exact, it is what the
drag rules, locks and quantization operate on, and it is one value rather than three
that can arrive apart.

## Why the row is the defect, not the unit

The Node row bundles 52 columns whose readers are nearly disjoint. Sixteen matrix floats
have one reader. Ten rule columns have one reader, in a panel. Seven tilt columns have
one reader, in a tab that is not even open most of the time. The only genuinely shared
values are the centre and radius -- and those are shared because eight consumers each
re-derive their own geometry from the node's position in TS, which is the drift rule
being violated once per consumer, not evidence that the bundle is load-bearing.

A complete snapshot reads as coherent at a glance and costs nothing to look at, which is
exactly why the divergence stayed invisible: nothing forces it into view. Counting
readers forces it into view.

## Measured against the rule

All 154 columns of all ten blocks, counting the TS files that call each generated
`read*` helper. The count is by FILE, so it is an upper bound on compliance: two
independent consumers inside one file read as one.

| columns | readers | verdict |
| --- | --- | --- |
| 106 | 1 | satisfies the rule today |
| 41 | 1 renderer + `decode-event-*` | the probe-trace decoder; see below |
| 7 | 3 to 9 | genuine sharing, all of it drift |
| 3 | 2 payloads, 1 format | `RingPoint`; a format defect, fixable alone |

### The trace decoder

`decode-event-*` reads nearly every column to print probe log lines, and is the second
reader in 41 of the 48 violations. It derives nothing and nothing depends on what it
read. Under the rule it is an **observer**, and that has to be written down as a named,
path-allowlisted exemption -- an unnamed one means any new file can quietly become a
second reader, and a rule nobody can satisfy stops being checked.

### The seven shared columns

`CX`/`CY`/`CZ` (9 readers), `Radius` (5), `Selected`, `KindId` (3 each). Every reader of
the centre wants to draw something AT the node -- highlight ring, label, interior beads,
rule-channel lines, tilt vectors, an edge endpoint check -- and computes its own
geometry from the centre. Each gets sent the thing it draws, positioned by Go, and stops
reading the centre. Confirmed instance: `InteriorBeadInstances` computes
`cx + readInteriorOX(...)`.

### RingPoint

`BuildViewStreamFrame` writes the block twice -- node ring surface points, then bead ring
surface points -- distinguished only by row range. Two payloads sharing one format, so
the column's meaning depends on which row you are in. Two blocks.

## The one thing this must not recreate

If TS receives the polar index and converts it to cartesian to place a mesh, there are
two conversions again -- Go's `Polar2cart` for edges, beads and the ring matrix, and
TS's for node placement -- kept in step by nothing. That is precisely the drift removed
in `2b846eb0a` this morning, where the ring, the pick surface and the band were composed
twice from the same angles.

So the index column and the render-form column are different channels with different
readers: the index goes to whatever reasons about position, and Go keeps composing the
matrix the GPU consumes. Same shape as the ring matrix today -- Go composes, TS copies.
The index is what position IS; the matrix is a projection for the GPU.

`SceneConstants` is already scene-level, so nothing per-node carries the step constants.

## Order

1. **`RingPoint` splits into two blocks.** Self-contained, removes 3 violations, no open
   question.
2. **The reader-count guard**, with the observer allowlist. Do it BEFORE the drift work
   so that work has a scoreboard, and so the allowlist is empty-by-default and loud.
3. **The seven shared columns, one reader at a time.** Go computes and sends each
   consumer's own geometry. Interior beads first: the violation is one line and the
   destination column exists.
4. **Position becomes the index.** `Index{Phi, Theta, R}` replaces `CX`/`CY`/`CZ` on the
   channel that reasons about position.
5. **One fd per column, row dissolved.** Last, because 1-4 stand on their own and this is
   the only step touching the generators and the fingerprint.

## What breaks

- **`BUF_LAYOUT_FINGERPRINT`** exists to catch offset drift; with no offsets there is
  nothing to catch. It goes at step 5, not before, or the intermediate states lose their
  only layout check.
- **`buildAggregate`** splices rows into a contiguous table. Per-column channels are
  already contiguous per column; it thins to a per-column write at step 5.
- **Every generated reader and writer**, at step 5 only.

## Verified, not assumed

- **fds are not a constraint.** OS max is 92,160 per process (soft limit 1,048,576 here).
  A spike spawned a child with inherited pipes writing a distinct pattern on each:
  256, 520, 1000 and 2000 pipes all delivered every byte on the correct fd, zero
  mis-routed. 52 columns x 10 nodes is 520. `MAX_NODE_STREAMS = 256` is a constant this
  repo chose, not a limit imposed on it.
- **Ordering is per-channel and sufficient.** One writer per fd means a column's own
  sequence cannot be reordered, so a stale or future value cannot appear. Cross-column
  skew is bounded by one tick and self-corrects, and the index carries position as a
  single value, so the case that would have mattered does not arise.

## How it is checked

Reader count is countable, so the rule is checkable rather than aspirational: parse the
generated `read*` helpers, count non-observer files calling each, fail naming any column
with more than one. That is step 2, and every later step moves its number down.

Beyond it: `scripts/stop-checks.sh` empty, and driving the editor -- the seven shared
columns are all things you can see, so a wrong fix lands as a ring, label, bead or
vector in the wrong place rather than as a silent number.

Risk: step 3 moves geometry from TS into Go one consumer at a time, and each move can
put a position somewhere subtly different from where TS put it. The ring-matrix change
is the precedent -- the two formulations were compared numerically BEFORE the switch.
Do that each time.
