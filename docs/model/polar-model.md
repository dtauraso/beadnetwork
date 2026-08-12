# Model — node positions & movement locks (the polar model)

[← MODEL.md](../../MODEL.md)

## Node positions & movement locks (the polar model)

Editor-time node geometry and lock propagation are **pure polar**. The scene sphere's center
is the only cartesian value that is **persisted and authoritative** — the world anchor. It is
not the only cartesian value that exists: the camera pose, port anchors, bead segments and
per-node world centers are cartesian too. The invariant is narrower and stronger than
"cartesian appears once": every other cartesian is **derived** from this anchor
(`sceneCenter + polar2cart(…)`) or **quarantined at the renderer edge** — none is persisted,
and none is a source of truth.

- **Scene sphere** — a first-class, persisted reference (NOT the derived content-sphere
  centroid, which moves with the nodes and is circular). It has a **cartesian center** (the
  one world anchor, in `scene.json`) and a **radius** that fits the diagram.

  **It is established once and never moves.** `LoadSceneSphere` reads it from `scene.json`,
  or content-fits it from the node centers and persists that immediately — persisting matters,
  because every node position is a polar measured ABOUT this center, so a center re-fitted
  over moved nodes on the next load would silently reinterpret the whole diagram.

  It is a SEPARATE entity from the camera pivot, and **both camera gestures leave it alone**:
  orbit must not move it, and `PanViewpoint` is a pure CAMERA move that deliberately does not
  touch it.

  > **Rejected: pan moves the sphere.** The model once said camera pan should translate the
  > center by the same delta, holding node world positions fixed while their scene polars
  > recomputed about the new center. Coupling pan/dolly to the sphere left `ui.SceneSphere`
  > diverged from the movers' held center until a later broadcast reconciled it with a jump —
  > the "zoom got canceled" symptom. It was decoupled, a separate scene-pan gesture was named
  > as the proper home, and that gesture was never built. The claim is now DROPPED rather than
  > pending: the sphere is a load-time-fitted constant.
  >
  > The cost, stated so it is a choice and not an accident: the polar frame is best
  > conditioned near the center it was fitted to. Pan the camera far away and drag there, and
  > you work at large `r`, where a small angular step moves a node a long way. If that becomes
  > real friction in the editor, THAT is the reason to revisit — and the trap to avoid is the
  > one above: a scene pan is its own gesture, never a side effect of a camera move.
- **A centre is a sum of polar vectors from ONE centre — the scene sphere's centre.**
  `scene centre --vector--> node`, `node --vector--> that edge's FIRST BEAD`, and
  `bead centre = (scene centre -> node) + (node -> bead)`. There is NO hierarchy between
  nodes: every node hangs off the scene centre directly, one hop (scene -> node -> bead,
  two levels — a rooted/parent layout was tried and rejected,
  `memory/project/project_layout_model_evolution.md`).
- **A node has ONE polar coordinate.** `(r,θ,φ)` about the scene-sphere center — the node's
  whole POSITION, in the QUANTISED integer form (`quantoffset.QuantizedOffset` — `ITheta`/`IPhi`/`IR`
  × per-node step constants, `nodes/Wiring/quantoffset/quantized_layout.go`), persisted
  (`nodes/<id>/position.json` `scenePolarR`/`scenePolarTheta`/`scenePolarPhi` +
  `quantITheta`/`quantIPhi`/`quantIR`). World = `sceneCenter + polar2cart(scenePolar)`.
  **A node carries NO stored coordinate for a NEIGHBOUR.** An earlier double-link
  "local polar" model existed here — each endpoint of a domain edge holding its own
  quantised bearing/distance to the OTHER node, re-quantized node-to-node on every drag —
  and was measured disagreeing with the node's own continuous scene polar (node 3 by
  +3.24 world units, node 4 by -3.08, against a step of 8.96 — two half-authoritative
  records for one position). It was deleted entirely, along with every mechanism that
  existed only to maintain it (the per-node neighbour-coordinate type and its
  requantize-on-drag propagation, and the file it persisted to).
