# The column is a channel

## The model

Every buffer column has **exactly one writer and exactly one reader**. A column stops
being a field at an offset inside a stride and becomes a channel: a named, typed entity
carrying one value from the goroutine that owns it to the one thing that draws it.

This is the bridge invariant applied one level down. A wire is a channel between two
nodes; a column is a channel between a goroutine and a renderer. Today the unit is the
row, so a node's 52 columns travel together to every consumer regardless of who wanted
which — the same shape as a node broadcasting its whole state to every neighbour.

## What the rule says about the buffer as it stands

Measured across all 154 columns of all ten blocks, by counting the TS files that call
each generated `read*` helper (excluding the layout files themselves):

| columns | readers | verdict |
| --- | --- | --- |
| 106 | 1 | satisfies the rule today |
| 41 | 1 renderer + `decode-event-*` | needs the observer ruling below |
| 7 | 3 to 9 real consumers | genuine sharing; each is also a drift-rule violation |
| 3 | 2 payloads in one format | `RingPoint`; needs splitting, not deduping |

The rule is already true for two thirds of the buffer. It is not a rewrite; it is a
short list of specific violations plus one decision.

### The decision: is the trace decoder a reader?

`decode-event-line.ts` and its siblings read nearly every column to render probe log
lines. That is the second reader in 41 of the 48 violations.

A logger derives nothing, and nothing downstream depends on what it read. It is an
**observer**, not a consumer, and the rule should say so explicitly. The alternative --
counting it -- makes the rule unsatisfiable without either deleting the trace or
duplicating every column, and a rule that cannot be satisfied stops being checked.

Stating it as an exemption is what keeps it honest: the exemption is named, so a
non-observer hiding behind it is visible. The check must therefore know the observer
list by path, not infer it, or any new file can quietly become a second reader.

### The seven genuinely shared columns

`CX`/`CY`/`CZ` (9 readers), `Radius` (5), `Selected`, `KindId` (3 each).

Every reader of the centre wants to draw something AT the node -- a highlight ring, a
label, interior beads, rule-channel lines, tilt vectors, an edge endpoint check -- and
computes its own geometry from the centre in TS. That is the drift rule's "the render
tree computes no positions", violated once per reader.

So these do not get fixed by copying the column. Each reader gets sent the thing it
draws, positioned by Go, and stops reading the centre at all. The reader count is the
progress bar: when a column reaches one reader, that consumer's drift is gone.

Confirmed instance: `InteriorBeadInstances` computes `cx + readInteriorOX(...)`. Go
sends an offset; TS adds the centre. Go should send the world position.

### RingPoint is two blocks wearing one name

`BuildViewStreamFrame` writes the `RingPoint` block twice -- node ring surface points,
then bead ring surface points -- distinguished only by row range. Its two readers are
not sharing a value; they are reading two different payloads that happen to share a
format. This is the one violation that is a genuine format bug rather than a
consequence of row-shaped thinking, and it is fixable on its own.

## Order

1. **`RingPoint` splits into two blocks.** Self-contained, no model question left open,
   and it removes 3 of the 48 violations outright.
2. **The observer exemption becomes a guard.** An allowlist of observer paths, and a
   check that every other column has one reader. Empty-by-default, loud when broken.
   Do this BEFORE the drift work, so the drift work has a scoreboard.
3. **The seven shared columns, one reader at a time.** Each is: Go computes and sends
   the consumer's own geometry; the consumer stops reading the node centre. Interior
   beads first -- the violation is one line and the destination column already exists.
4. **Only then, addressing.** Columns identified by name rather than by offset within a
   stride. This is what removes `BUF_LAYOUT_FINGERPRINT` and makes adding a column stop
   renumbering its neighbours. It is last because steps 1-3 are worth doing even if this
   is never done, and because it is the only step that touches the generators.

## What breaks

- **Every generated reader and writer**, at step 4 only. Steps 1-3 leave the addressing
  alone.
- **`BUF_LAYOUT_FINGERPRINT`** exists to catch offset drift. With no offsets there is
  nothing for it to catch, and it goes at step 4 -- not before, or the intermediate
  states lose their only layout check.
- **`buildAggregate`** splices rows into a contiguous table. Per-column arrays are
  already contiguous, so it thins to a per-column write at step 4.
- **Frame atomicity.** Today a node frame is one self-describing snapshot of one node at
  one tick. Per-column sends mean a consumer can hold columns from different ticks.
  Within one node's own stream, ordering prevents tearing only if that node's changed
  columns go out in one frame -- so that has to be a stated rule, not an accident.
- **The per-goroutine fd map is NOT affected.** Columns are an encoding, not a stream.
  One fd per column would be 154 x N pipes against a 256 ceiling; the streams stay
  per-goroutine and carry column updates.

## How it is verified

Reader count is countable, so the rule is checkable rather than aspirational: a guard
that parses the generated `read*` helpers, counts the non-observer files calling each,
and fails naming any column with more than one. That guard is step 2, and every
later step moves its number down.

Beyond it: `scripts/stop-checks.sh` empty, and driving the editor -- the seven shared
columns are all things you can see, so a wrong fix shows up as a highlight ring, label,
bead, or tilt vector in the wrong place rather than as a silent number.

Risk worth stating: step 3 moves geometry from TS into Go one consumer at a time, and
each move is a chance to put a position somewhere subtly different from where TS put
it. The ring-matrix change is the precedent -- the two formulations were compared
numerically BEFORE the switch, not after. Do that each time.
