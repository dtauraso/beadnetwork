---
branch: task/arith-math-into-code
---

# Run the arithmetic page's maths in Node1

The page (`docs/pair-node/arith.html`) and the code (`nodes/Node1/machine.go`) agree on every
decision but share no arithmetic. `TestOneRoundIsSignAndRemainder` is what holds them together:
it sweeps both arrangements and all (tilt, arrival) pairs asserting the page's forms pick exactly
what `settled` and `step` pick — but it derives the page's forms in test-local arithmetic and
never calls the machine. The change is to make the code compute the page's way, so the agreement
is structural rather than proven after the fact.

The section that drives this is arith.html's **"Which end moves, and whether it goes +1, −1 or
stays"**. Not the stops-at card, not "The tilt it stops on", and not the pages that cite
`restingLengths`— those follow the code, they do not lead it.

Delete this file when the change lands.

## What changes, and what does not

**The mode stays data.** Nothing here branches on which mode is running, before or after. That
property is what `machine.go` was folded into one file to get, after three attempts that
parameterised the rule failed by leaking a change for one mode into the other; the header records
it and the header stays true. A mode contributes a number and NOTHING ELSE — the change is that
the number is one stopping count instead of a list of resting lengths.

**The update rule stops being a state machine and becomes arithmetic.** That is the whole of it.

```
now   fromRest = min over r in restingLengths[mode] of |angleLength(t, a) − r|
      settled  = fromRest == 0
      step     = fromRest(t+1) <= fromRest(t−1) ? t+1 : t−1

page  c        = the count from the nearer end of the tilt line, on a ring of halfTurn
      settled  = c == the count this mode stops at
      step     = walk c one slot toward that count, tie goes up
```

The current form folds to an angle length, takes a minimum over a list, then re-evaluates that
whole measurement at both neighbours and compares the two results — a machine asking "which of my
next two positions is better" without ever computing where it is going. The page's form computes
one count and compares it to one number.

## The claim the change rests on

The page prints two four-line blocks, one per arrangement:

```
parallel        c = quarter : stay     c < quarter : +1     c > quarter : −1
perpendicular   c = 0       : stay     c ≥ quarter : +1     c < quarter : −1
```

These read as two different rules and are one: **`c` lives on a ring of `halfTurn` slots, each
arrangement stops at one count on it (perpendicular 0, parallel `quarterTurn`), and the node walks
`c` one slot the short way toward its own stop, a tie going up.** Parallel's `c < quarter → +1` is
"the stop is nearer going up". Perpendicular's `c ≥ quarter → +1` is the same sentence about 0,
short way round.

This is the load-bearing claim and it is NOT yet verified. Step 2 exists to verify it against the
rule that runs today, on every pair and both lattices, before anything is switched over. If it
fails, the change stops there and the plan is wrong.

## Both ends are stored, and the update drives the nearer one

A second change, and the reason the page and the code still read differently even once the
arithmetic matches.

Today `Top` is the only stored tilt and `Bottom` is `Top.opposite`, a link. So every update is
written to `Top` — including the half where the count was read off the BOTTOM.

Target: store both ends, and have an update write the end it measured.

```
nearer end is the top      →  Top    = Top ± 1        then Bottom = Top.opposite
nearer end is the bottom   →  Bottom = Bottom ± 1     then Top    = Bottom.opposite
```

Behaviour is unchanged — the ends are a half turn apart, so moving either moves both — but the
code then says WHICH end was driven, and the page's two halves become symmetric: each measures an
end and moves that end.

This is not cosmetic once the arithmetic lands. `c` is measured at the nearer end, so a `±1` on
`c` is a `±1` on THAT end. Writing it to `Top` regardless would need a sign flip in the half where
the bottom was measured — reintroducing by hand the reversal the page's keynote says does not
exist when each end is read on its own count.

What it touches beyond the arithmetic:

- `Node.Top`'s doc comment calls itself "The ONE writer, full stop". That becomes two writers of
  one line, and the comment has to say how they cannot disagree: each writes the other from its
  own `opposite` in the same step.
- Persistence is unaffected — `position.json` keeps one angle, and the second end is still
  derivable. Storing it is about which end an update names, not about what survives a reload.
- `syncTiltIndex` already reports all three arrows; it keeps doing so from whichever end moved.
- The page's second half changes `t_after` to `b_after`, and says `t` follows.

## Order

0. **The sweep runs both lattices.** `TestOneRoundIsSignAndRemainder` is `const points = 24` with
   literal 6s and 12s throughout. It cannot be the safety net for a change that must hold on 24 and
   48 while it says 24. Literals become `r.quarterTurn`/`r.halfTurn` and the body runs under both.
   This commit changes no production code.
1. `machine.go`: add the page's arithmetic beside the existing rule, not replacing it — the count
   from the nearer end, and a per-mode stopping count. Nothing calls them yet.
2. Extend the sweep to assert the new functions equal `settled`/`step` on every pair, both
   lattices. This is the safety net for step 3 and the test of the claim above; it must be green
   before step 3.
3. Switch `settled` and `step` to the new arithmetic. `restingLengths` becomes one stopping count
   per mode; `resting` and `fromRest` go. Update the header where it says a mode contributes a
   LIST, and where it describes step as a comparison of the two neighbours.
4. Store both ends on `Node`, the update driving the nearer one and writing the other from its
   `opposite` in the same step. Rewrite `Top`'s "ONE writer" comment to say why two writers of one
   line cannot disagree.
5. Repoint the nine `data-src` citations — `restingLengths` ×7 (`formulas.html`, `dots.html`,
   `perpendicular.html`) and `resting` ×2 (`updates.html`). `bash tools/check-docs-symbols.sh`
   names each failure.
6. `updates.html`: the resting-length card and the two keynote lines say "a list of numbers". The
   argument survives — the mode is still data the rule cannot interrogate — so this is list→count,
   not a rewrite of the case being made.
7. `MODEL.md` and `.claude/rules/node-kinds.md` — grep for `restingLengths` and for "a mode
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
  `r.quarterTurn` and `r.halfTurn`, and `c` lives on a ring of `halfTurn`, not of `points`. Step 0
  exists because the sweep cannot catch a 24-only assumption while it is itself 24-only.
- **Setting mode.** `TiltMachineNone` rests everywhere, which the list expressed as "every angle
  length". A single stopping count cannot say that, so setting needs its own answer — settled
  wherever it stands, without reaching a comparison — and `machineFor` must still know it.
- **The tie.** `step` prefers up on a tie. The page's `≥` in the perpendicular branch carries that,
  and "tie goes up" on the c-ring must reproduce it at every tied pair, not most of them.
- **Which way `c` counts.** `c` is `(t − a)` reduced, not `(a − t)`. The two differ by a sign and
  agree at the stops, so a swap survives `settled` and breaks `step` — in a direction the tie cases
  hide. The sweep must compare `step`'s chosen state, not just its distance.
