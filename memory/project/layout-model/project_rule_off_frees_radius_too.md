---
name: project-rule-off-frees-radius-too
description: A drag rule that is absent or inactive frees r as well as phi/theta, so the node follows the drag plane and can land far off the sphere — decided behaviour, not a bug
metadata:
  type: project
---

Turning a node's drag rule OFF in the polar rules panel frees **r, φ and θ together**, so
dragging that node follows the camera-facing drag plane and can land hundreds of world
units from the scene sphere. Observed 2026-08-14: node 2 went from r≈64 to r≈519 in one
drag with its rule off, and looked "flung".

This is DECIDED behaviour, chosen over two alternatives (give `r` its own independently
toggleable field; or have the toggle disarm only φ/θ and keep the radius hold). "Off" means
fully unconstrained.

**Why it looks like a bug and is not:** `polar.DragRule` has no radius field.
`TrimToDragRule` (`src/Node/Wiring/nodedrag/node_drag.go`) holds `out.R = have.R`
unconditionally whenever a rule applies, so the radius hold is a side effect of having ANY
rule — it is the only thing keeping a ruled node near the sphere. Dropping the rule drops it.

**The path predates the panel.** Only nodes 2 and 3 carry a `drag` key; nodes 4–9 have
none and take the identical `rule == nil` branch, so they have always dragged this freely.
The active toggle gave 2 and 3 access to existing behaviour, it did not create it.

**Do not "fix" this by:** clamping the drag target, bounding r in the gesture FSM, or
re-adding the radius hold for inactive rules. The drag math is sound — `RootMove`'s
`target.Sub(nm.SceneCenter())` is correct because `SceneCenter` is the scene ORIGIN, not the
node's own center (`nodegeom.NodeGeom`: world = SceneCenter + Polar2cart(ScenePolar)). That
frame was checked and cleared during this investigation; don't re-suspect it.

Related: [[project-topology-delta-position-disagreement-stays]],
[[feedback-bounds-and-who-caused-it]].
