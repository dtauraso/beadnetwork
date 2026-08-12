# Model — scenes

[← MODEL.md](../../MODEL.md)

## Scenes

A **scene** is a complete, independently loadable topology tree — its own `counts.json`,
its own `nodes/`, its own `view/`. There is more than one: they are SIBLING directories next
to each other (`nodes/Wiring/scene/scene_tabs.go`'s `SceneTabs`, e.g. `topology/` and
`topology-pair/`), not one topology with variants inside it. Go owns the list, the labels,
which one is selected, and the switch — this is the same shape as the overlay toggles, not
a new mechanism. The `-topology` path the extension host launches with is the fixed ANCHOR;
which sibling actually loads is resolved from it (`ResolveScenePath`) and persisted at the
anchor, never inside a scene (a selection stored inside scene B would be unreachable while
scene A is loaded — `.claude/rules/persistence-ownership.md`). TS renders the tab strip
from the VIEW frame and forwards a click as one addressed edit (`kind="scene"
attr="selected"`); it holds no list, no labels, no selection of its own.

A tab switch is not an in-process rebuild. There is no teardown of live node goroutines or
their in-flight beads mid-traversal — persisting the new selection ends the Go process, and
the extension host's already-looping runner respawns it, which re-reads the selection and
loads the other tree. This is deliberate: a respawn is machinery that already exists (the
`.go` file watcher triggers the same path on every edit), so switching scenes buys nothing
by adding a second, in-process path to do the same thing.

**A scene may fork small pieces of node behavior, and each fork is a named, reasoned
choice — not a tuning knob:**

- **Drag mode** (`SceneTab.QuantizedDrag`). The default is the CONTINUOUS drag: a node is
  one point, an edge is the distance between two of them, and a drag says where the point
  went — the bead count on that edge then fills whatever line that leaves. The alternative,
  quantized drag steps the node one bead length (`lattice.BeadStepR`) at a time, matching its
  own chain beads' jump size — it exists so a dragged node and its own beads move in the
  same quanta, instead of the node gliding while its beads jump. That only works when the
  step is small AGAINST THE SCENE: the ring spans ~500 world units, so a ~9-unit step reads
  as smooth, while a two-node scene ~40 across moves a fifth of itself per step and cannot
  be positioned at all. No committed scene uses it today — both tabs drag continuously — but the path
  stays reachable (and is what an unrecognised tree gets) rather than deleted, since
  deleting a still-tested, still-reachable mechanism is a model decision, not a cleanup.
- **Coplanar rings** (`SceneTab.CoplanarEdges`). A node's ring plane normally follows its
  INWARD pole (toward the scene centre), which says nothing about where its neighbour is,
  so an edge lies in that plane only by coincidence — the chain then runs through the
  tori's holes rather than across their faces. With this on, for a node with exactly ONE
  neighbour, the ring axis is swung the smallest amount off the inward pole that still puts
  the edge inside the ring plane, so the chain, the node's torus, and the beads' own tori
  all lie in one plane. A node with more than one neighbour keeps its inward pole — no
  single plane contains two non-collinear edges — so this is inert there by construction,
  which is why it is a per-scene choice rather than a global rule.

**The drawn ring axis and the navigation pole are two different streamed values, on
purpose.** `PoleTheta`/`PolePhi` (`Buffer/layout.go`) is a node's own INWARD pole — its own
scene-polar direction reversed, pointing back at the scene centre — and is what navigation
reads (`buffer-nav.ts`'s `NavNode.pole`). `RingAxisTheta`/`RingAxisPhi` is the axis the
node's RING is actually drawn on, defaulting to the torus's own +Z normal (unrotated) and
diverging from the inward pole only under coplanar rings (above). Keeping them as separate
columns is what lets a scene ask for coplanar rings without touching what navigation reads,
and vice versa. Today every node's own local FRAME is held at world +y regardless of its
pole — an earlier version rotated node 1's frame alone onto its streamed pole, which read as
a property of that one node rather than of the feature, and was reverted so every frame
reads the same way. `PoleTheta`/`PolePhi` is consequently live on the wire (navigation still
reads it) and not rendered as a frame rotation for any node — whether a node's frame should
instead follow its own pole is an **open question**, not decided here.
