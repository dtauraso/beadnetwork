# Which lattice a node lives on — resolved

This started as a PLAN framing the problem as "pick one of two lattices". That framing was
WRONG, and this file now records why: the symptom, the (accurate) finding about two
incommensurate lattices, the three options considered under the wrong framing, what was
actually measured live, and what was built instead — `walkBeadPath`
(`nodes/Wiring/quantized_move.go`), which has no grid in the drag path at all, so the two
lattices cannot disagree there because neither of them is present. Kept as history, not as
a plan to act on.

## The symptom

From live editor use: "the node drag and the beads moving appear to use different offsets."
Drag a node and its own chain of beads slides against it instead of moving with it.

## The finding: there are TWO lattices, and they are incommensurate

| | radial step | angular step | governs |
|---|---|---|---|
| **Scene** (`nodes/Wiring/quantized_layout.go`) | **20.0** | π/12 = **15°** | a node's ABSOLUTE position about the scene centre |
| **Local polar** (was `nodes/wire/layout_holder.go`, deleted with the whole local-polar model — MODEL.md "the polar model") | **8.96** | **1°** | a node's distance and bearing to each NEIGHBOUR — what `chain_beads.go` lays beads out from |

20 is not a multiple of 8.96 (it is 2.23×), so a node cannot sit exactly on both at once.
"One lattice", as [bead-lattice.md](bead-lattice.md) describes it, was only ever true of the
node↔neighbour relation. A node's own position has always been quantized on a separate,
coarser grid, and nothing reconciled the two — this part of the finding still stands.

This is not a new defect. It is a newly VISIBLE one: the local-polar cell used to be 2.0
world units, so the disagreement was around a world unit and read as nothing. Collapsing
onto the bead lattice made that cell 8.96, and the same structural gap now reaches 4.48
units — big enough to watch.

## What was measured

Driving a live drag and reading the fixed 1-degree scene-angular tick against a bead at
this graph's radii (28–70): radial steps came out around **9.0 world units**, in line with
the bead step, but ANGULAR steps came out **0.48–1.35 world units** — seven to eighteen
times SHORT of a bead. That is the concrete shape of "incommensurate": not a uniform
mismatch, but a fixed angular tick that shrinks in arc length the wider the graph gets,
against a radial tick that stayed roughly bead-sized by coincidence. A node dragged
sideways crept in tiny angular steps while a node dragged outward jumped a full bead
per step — the same drag looked like two different speeds depending on direction.

## Why the drag makes it worst

`layoutQuantizer.RootMove`'s own doc comment stated the intent plainly: the dragged node's
position is "the drag target itself — CONTINUOUS, not snapped to any grid ... the node's
position is free; only each neighbor's DISTANCE to it is quantized". So during a drag there
were effectively THREE positions in play — the raw cartesian pointer target that got drawn,
the scene-quantized triple that got persisted, and the local-polar quantization the beads
were computed from.

`commitNodeMoveLocal` already computed the quantized one (`measureScalar` → `nm.quantOffset`)
and persisted it. Only `nm.applyCenter(newPos, …)` drew the RAW target. So the quantized
position existed at commit time and was simply not the value the renderer was shown.

## Options considered (under the now-rejected "pick one lattice" framing)

Each is stated with what it would actually have bought, because the smallest change would
not have fixed the reported symptom. None of these were built.

### 1. Draw the scene-quantized position

`applyCenter` uses the position implied by the already-computed `quantOffset` instead of the
raw drag target. One line, no migration.

Would have bought: the drawn position and the PERSISTED position stop disagreeing, and a
drag becomes stepped rather than continuous.

Would NOT have bought: agreement with the beads. The node would step on the 20-unit/15°
scene grid while its chains are laid out on the 8.96 local one, so the two would still slide
against each other — less arbitrarily, but visibly.

### 2. Make the scene lattice the bead lattice

`stepR` 20 → `wire.BeadStepR` (8.96), angular steps 15° → 1°, so there would genuinely be
one lattice and a node's position would be a bead-distance multiple in the same terms its
neighbour distances are.

Would have bought: node and beads share one quantization by construction.

Would have cost: a migration of every stored `quantITheta`/`quantIPhi`/`quantIR` in
`topology/nodes/*/meta.json` and `position.json`, with the same distance-preserving
conversion the local polars just went through. Angular resolution would go from 15° to 1°,
so stored angular indices would multiply by 15.

### 3. Stop storing an absolute scene position at all

Derive every node's position from local polars alone, so the only stored geometry is
node-to-neighbour and there is no second quantization to disagree with.

Would have bought: the incommensurability cannot be expressed, rather than being made to
line up.

Would have cost: a rebuild of the layout model, raising where the graph anchors, what
happens to a node with no edges, and whether a solve stays stable under a drag —
`project_layout_model_evolution` records several rejected layout models this would have
needed checking against.

## What was actually built instead

None of the three. The framing itself — "a node's position lives on a grid, pick which
one" — assumed a grid was the right shape for a drag path at all, and that assumption is
what broke. The model David specified instead: a bead is a polar VECTOR of fixed length
(one `wire.BeadStepR`, end to end, identical in every direction by construction), and a
drag is a PATH of those vectors combined — "take the dragging of the node and fit it to a
path of the polar vectors ... the dragging should be vectors combining."

`walkBeadPath` (`nodes/Wiring/quantized_move.go`) implements this: starting from the node's
current drawn position, it advances toward the raw drag target one full `BeadStepR`-length
stride at a time, recomputing direction fresh every stride so the path curves toward a
moving target, and stops once the remaining distance is under one bead rather than sliding
a fractional bead. There is no radial index, no angular index, and no fixed tick anywhere
in this function — a step's length is the only invariant, and it is the same length in
every direction, so there is no separate radial/angular grid to reconcile and no
direction-dependent shortfall. The fixed 1-degree angular tick this doc measured as 7x–18x
short is gone, not reconciled with the radial one.

The scene-quantized `(iTheta,iPhi,iR)` triple still exists (`measureScalar`,
`nm.quantOffset`), but only as a self-describing CACHE measured back OFF the walked
position for `position.json` (`quant_offset_persist.go`'s own doc comment: "rides along as
a self-describing cache of the drag-time snap cells, NOT the position source") — nothing
downstream reconstructs the drawn/committed position from it. So the two lattices this doc
found still exist on disk, but the drag path that used to straddle them no longer reads
from either; it walks the bead vector. See [bead-lattice.md](bead-lattice.md) for the bead
model this vector reuses.

## Facts that stayed true and narrowed the eventual work

- **The cartesian drag target has exactly ONE live producer.** `RootMove`'s only callers are
  the drag path (`gesture_actions.go`) and `distance_groups.go`, and the latter has no
  references outside its own file — consistent with
  `memory/project_distance_group_decentralization_deadend.md`, which records that feature as
  built and reverted.
- **The boundary conversion already existed** in `commitNodeMoveLocal` — `cart2polar` once,
  then `measureScalar` — and still does, now feeding the post-walk cache measurement instead
  of the drawn position itself.
