# Run the arithmetic page's maths in Node1

The page (`docs/pair-node/arith.html`) and the code (`nodes/Node1/machine.go`) agree on every
decision but share no arithmetic. `TestOneRoundIsSignAndRemainder` is what holds them together:
it sweeps both arrangements and all 24×24 (tilt, arrival) pairs asserting the page's forms pick
exactly what `settled` and `step` pick. The change is to make the code compute the page's way, so
the agreement is structural rather than proven after the fact.

Delete this file when the change lands.

## What the code does now

```
fromRest = min over r in restingLengths[mode] of |angleLength(t, a) − r|
settled  = fromRest == 0
step     = fromRest(t+1) <= fromRest(t−1) ? t+1 : t−1
```

Fold to an angle length, take a minimum over a per-mode list, evaluate it at both adjacent tilts,
compare. The mode enters as DATA — a row in `restingLengths` — and no function branches on it.

## What the page says

```
u = t − a (mod 24)          one count, from the top
u < 12  → nearer end is the top,    its angle is u
u ≥ 12  → nearer end is the bottom, its angle is v = u − 12

parallel        angle = 6 : stay     angle < 6 : t+1     angle > 6 : t−1
perpendicular   angle = 0 : stay     angle ≥ 6 : t+1     angle < 6 : t−1
```

One count, one bit, comparisons against the quarter. No list, no minimum, no second evaluation at
the neighbours.

## What it costs — read before starting

**It reverses the property `machine.go` was built for.** Its header records three earlier attempts
that failed by parameterising the rule, and states the fix: "Nothing here branches on which mode is
running: fromRest is a minimum over a set of numbers and cannot ask who it is working for, step
never sees a mode at all." The page's form branches per arrangement. That header must be rewritten
in the same commit — code and comment disagreeing is worse than either choice.

Whether that trade is worth making is the open question this plan does not settle. The failure the
split defended against was a change to one mode landing in the other; a `switch mode` in three
functions is exactly the shape that allowed it. Against that: the branches are four lines each and
the sweep would catch a cross-contaminating edit immediately.

**Nine `data-src` citations point at symbols that would go** — `restingLengths` (×7) and `resting`
(×2), across `dots.html`, `perpendicular.html`, `updates.html`, `formulas.html`. `check-docs-symbols`
fails on a citation whose target no longer exists.

**`updates.html` argues the current design at length**: "the mode is in R, as a list of numbers,
which is why fromRest cannot ask which mode it is computing for", plus the whole "Where each mode
rests" card. That page would be describing a machine that no longer exists.

## Order

1. `machine.go`: add the page's arithmetic beside the existing rule, not replacing it — a
   `stopsAt(r)` and a count/bit helper. Nothing calls them yet.
2. Extend the sweep to assert the new functions equal `settled`/`step`/`fromRest` on every pair.
   This is the safety net for step 3 and must be green before it.
3. Switch `settled`, `step` and `fromRest` to the new arithmetic. Delete `restingLengths` and
   `resting`. Rewrite the header comment to say what the rule now is and why the mode branch is
   acceptable here.
4. Repoint the nine citations. `bash tools/check-docs-symbols.sh` names each failure.
5. Rewrite `updates.html`'s resting-length card and the two keynote lines that argue for the list.
6. `MODEL.md` and `.claude/rules/node-kinds.md` — grep for `restingLengths` and for "a mode
   contributes"; fix what is now false.

## Verification

- `bash scripts/stop-checks.sh`, clean output, at every step.
- The equivalence sweep stays in the suite after step 3 — it then compares the code against the
  page's own statement of it, which is the thing worth guarding.
- Make one check fail on purpose before believing it (flip an inequality, confirm the sweep names
  the pair).
- `machineFor` currently rejects an unknown mode via the `restingLengths` lookup. Its replacement
  must still panic on an unknown mode rather than defaulting to an arrangement.

## Risks

- **Ring size.** The page is written for 24 slots; the code runs 24 and 48. Every 6 and 12 becomes
  `r.quarterTurn` and `r.halfTurn`. The sweep runs both lattices — keep it that way.
- **Setting mode.** `TiltMachineNone` rests everywhere, which the list expressed as "every angle
  length". Under the new form it is a third case that must return settled without reaching a
  comparison.
- **The tie.** `step` prefers up on a tie. The page's `≥` in the perpendicular branch carries that;
  a `>` would silently change behaviour at 96 of the 1152 pairs.
