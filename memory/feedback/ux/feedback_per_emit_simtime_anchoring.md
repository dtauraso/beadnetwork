---
name: Anchor pulse animations at emit-time, not mount-time
description: Validated 2026-05-04 — per-pulse emit-timestamp anchoring + concurrent rendering on one edge is the right shape for view/model temporal coupling in this codebase
type: feedback
---

When a pulse / particle / animation represents a model emit, anchor its rAF math at the emit's step-pinned `simTime` and render all queued instances concurrently. Do **not** queue serially behind the head and capture `swapStart` at React-mount time.

**Why:** validated 2026-05-04 on `task/unified-sim-clock`. The user's report was "fold node halo is more stable and snappier at closer to the right times" after switching `AnimatedEdge` from "render `pulses0[0]` only, capture swapStart at mount" to "render all pulses concurrently, anchor each at its captured `simStart`." Three niches converge on this shape — games (per-particle birth time), DAW transport (per-clip start position), packet-flow viz (per-packet spawn) — so it generalizes to any future emit→pulse style animation in this plugin, not just edge pulses.

**How to apply:** if you find yourself writing "head-of-queue plays first, next mounts on completion" for a model→view animation, that's the wrong shape — push the emit timestamp through and let instances animate independently. The "view sees emit when previous viz finished" decoupling is the symptom.

Companion: fix-a (`getSimTime()` step-pinning in `runner.ts`) is the prerequisite — without it, emit-time and animation-time are drifting against each other inside the same synchronous step, so even per-emit anchoring would still be off by 0.1–5ms per step.
