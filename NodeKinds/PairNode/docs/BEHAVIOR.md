# PairNode — behavior

[← SPEC.md](../SPEC.md)

The `## View`, `## Ports`, and `## Description` sections stay in SPEC.md because
`NodeKinds/gen/kindscan/spec_md_table.go`'s `readSpecMDLines` reads exactly
`NodeKinds/<Kind>/SPEC.md` and `parseSpecMD` parses only those sections from it. Everything
else the generator does not read lives in the pages below.

- [firing-rule.md](firing-rule.md) — when this node acts: the reactive loop, the In
  arrival, and the three `TiltEditIn` edits (panel click, START, RESET).
- [vector-channel.md](vector-channel.md) — the per-cycle vector channel: the coplanar
  normal, the bottom tilt vector, and what this node sends.
- [tilt-machine.md](tilt-machine.md) — what happens on receiving a vector: the acute
  test, the two-mode state machine, and how a pair decides which mode it runs.
- [separation.md](separation.md) — why a tilt does not move the node.
- [pacing.md](pacing.md) — pacing and clock speed, and the human-edit speed override.
- [third-vector.md](third-vector.md) — the third drawn vector: the last received
  direction.
