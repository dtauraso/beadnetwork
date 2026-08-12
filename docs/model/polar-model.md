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

See [docs/model/polar-model-drag.md](polar-model-drag.md) for how a drag mutates the touching
edge beads (add/remove CRUD, the angle gate, the smallest-displacement commit rule) and how a
mutual pair offsets its two chains.
