# Which lattice a node lives on

A PLAN, not a decision. It records a finding, the options that follow from it, and what
each one costs. Nothing here is implemented.

## The symptom

From live editor use: "the node drag and the beads moving appear to use different offsets."
Drag a node and its own chain of beads slides against it instead of moving with it.

## The finding: there are TWO lattices, and they are incommensurate

| | radial step | angular step | governs |
|---|---|---|---|
| **Scene** (`nodes/Wiring/quantized_layout.go`) | **20.0** | π/12 = **15°** | a node's ABSOLUTE position about the scene centre |
| **Local polar** (`nodes/wire/layout_holder.go`) | **8.96** | **1°** | a node's distance and bearing to each NEIGHBOUR — what `chain_beads.go` lays beads out from |

20 is not a multiple of 8.96 (it is 2.23×), so a node cannot sit exactly on both at once.
"One lattice", as [bead-lattice.md](bead-lattice.md) describes it, was only ever true of the
node↔neighbour relation. A node's own position has always been quantized on a separate,
coarser grid, and nothing reconciles the two.

This is not a new defect. It is a newly VISIBLE one: the local-polar cell used to be 2.0
world units, so the disagreement was around a world unit and read as nothing. Collapsing
onto the bead lattice made that cell 8.96, and the same structural gap now reaches 4.48
units — big enough to watch.

## Why the drag makes it worst

`layoutQuantizer.RootMove`'s own doc comment states the intent plainly: the dragged node's
position is "the drag target itself — CONTINUOUS, not snapped to any grid ... the node's
position is free; only each neighbor's DISTANCE to it is quantized". So during a drag there
are effectively THREE positions in play — the raw cartesian pointer target that gets drawn,
the scene-quantized triple that gets persisted, and the local-polar quantization the beads
are computed from.

`commitNodeMoveLocal` already computes the quantized one (`measureScalar` → `nm.quantOffset`)
and persists it. Only `nm.applyCenter(newPos, …)` draws the RAW target. So the quantized
position exists at commit time and is simply not the value the renderer is shown.

## Options

Each is stated with what it actually buys, because the smallest change does not fix the
reported symptom.

### 1. Draw the scene-quantized position

`applyCenter` uses the position implied by the already-computed `quantOffset` instead of the
raw drag target. One line, no migration.

Buys: the drawn position and the PERSISTED position stop disagreeing, and a drag becomes
stepped rather than continuous.

Does NOT buy: agreement with the beads. The node would step on the 20-unit/15° scene grid
while its chains are laid out on the 8.96 local one, so the two still slide against each
other — less arbitrarily, but visibly.

### 2. Make the scene lattice the bead lattice

`stepR` 20 → `wire.BeadStepR` (8.96), angular steps 15° → 1°, so there is genuinely one
lattice and a node's position is a bead-distance multiple in the same terms its neighbour
distances are.

Buys: node and beads share one quantization by construction. "A node moves one bead distance
at a time" becomes true of the node's absolute position, not just of its neighbour distances.

Costs: a migration of every stored `quantITheta`/`quantIPhi`/`quantIR` in
`topology/nodes/*/meta.json` and `position.json`, with the same distance-preserving
conversion the local polars just went through. Angular resolution goes from 15° to 1°, so
stored angular indices multiply by 15 and the layout is far more finely adjustable — that is
a behaviour change worth wanting on its own, not just a side effect. Note the last migration
of this shape could not hold while the editor was running, because each node's mover rewrites
its own files from memory; the durable fix there was normalizing at LOAD
(`LayoutHolder.LoadLocalPolars`), and the same approach applies here.

### 3. Stop storing an absolute scene position at all

Derive every node's position from local polars alone, so the only stored geometry is
node-to-neighbour and there is no second quantization to disagree with.

Buys: the incommensurability cannot be expressed, rather than being made to line up.
Arguably the truest to the model — the connection is the thing, and a node's absolute
position is a consequence of its relations.

Costs: a rebuild of the layout model. Position becomes a solve over the local-polar graph
rather than a stored value, which raises questions this plan does not answer — where the
graph is anchored, what happens to a node with no edges, and whether the solve stays stable
under a drag. `project_layout_model_evolution` records several rejected layout models; this
would need checking against them before it is attempted.

## Facts that narrow the work once a lattice is chosen

- **The cartesian drag target has exactly ONE live producer.** `RootMove`'s only callers are
  the drag path (`gesture_actions.go`) and `distance_groups.go`, and the latter has no
  references outside its own file — consistent with
  `memory/project_distance_group_decentralization_deadend.md`, which records that feature as
  built and reverted. So converting the drag to index arithmetic is contained.
- **The boundary conversion already exists** in `commitNodeMoveLocal` — `cart2polar` once,
  then `measureScalar`. Nothing new is needed to get from a pointer ray to an index triple.
- **The grab offset added by `gesture_actions.go` is a cartesian `vec3`**, which is contrary
  to `memory/feedback_abc_times_constant_not_rederive` — positions should move by index
  arithmetic, with cartesian confined to the boundary. Whichever option is chosen, that
  offset should become an index delta captured at the drag-start edge.

## What has to be decided

Which lattice a node's POSITION lives on. Everything above follows from that one answer, and
none of it should be built before it.
