---
name: project-theta-phi-tilted-camera
description: Apparent θ (height) mismatch in the editor is usually a φ (longitude) difference seen through a tilted camera; measure θ from world +y, not the screen.
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
