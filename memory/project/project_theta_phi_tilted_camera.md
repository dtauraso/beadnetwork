---
name: project-theta-phi-tilted-camera
description: Apparent θ (height) mismatch in the editor is usually a φ (longitude) difference seen through a tilted camera; measure θ from world +y, not the screen. Layout uses a moving local pole, so the chart singularity is never reached.
metadata:
  type: project
---

**θ vs φ through a tilted camera.** What looks like a θ (height/latitude) difference
between two nodes in the editor is often a φ (azimuth/longitude) difference projected by
a rotated camera — screen-up ≠ world +y. Measure θ from world **+y**, not from the
screen, before claiming a θ bug. The polar frame markers (NavGuides.tsx: +y green / +x
red / +z blue, camera-independent) exist to make the world frame visible for exactly this.

Validated 2026-06: the reported "persistent θ mismatch between nodes 3 and 7" was NOT a θ
bug — 3 & 7 share θ exactly (θ-lock holds, `pair_theta` d=0.0000) and differ only in φ
(~171° apart); the mismatch was the φ separation seen through a tilted camera.

**No pole singularity in the layout.** The (θ,φ) chart is singular at its pole, but the
layout never measures against a fixed world +y. `requantizePoleTraced`
(`nodes/Wiring/quantized_move.go`) recomputes a **local measurement pole**,
`localPole(dirVecs)`, from the whole neighbor set on each move and persists it (`SetPole`).
Because that pole tracks the neighbors' actual directions — it moves a little as they move
— no node sits at the pole, so φ never blows up. Not a special-case guard; the moving pole
just never lands the singularity on a real node.
