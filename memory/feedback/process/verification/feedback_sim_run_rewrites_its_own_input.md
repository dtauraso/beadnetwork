---
name: feedback-sim-run-rewrites-its-own-input
description: Byte-comparing two sim builds requires ONE input snapshot copied once; re-copying topology/ per run feeds the binaries different inputs and manufactures fake diffs
metadata:
  type: feedback
---

Running the sim rewrites the `topology/` it read (float noise in edge JSONs —
see [[feedback-verify-run-rewrites-edge-json]]). So a comparison harness that does
`cp -R topology $RUN/` fresh for each binary feeds each one a DIFFERENT input,
because the earlier run already mutated the source.

**Why:** this manufactures diffs that look exactly like a real regression, and
manufactures agreement too. Chasing one cost a full bisect: `node.bin` bytes at
offsets 270 and 355 appeared to differ, and the offsets MOVED between runs, which
reads as timing noise in the frame. Neither was true — same-snapshot runs of the
baseline, the previous commit and the new commit are byte-identical.

**How to apply:** copy the scene dirs ONCE into a snapshot, then copy that
snapshot per run. Always run the control (same binary twice) from the same
snapshot, and treat a differing byte OFFSET between control and change as a sign
the harness is wrong before concluding the code is. Related:
[[feedback-headless-repro-verifies-persistence]], [[feedback-debug-data-before-theory]].
