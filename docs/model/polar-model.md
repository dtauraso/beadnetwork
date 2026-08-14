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
  `memory/project/layout-model/project_layout_model_evolution.md`).
- **A node is a point and an edge is the triple that closes the triangle.** A node's own
  point is `(r,φ,θ)` about the scene-sphere centre, in the QUANTISED integer form
  (`quantoffset.QuantizedOffset` — `ITheta`/`IPhi`/`IR` × per-node step constants,
  `nodes/Wiring/quantoffset/quantized_layout.go`), persisted (`nodes/<id>/position.json`
  `scenePolarR`/`scenePolarPhi`/`scenePolarTheta` + `quantITheta`/`quantIPhi`/`quantIR`).
  An edge carries `D`, the triple from its source to its target, persisted in that edge's
  own file under its source (`nodes/<src>/edges/<label>.json`, `deltaPolar*`). The two are
  one triangle:

  ```
  A = the source's own point       D = the edge's triple       B = the target's point
  A + D = B
  ```

  **`D` starts AT THE SOURCE**, and everything else follows from that. Because the vector
  begins there, its `φ` IS the angle from that node's own +y pole to its neighbour, its `θ`
  the turn around that pole, its `r` the distance — so the out-angle constraint is read off
  the very number it is applied to.

  **`D` is NOT `B − A` component-wise.** That is a difference of two coordinates which both
  start at the scene centre, and it is a different quantity: clamping it pinned exactly
  90.0000° while the angle in the triangle measured 99.79°, always past π/2. No constant
  closes that gap, because the real angle also depends on how far out the source sits and
  where its neighbour lies around the pole. `polar.Compose` (a followed by b) and
  `polar.Between` (from one point to another) resolve onto common axes — the one operation
  the three numbers cannot do separately. `Neg` is arithmetic on the numbers (`r`, `π−φ`,
  `θ+π`) and is exact to 1e-13.

  Placement WALKS: a node's point is its source's point composed with `D`, outward from each
  seed (a node no edge reaches, or the first node of a ring), which is the only point read as
  final.

  A move is the same triangle. Drag a node by `Δ`: every side it touches loses `Δ`, and the
  node at the other end of each of those edges takes that same `Δ` onto its own side. No node
  reads another node's position — it is TOLD the difference (`movemsg.Msg.Delta`, applied in
  `owners.Deltas`).

  **Do not "tidy" a triple after arithmetic.** Folding a result back into canonical ranges
  (`r ≥ 0`, `φ ∈ [0,π]`) removes nothing: each fold pays for itself by rewriting the other two
  components, so the triple points somewhere the arithmetic never said. One such fold broke
  `A + (B − A) = B` and walked node 2 outward 118 → 373 → 1891 across reloads, stacked on
  node 3 — the blow-up this page says cannot happen, reached through an arithmetic that
  quietly moved what it claimed to tidy.
