# The bead lattice — derivation history

[← bead-lattice.md](bead-lattice.md)

## The lattice is derived, not the bead

This flipped once, and the flip is worth recording in full because both directions look
reasonable in isolation.

**First draft (rejected then, tried, then also rejected):** derive `LocalStepR` itself
(8.96) to match a torus-touching bead of radius 4.0, so the node lattice's own cell is
sized by the bead. Rejected before it was built: that would have re-interpreted every
stored `iR` against a coarser step and shifted every authored position on load.

**Second draft (built, shipped, then reverted — this is the one that actually ran):** make
`LocalStepR = 2.0` (a small, hand-picked node-lattice cell) the primitive, fix
`BeadStepCells = 4` as a constant, and let the bead step — and therefore the bead's own
visible size — fall out of that:

	BeadStepCells   = 4
	BeadStepR       = BeadStepCells * LocalStepR = 4 * 2.0 = 8.0
	BeadTorusOuterR = BeadStepR / 2 = 4.0
	ShadingParamBeadRadius = BeadTorusOuterR / (1 + ShadingParamBeadRingTubeRatio)
	                       = 4.0 / 1.12 = 3.5714...

This kept `iR`'s meaning intact (still counted in 2.0-unit cells; exact tangency only
needed a separation snapped to every 4th cell, a shift of at most 4 world units on first
load, not a re-interpretation) — solving the problem the first draft was rejected for. But
it made the bead's SIZE a dependent variable: the visible bead SHRANK to 3.5714..., about
11% smaller than the 4.0 David had actually picked. That was visibly wrong once it shipped
and was looked at, not a theoretical concern — a bead is the one thing a person looks
directly at, and "the lattice happens to imply this exact rendered size" is not a good
enough reason for that size to be someone else's decision.

**Third draft (built, shipped, then superseded):** the bead's own radius became the AUTHORED
primitive and the node lattice's cell derived from it, with `BeadStepCells` still fixed at 4
so the bead lattice stayed a coarse SUBLATTICE of the node lattice and `iR`'s cell-counted
meaning was preserved — only the cell's SIZE changed:

	BeadStepCells           = 4
	LocalStepR              = BeadStepR / BeadStepCells            = 2.24

**Current:** the sublattice is gone. There is no `BeadStepCells` and no `LocalStepR` — the
radial grid IS the bead lattice, one index step per bead step, and that identity is asserted
at load:

	BeadRadius              = 4.0                                   (authored — lattice.BeadRadius)
	BeadRingTubeRatio       = 0.12                                  (authored — lattice.BeadRingTubeRatio)
	BeadTorusOuterR         = BeadRadius * (1 + BeadRingTubeRatio) = 4.48
	BeadStepR               = 2 * BeadTorusOuterR                  = 8.96
	constantR               = BeadStepR                            = 8.96

`constants.json`'s `constantR` carries that 8.96, and `loadspec.LoadSceneConstants` fails
the load by path if it disagrees with `lattice.BeadStepR` — the radial grid is required to
match the bead spacing rather than merely happening to. A node's radial position is
`indexR * constantR` (`.claude/rules/persistence-ownership.md`), so an authored separation
is a whole number of bead steps by construction, with no 4-cell conversion left in between.

Tangency is UNCHANGED by every one of these flips — two beads' tori always touch at exactly
half `BeadStepR`, and a bead's torus is tangent to a node's torus the same way; what moved
each time was only which of {bead radius, lattice step} is free and which is computed, and
how many node cells make a bead step. The third draft's flip cost a one-time ~12% expansion
of the whole graph (`LocalStepR` 2.0 → 2.24 with stored indices not rewritten to
compensate), accepted to get the bead's own size right; collapsing the sublattice afterwards
made the radial index step 8.96 outright, so a stored index counts bead steps directly.
