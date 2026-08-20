# Model — dragging a node (bead CRUD)

[← polar-model.md](polar-model.md)

## Moving a node: CRUD on the touching edge beads

- **Moving a node is CRUD on the edge beads that touch it (drag placement).** N chains
  connect to a node; you move the node by removing links from those chains or adding links
  to them — that is the whole mechanism. There is no solver, no constraint system, no
  enumeration across neighbours: each touching bead decides for itself
  (`src/Node/Wiring/beadcrud/bead_crud.go`'s `BeadCrudDecide`, wired in `CommitNodeMoveLocal`,
  `src/Node/Wiring/nodemove/commit_node_move.go`).

  The drag gives the node's own polar vector `v` (its previous position to its
  destination). Each touching bead has its own **source point** — the previous bead's
  centre along its chain, or the chain origin on the neighbour's torus surface when it is
  the only bead (`src/Node/Wiring/beadcrud/touching_beads.go`'s `DragTouchingBeads`) — NEVER the
  touching bead's own centre; using the centre instead is wrong by one bead. The **third
  polar vector** runs from the bead's source point to the node's destination point.
  Compare its length to one bead length (`lattice.BeadStepR`):

  - too small → that bead is **removed**, and the bead before it becomes the touching bead.
  - too large → a bead is **added** (subject to the angle gate below), and it becomes the
    new touching bead.
  - exactly one bead length → nothing changes.

  **The angle gate applies to ADD only, never to REMOVE.** The angle between `v` and the
  edge-bead vector (source → the touching bead's own centre): > 90 degrees blocks the add
  (the node did not move far enough AWAY from the bead to open a gap beyond it — an obtuse
  angle means the drag is heading back across the bead); ≤ 90 degrees admits it, subject to
  the `|third|` test above. A removal is decided by `|third|` alone.

  **There is no selection and no summation, and the node's new centre comes from the BEAD
  OPERATION, never from `v`.** Every touching bead performs the same judgement against the
  same `v`, but `v` (the drag) supplies only the third-vector test and the angle gate above
  — it never sets the node's new position or its direction of travel
  (`src/Node/Wiring/beadcrud/bead_crud.go`'s `BeadCrudImpliedCentre`, resolved across every touching
  bead by `ResolveBeadCrudMove`):

  - **REMOVE** → the node's new centre IS the removed bead's own former centre exactly — it
    takes that bead's place.
  - **ADD** → a new bead is inserted one bead length closer to the node than the old
    touching bead (the "next chain position"); the node's new centre is one bead length
    BEYOND that new bead, away from the neighbour, along the SAME chain axis (never the raw
    drag direction).
  - every touching bead's verdict is "none" → the node does not move; with no touching beads
    at all (a free node with no incident edges) the raw target is used directly.

  One drag event can remove beads from some edges and add them to others at once, each
  independently implying its own new node centre. A node with several neighbours has
  touching beads on several different chain axes, so their implied centres essentially
  never coincide — that is the ORDINARY multi-neighbour case, not a conflict (treating
  disagreement as a conflict and holding the node still made every multi-neighbour node
  immovable). The commit is the implied centre with the SMALLEST displacement from the
  node's previous centre among all verdicts — never an average, never nearest-to-cursor,
  never the raw drag target. Movement stays one bead at a time; an edge whose verdict
  implied a larger step reaches it over successive pointer-move events instead of in one
  jump (the edge step count re-counts against the live distance every commit).

  Bead count on an edge falls out of the resulting geometry as one integer subtraction
  (`src/Node/Wiring/edgegeom/chain_length.go`'s `EdgeStepCount`), with the near end tangent to the node's
  own torus by construction of the placement formula and one uniform global bead size
  (`src/Node/BeadAnimation/lattice/bead_lattice.go`).
- **A mutual pair (two nodes each pointing an edge at the other) offsets its two chains to
  opposite sides**, so they do not draw on top of each other. `edgegeom.ParallelChainOffset`
  (`src/Node/Wiring/edgegeom/parallel_chain_offset.go`) computes the offset from the pair's own two centres and
  the scene centre, in CANONICAL id order (NUMERICALLY smaller id first — a string compare
  would put "10" before "2" and hand both ends the same side, collapsing the two chains back
  onto one line) so both
  endpoints derive the SAME side independently — neither node needs to know what the other
  decided. The offset stays INSIDE that pair's own ring plane (not along a fixed world
  axis), so it composes with coplanar rings rather than fighting them. `chain_beads.go` is
  guarded against doing this vector math itself (`src/Node/BeadAnimation/check-no-sqrt-in-chain-beads.sh`);
  it calls into `src/Node/Wiring/edgegeom/port_geometry.go` for it, same split as `edgegeom.EdgeCenterDistAndDir`.