- **A node holds its own side of every edge it touches** (`owners.Deltas`,
  `nodes/Wiring/nodeactor/owners/deltas.go`): the triple FROM ITSELF TO the node at the
  other end, stored from-self whichever way the edge points, so a move is uniform across
  in-edges and out-edges alike. The edge's own `D` is the out entry as-is and the negation
  of the in entry.

  The angle constraints (`φ = π/2`, `|θ| ≤ π/2`) are constraints ON `D`. They always were —
  they describe where a node sits about the one it hangs from, not its place in the world —
  so they are applied to the triple directly, with no holder frame to convert in and out of
  (`nodes/Wiring/nodeactor/node_drag_trim.go`).

  **The rule is carried by the node it binds, by id, and is applied BY that node.** Each node
  states its own `orbit` in its own `meta.json` (`polar.OrbitRule`); absent means free, and
  most nodes say nothing. No node reads another's rule: a neighbour that wants a node moved
  computes the `Δ` from ITS OWN numbers — its own point before and after, and its own side of
  the edge before and after (`nodes/input/drag.go`) — and TELLS it, and the node told
  trims that `Δ` against its own rules before committing it (`TrimOwnDrag`). A `Δ` an input
  node states to equalise its outgoing paths is therefore a request, not an imposition: a
  target whose own rule holds `D.r` keeps its distance and takes only the angles.
  A rule also holds `D.r`, so a node with one ORBITS its holder — its own drag cannot change
  the distance between them, which is what keeps an input node's two outgoing paths equal
  without anything being moved afterwards to repair them.

  They used to be two package constants reached through the HOLDER's kind — a node was
  constrained because something of kind `Input` pointed at it. That rule could not name the
  node it was about, so every target inherited the same angles, and the length a drag stated
  was then imposed on the target's siblings (`HeldSiblings`, deleted): dragging node 2
  teleported node 3, and the two read as welded together. What remains keyed by kind is only
  what is genuinely about being an input node — the shared length across its outgoing paths,
  and the half-turn snap on its own drag — and it is no longer a kind CHECK inside shared
  code. **A node owns the function that trims it, not only the numbers it trims to.** The
  kind states its own drag behaviour from its own package (`nodedrag.RegisterTrim` /
  `RegisterRequest` in `nodes/input/drag.go`), exactly as it states its ports; a kind that
  registers nothing is trimmed by its own orbit rule alone. `nodeactor` composes the delta,
  asks the node to trim it, and commits — it does not know what an `Input` is.

  `D.r` is a genuine DISTANCE — the length of the vector to the neighbour — so it is always
  at or above zero and "the longest path" is a plain maximum. (It was briefly a difference of
  two radii, and so could be negative; that was the same mistake as clamping the wrong phi,
  and it is gone with it.)

  **This is not the rejected double-link.** An earlier "local polar" model had each endpoint
  of an edge holding its OWN quantised bearing/distance to the other node, re-quantized
  node-to-node on every drag, and was measured disagreeing with the node's own scene polar
  (node 3 by +3.24 world units, node 4 by −3.08, against a step of 8.96 — two
  half-authoritative records for one position). Here there is ONE record: the edge file,
  owned by the one `edgeMover` that writes it. A target's in-entry is not a second
  authority — it is what its source TOLD it, it is never persisted, and no node's position
  is derived from it. The loader asserts the triangle closes on load, per component, and
  panics with both ids and all three triples if it does not.
- **Every bead on an edge is a placement ALONG that edge's stored path vector** — index ×
  `lattice.BeadStepR` from the source's rim along the path's direction, the first bead
  included. The first bead is no longer a separately-authored vector; its aim and the
  chain's length both read the one stored path above. The placement is owned by that bead's
  own goroutine (`nodes/wire/beadchain/bead_actor.go`'s
  `Bead`) — ownership + message passing, one writer, no locks/atomics
  (`tools/network/concurrency/check-no-network-locks.sh`, empty allowlist) — resolved from the node's own live
  aim broadcast (`BroadcastChain`, see the Chain bead bullet above) rather than a second
  stored copy. It is NEVER stored as an independent absolute position — it is computed at
  ONE site by summation: the node's world center (already `sceneCenter +
  polar2cart(scenePolar)`) plus this node-local vector, at the render/decode boundary
  — the summation site is GONE. The buffer used to stream a node's world centre and each
  bead's NODE-LOCAL offset as two columns, summed at one place on decode, so that moving a
  node cost one centre write rather than degree × N bead positions. That arrangement
  belonged to a chain a NODE laid toward a neighbour whose position it cached. Beads are
  now placed by the EDGE they travel, on the segment that edge already holds, and the
  buffer streams the world position itself: no offset, no origin, nothing to sum
  (`tools/topology-vscode/src/webview/three/scene/edges/edge-bead-blocks.ts`). Beads after the first
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

See [docs/model/polar-model-drag.md](polar-model-drag.md) for how a drag mutates the touching
edge beads (add/remove CRUD, the angle gate, the smallest-displacement commit rule) and how a
mutual pair offsets its two chains.
