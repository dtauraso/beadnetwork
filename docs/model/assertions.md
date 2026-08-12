# Model — assertions and allowed vocabulary

[← MODEL.md](../../MODEL.md)

## Assertions

A `panic` in `nodes/`, `Buffer/`, or `Trace/` is an **assertion**, not error handling. It
fires only via a code bug — never via ordinary traffic, malformed input, or load. Input the
network cannot trust is rejected at parse (`validateNoFanIn`, `validateSpec`); by the time a
value reaches a goroutine's own state, a violated bound means the code is wrong.

So an assertion states an invariant the structure is supposed to guarantee, and the right
response to one firing is to fix the code, not to widen the bound.

**Panic rather than drop, grow, or block.** Dropping hides the broken drain the bound exists
to catch. Growing defers the same crash to a worse place. Blocking couples this goroutine's
pacing to another's, which is the coupling the whole model exists to avoid.

**The message is the whole value.** It is read exactly once, by whoever is debugging, and it
is the only context they get. It must:

1. **Open with a site tag** — the detecting function, method, or subsystem, then a colon —
   so the message greps back to its source: `paced_wire: `, `NodeMover(%s): `,
   `BuildEdgeStreamFrame: `.
2. **Name the invariant and the actual values**, not a category. `pending exceeded %d events
   on wire -> %s.%s`, not `limit exceeded`.
3. **Name the mechanism that should have prevented it.** This is what turns a crash into a
   diagnosis: *"the per-cycle drain (edgeMover.writeStreamFrame -> DrainPendingEvents) is not
   running"*, or `allocateWires`' *"validateNoFanIn should have rejected this fan-in at
   parse"* — which names the earlier gate that let it through.

**No `recover()` in the network.** Swallowing an assertion converts a loud, located failure
into a silent wrong answer.

Guard: `tools/network/quality/check-panic-message.sh` (site tag + substance + no `recover()`). It enforces
the shape, not the content — (3) is the part only a human can write, and the part that pays.

## Allowed vocabulary

- bead, in-flight, held (node-local) state
- channel, input port, output port (one edge per input port — no fan-in)
- arc length, pulse speed (world-units per tick), ticks-to-cross,
  tick-count processing window
- tick, human-speed clock (the one system monotonic clock scaled to ticks
  at human speed), scale, `SleepCycle`, `Tick`
- node receives, node holds, node fires, wire advances, wire delivers,
  wire emits position