- **A node has ONE polar vector PER EDGE, pointing to that edge's STARTING bead.** This
  vector is owned by that edge's FIRST BEAD's own goroutine (`nodes/wire/beadchain/bead_actor.go`'s
  `Bead`) — ownership + message passing, one writer, no locks/atomics
  (`tools/network/concurrency/check-no-network-locks.sh`, empty allowlist) — resolved from the node's own live
  aim broadcast (`BroadcastChain`, see the Chain bead bullet above) rather than a second
  stored copy. It is NEVER stored as an independent absolute position — it is computed at
  ONE site by summation: the node's world center (already `sceneCenter +
  polar2cart(scenePolar)`) plus this node-local vector, at the render/decode boundary
  (`tools/topology-vscode/src/webview/three/scene/node-stream-blocks.ts`'s `getChainBeads`,
  `cx + readChainBeadOX(...)` and its Y/Z siblings) — the buffer streams the node's world
  center and each bead's NODE-LOCAL offset as two separate columns on purpose (constant-time
  node moves: moving a node costs one center write, not degree × N bead positions), and
  this is the one place they are summed into an absolute bead centre. Beads after the first
  keep their existing chain-relative placement (index × `lattice.BeadStepR` along the same aim)
  — this model change is about the node's coordinate, the per-edge first-bead vector, and
  this one summation site, not the rest of the chain.
- **The node STORES ITS SUM** — its own world center (`NodeMover.geom`, written once per
  move by `ApplyCenter`) — so a bead's polar view starts from that stored sum rather than
  re-walking to the scene centre on every read. With no ancestors, only that node can
  invalidate it, and it owns it (single-writer, its own goroutine).
- **No blow-up, by construction.** The offset is STORED and only carried through the
  composition or nudged one component — it is NEVER re-derived as `cart2polar(node − center)`
  from a live world during a propagation wave. That reconstruction against a mid-moving center is the
  bug that made positions fly to infinity. A moved center rigidly translates its satellites
  (offset unchanged ⇒ locks stay satisfied ⇒ the wave terminates). This is STRUCTURAL, not a
  test: the reconstruction that caused the blow-up has no call site to write. Nav is held
  polar-only by `tools/webview/check-polar-only-nav.sh`.
- **Panel-authored locks must be structurally incapable of a position blow-up.** If one
  happens, the implementation is wrong (an offset was reconstructed from a moving reference),
  not the locks.
- **Moving a node is CRUD on the edge beads that touch it (drag placement).** N chains
  connect to a node; you move the node by removing links from those chains or adding links
  to them — that is the whole mechanism. There is no solver, no constraint system, no
  enumeration across neighbours: each touching bead decides for itself
  (`nodes/Wiring/beadcrud/bead_crud.go`'s `BeadCrudDecide`, wired in `CommitNodeMoveLocal`,
  `nodes/Wiring/layoutquant/commit_node_move.go`).

  The drag gives the node's own polar vector `v` (its previous position to its
  destination). Each touching bead has its own **source point** — the previous bead's
  centre along its chain, or the chain origin on the neighbour's torus surface when it is
  the only bead (`nodes/Wiring/beadcrud/touching_beads.go`'s `DragTouchingBeads`) — NEVER the
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
  (`nodes/Wiring/beadcrud/bead_crud.go`'s `BeadCrudImpliedCentre`, resolved across every touching
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
  (`nodes/Wiring/nodegeom/chain_length.go`'s `EdgeStepCount`), with the near end tangent to the node's
  own torus by construction of the placement formula and one uniform global bead size — see
  `docs/bead-model/bead-lattice.md`.
- **A mutual pair (two nodes each pointing an edge at the other) offsets its two chains to
  opposite sides**, so they do not draw on top of each other. `nodegeom.ParallelChainOffset`
  (`nodes/Wiring/nodegeom/port_geometry.go`) computes the offset from the pair's own two centres and
  the scene centre, in CANONICAL id order (NUMERICALLY smaller id first — a string compare
  would put "10" before "2" and hand both ends the same side, collapsing the two chains back
  onto one line) so both
  endpoints derive the SAME side independently — neither node needs to know what the other
  decided. The offset stays INSIDE that pair's own ring plane (not along a fixed world
  axis), so it composes with coplanar rings rather than fighting them. `chain_beads.go` is
  guarded against doing this vector math itself (`tools/network/beads/check-no-sqrt-in-chain-beads.sh`);
  it calls into `nodes/Wiring/nodegeom/port_geometry.go` for it, same split as `nodegeom.EdgeCenterDistAndDir`.
