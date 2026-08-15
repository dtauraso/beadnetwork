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

**Current (this file, as shipped):** the bead's own radius is the AUTHORED primitive, and
the node lattice's cell is what derives from it — `BeadStepCells` stays fixed at 4 (the
property the second draft got right — the bead lattice is still a coarse SUBLATTICE of the
node lattice, so `iR`'s cell-counted meaning is preserved, only the cell's SIZE
changes), but now the derivation runs the other way:

	BeadRadius              = 4.0                                  (authored — lattice.BeadRadius)
	BeadRingTubeRatio       = 0.12                                  (authored — lattice.BeadRingTubeRatio)
	BeadTorusOuterR         = BeadRadius * (1 + BeadRingTubeRatio) = 4.48
	BeadStepR               = 2 * BeadTorusOuterR                  = 8.96
	BeadStepCells           = 4                                    (unchanged)
	LocalStepR              = BeadStepR / BeadStepCells            = 2.24

Tangency is UNCHANGED by this flip — two beads' tori still touch at exactly half
`BeadStepR`, and a bead's torus is still tangent to a node's torus the same way; only which
of {bead radius, lattice step} is free and which is computed has swapped. What DOES change:
`LocalStepR` grows from 2.0 to 2.24, a node-lattice cell now measures 12% more, and stored
`iR` values are NOT rewritten to compensate — every existing separation, still counted
in the same integer cells, now spans more world units, so the whole graph expands ~12% on
load. That expansion is the accepted, agreed cost of getting the bead's own size right; it
is not compensated for anywhere in this file or its callers.
