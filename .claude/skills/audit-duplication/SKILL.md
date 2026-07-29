---
name: audit-duplication
description: Periodic structural-duplication audit (dry4go) over the Go tree. Finds copy-paste that survived renaming — which grep cannot. NOT a guard, NOT in stop-checks; its false-positive rate here is structural and permanent.
---

# Duplication audit — dry4go

Run it, read it against the baseline below, act only on what is NOT in the baseline.

```
GOFLAGS=-mod=mod go run github.com/unclebob/dry4go/cmd/dry4go@latest --threshold 0.9 ./nodes ./Buffer
```

Pinned-version note: unlike `tools/audit-mutation-crap.sh`, this uses `@latest` because it
is a review aid, not a measurement whose numbers get compared across runs. If you ever start
tracking the count over time, pin it.

## What it actually does

Parses every Go function, normalizes it to a syntax tree, and scores pairs by Jaccard
similarity over structural fingerprints. **Identifiers, local names, and literal values
normalize away.** That is the whole point: it finds copy-paste that survived a rename, which
grep cannot, because the two copies may share no tokens at all.

## Read a score of 1.00 as "same SHAPE", never "same behavior"

Literals normalize away, so `x%2==1` and `x%2==0` are indistinguishable to it. In this repo
that is not hypothetical: **`selectleft` and `selectright` are deliberately mirror-image
logic and score as a perfect duplicate.** Merging them would be a bug, not a cleanup.

## Baseline — do NOT report these

Measured after the self-construction rework landed. All are deliberate:

| pair | lines | why it is intentional |
|---|---|---|
| `Time` ↔ `TimeStart` node bodies | 115 | `TimeStart`'s own doc: "a clone of the Time kind … split out so node 2's future divergence has a home of its own" |
| `pulse` ↔ `PulseLeft` ↔ `PulseRight` | 63 ×3 | same clone-for-future-divergence pattern |
| `Time` ↔ `TimeStart` init blocks | 32 | `RegisterBuilder` declaration — see "the cost this measured" below |
| `selectleft` ↔ `selectright` init blocks | 27 | same, and the bodies are mirror-image by design |
| `firing_rule_lean_test.go` across kinds | — | per-kind test boilerplate |

Total at threshold 0.9: ~95 pairs. The overwhelming majority are the above plus test
scaffolding.

**A kind CAN legally share code** — `nodes/gatecommon` is the shared spine and
`check-dep-rules.sh` permits kind→spine imports. `gatecommon.DriveHeld` and `RunGate` are
existing proof. So "kinds must repeat themselves" is FALSE; these clones are a deliberate
choice about anticipated divergence, not a constraint.

## The cost this measured

Before the self-construction rework: 93 pairs. After: 95. It went UP.

Removing the central reflection pipeline meant every kind writes its own explicit
`RegisterBuilder` block, and those blocks are near-identical across kinds — the two 32/27-line
entries above did not exist before. The rework bought compile-time safety (a renamed field is
now an error instead of a silently nil one) and paid for it in repetition. That is a real
trade, and worth remembering before anyone proposes "reducing duplication" in the init blocks:
the previous way of avoiding it is the thing that was deliberately deleted.

## The actual signal

**A new high-scoring pair OUTSIDE `nodes/<Kind>/`.** Two functions in `nodes/Wiring`,
`Buffer`, or `nodes/wire` converging is worth reading — that is the spine, where sharing is
the norm and accidental divergence-then-reconvergence is a real smell.

Everything inside a node-kind directory is presumed deliberate until shown otherwise.

## Why this is a skill and not a guard

As a `tools/check-*.sh` it would need an allowlist covering most of `nodes/` — a guard that
guards nothing. Its false positives here are structural and permanent: the clone families are
staying. Run it when you want a map, not on every commit.
