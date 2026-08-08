---
branch: task/arith-math-into-code
---

# The update rules in the code are the update rules on the page

`nodes/Node1/machine.go` and arith.html's **"Which end moves, and whether it goes +1, −1 or
stays"** decide the same thing by different arithmetic. The code is to compute the page's way.
That is the whole plan.

Delete this file when the change lands.

## The two forms

```
code   fromRest = min over r in restingLengths[mode] of |angleLength(t, a) − r|
       settled  = fromRest == 0
       step     = fromRest(t+1) <= fromRest(t−1) ? t+1 : t−1

page   c        = the count from the nearer end of the tilt line
       settled  = c == the count this mode stops at
       step     = the measured end moves one slot toward that count
```

The mode is data in both, and stays data: a list of resting lengths becomes one stopping count.
Nothing branches on which mode is running, before or after.

The page prints the rule as two four-line blocks — parallel stopping at a quarter turn,
perpendicular at 0 — and they are one rule: `c` lives on a ring of `halfTurn`, and the node walks
it the short way toward its own stop, a tie going up.

## Both ends are stored

`c` is measured at the nearer end, so `c ± 1` moves THAT end. Today `Top` is the only stored tilt,
so an update measured at the bottom would have to be negated to write it to the top — putting back
by hand the reversal the page says is not there. Storing both ends is what lets the rule be written
as the page states it:

```
nearer end is the top      →  Top    = Top ± 1        then Bottom = Top.opposite
nearer end is the bottom   →  Bottom = Bottom ± 1     then Top    = Bottom.opposite
```

The ends are a half turn apart and each is written from the other's `opposite` in the same step, so
they cannot disagree. `position.json` still keeps one angle. `Node.Top`'s "The ONE writer, full
stop" comment says why that is still true of the line, with two writers of it.

## Order

0. The sweep runs both lattices. `TestOneRoundIsSignAndRemainder` is `const points = 24` with bare
   6s and 12s on 71 lines and `halfTurn`/`quarterTurn` on none, so it cannot see a confusion
   between `halfTurn` and `points` — which is the one the new form can make. No production code.
1. `machine.go`: the page's arithmetic added beside the existing rule. Nothing calls it.
2. The sweep asserts the two agree on every pair, both lattices.
3. `settled` and `step` switch over. `restingLengths` becomes one stopping count per mode;
   `resting` and `fromRest` go.
4. Both ends stored, the update driving the one it measured.
5. The nine `data-src` citations: `restingLengths` ×7, `resting` ×2. `check-docs-symbols.sh` names
   each.
6. `updates.html` — list→count. Its argument survives; the mode is still data the rule cannot
   interrogate.
7. `MODEL.md` and `.claude/rules/node-kinds.md` — grep `restingLengths` and "a mode contributes".

## Open before step 3

**Setting rests everywhere**, which the list said as "every angle length". One stopping count
cannot say it, so `TiltMachineNone` needs its own spelling — settled wherever it stands, without
reaching a comparison — and `machineFor` must still panic on a mode it does not know.

**The tie.** `step` prefers up. The page's `≥` carries it; a `>` moves 96 of the 1152 pairs at 24.

**`c` is `(t − a)`, not `(a − t)`.** The two agree at the stops and differ in `step`, so the sweep
compares the state chosen, not the distance.

## Verification

`bash scripts/stop-checks.sh`, clean output, at every step. Make one check fail on purpose before
believing it. The sweep stays in the suite afterwards — it is then the code checked against the
page's own statement of it.
