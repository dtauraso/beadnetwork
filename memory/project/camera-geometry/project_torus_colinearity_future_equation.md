---
name: project_torus_colinearity_future_equation
description: port∈torus is polar-only (no node-center z-drag); ports snap to their rings and the edge bends between them — port/edge/torus colinearity is not measured or enforced, and no rebuild is planned
metadata:
  type: project
---

The `port ∈ torus` lock was made polar-only in `task/torus-polar-rebuild`
(merged 2026-07-05, commit 2b19c2cd): it now ONLY pins a port to its own node's
border ring via `portWorldPosAimed`'s polar ring-projection about the node
sphere. The old behavior that dragged node CENTERS in cartesian z
(`newWorld.Z = fromWorld.Z` in lockRecalc + `applyPortTorusColinearity`) was
deleted — it fought position equations and caused the `(3,r)=(6,r)` oscillation
(nodes swung to ~5165 then settled; the torus z-lock and the r-equation both
wrote a node's z).

**What the polar lock does:** each torus-locked port snaps to its own node's
border ring. The edge between two such ports runs straight from ring point to
ring point and bends between them. Colinearity of the two ports, the edge, and
the node centers is not measured and not enforced. No colinearity rebuild is
planned — do not re-open it as queued work.

Still present in the code (not dead; underpins the polar port∈torus lock):
`portTorusLocked`, `portWorldPosAimed` / `ringProjectDir` / `partnerTorusLocked`
/ `ringAnchorDir`, and the eqPortTorus authoring/persist/display path.

**2026-07-05 (`task/eq-show-source-node`): `eqPortTorus` authoring is now
own-node-only.** The torus slot is no longer a free second-node pick — it is
PRESET to the port's own node (the sticky Center), enforced at the commit site
(`gesture.go addPortTorusLock` always sets `TorusNode = PortNode`) and in the
webview form (`TypedPortTorusForm` renders the torus cell from the port's own
node label, never a typed input). A cross-node `port ∈ torus` lock can no
longer be authored.
