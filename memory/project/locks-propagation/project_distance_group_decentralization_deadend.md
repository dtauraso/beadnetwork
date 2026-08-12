---
name: project_distance_group_decentralization_deadend
description: Per-edge decentralized distance groups were built and reverted; the unsolved part is what a child does when its parent moves
metadata:
  type: project
---

The "distance home button" groups ran as a CENTRAL update (dispatch computes each pair's
new position and calls RootMove, with `waitForCenterSettle` between pairs). A decentralized
rework — dispatch fans a target length to each edge, the edge tells its endpoint to move —
was built, merged, and REVERTED the same night (2026-07-29, revert `dffdc037`). Do not
re-propose it as-is.

**Why it failed, both times:**

1. **Chained pairs.** Groups chain: in `time` = (2,5)(2,4)(4,7)(4,6), node 4 is the TARGET
   of (2,4) and the SOURCE of (4,7)/(4,6). With every edge computing at once, (4,7) and
   (4,6) used node 4's position from before it moved. Target 93.027 → 2->5 and 2->4 reached
   it, 4->7 and 4->6 sat at 88.657 and 80.728. Deleting the ordered loop did not delete the
   dependency; it moved it into the edges.
2. **Shared targets.** In `gate`, node 8 is the TARGET of both (3,8) and (5,8). Both
   measured their displacement before either landed, 8 applied both, and overshot past
   either length — satisfying NO pair, where the central version satisfied two. A
   displacement is only valid against the position it was measured from.

**The question any design must answer first: when a parent moves, what happens to the
child?** A node's world position is its SCENE polar, not its stored `LocalPolar` to a
neighbour, and the existing one-hop propagation (`neighborSetCRequantize`) deliberately
makes a neighbour STAY PUT and re-quantize. So children do not follow parents for free, and
"each node just sets its own stored r" does not place it.

**David's stated direction (not built):** a node receives ONE delta and applies it as a
polar PERPENDICULAR-BISECTOR length, so both of its edges keep the same new distance —
which invites cartesian for world-scene things. He considers this the same problem he is
already working on from another angle, so it should be settled there rather than
re-litigated here. See [[project_layout_model_evolution]],
[[feedback_abc_times_constant_not_rederive]], [[project_lock_propagation_decentralized]].
