---
name: project_wire_is_straight_line_not_chain
description: The original bead-item chain model's O(N²) drag lag (~1.5s at N≈40) was human-clock gating + momentum-free midpoint relaxation, NOT "a chain of beads" per se. A corrected primitive (nodes/wire/bead_actor.go, body-force geometry + broadcast wake) is built and tested in isolation but as of task/two-clock-beads has NO production call site — the live chain-bead render path is still chainBeads()'s unchanged per-frame computation.
metadata:
  type: project
---

Wires render as straight `wireSegment` / `lerp` lines drawn by `PacedWire` + `SingleEdgeTube`, with one moving `PulseBead`. A "bead-item chain" wire model (a wire as N bead-sized goroutines that relax to straight via midpoint-of-neighbors, born/retired to hold spacing, the pulse hopping item-to-item) was fully built and then **reverted** (revert commit `e7faf250`; MODEL.md chain section dropped in `45521898`).

**The measurement (~1.5s at N≈40) is right. The original conclusion — "don't re-propose the chain model" — was attached to the wrong cause.** Re-diagnosed under `task/two-clock-beads` (PLAN.md "two clocks per bead"): the O(N²) latency had TWO causes, and "a chain of beads" was not one of them:

1. **Human-clock gating.** The original model advanced/relaxed beads on the human-speed presentation clock (one hop per tick, `MsPerTick`). A goroutine paced by that clock cannot resolve position at machine speed no matter how the relaxation math is fixed — the tick rate becomes the propagation rate.
2. **Momentum-free midpoint relaxation.** Neighbor-only midpoint averaging is discrete diffusion — the chain-spanning bend is the slowest Jacobi mode, decaying only ~1/N² per propagation step. Going async/goroutines does NOT fix this on its own: parallelism removes the per-round N-beads factor, but the N² propagation steps form a sequential dependency chain (each step's input is the previous step's output) and can't be parallelized away.

**The escape from cause 2 is a body force, not "no chain": give each bead the node's live transform directly (dependency depth 1, one broadcast hop) instead of averaging with a neighbour.** The escape from cause 1 is: geometry must never be gated on the human clock — a bead that is a goroutine can run its position update on machine time and its animation on human time simultaneously, over disjoint state, because position and animation have separate writers.

**A primitive that fixes both causes has been BUILT and TESTED IN ISOLATION, but is NOT YET WIRED IN.** `nodes/wire/bead_actor.go`'s `Bead` is a chain (render/placeholder) bead as its own goroutine — one `select` with no `default:`, geometry/animation/mode as three structurally distinct channel sets, position computed once per broadcast (`BroadcastChain`, a single close reaching every bead in the group at once) directly from the node's transform, never from a neighbour bead's position. `TestOneHopNotN`/`TestNoDiffusion`/`TestFrameBudgetN1000` (`nodes/wire/bead_actor_test.go`) are the tests that would have caught the original defect, and they pass against the primitive directly — but `chainBeads()` (`nodes/Wiring/chain_beads.go`), the function that actually feeds the live buffer, does not call any of it yet. No editor behaviour has changed. See MODEL.md's "Chain (render/placeholder) bead" bullet, which states the same not-yet-integrated status.

**Key distinction David drew, unchanged:** neighbor-to-neighbor messaging (local, bounded fan-out — it's CSP/actors) is NOT the thing to avoid; GLOBAL scope (one place knowing all nodes, or all-to-all) is. A body-force broadcast from ONE node to ITS OWN chain's beads is local, bounded fan-out, not global scope.

Two fixes from the original detour were kept even before this correction: node body follows local React Flow position (drag fix), and the straight `wireSegment`/`lerp` in PacedWire (the in-flight VALUE bead's transport, unrelated to the chain/render bead — see MODEL.md's split between the two). See [[feedback_uniform_pulse_speed]] and [[feedback_ease_of_fix_is_confounded]].
