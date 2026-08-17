# Beads are the edge — what this deletes, and the open model questions

[← beads-are-the-edge.md](beads-are-the-edge.md)

**These questions were answered by building it; read
[beads-are-the-edge-staging.md](beads-are-the-edge-staging.md) for the answers.** In
particular the per-edge-stream question below was answered NO — the streams kept their
owner, because only the wire's drive moved. Three names here no longer resolve: every test
cited is gone (all test files in this repo were deleted — CLAUDE.md, "There are no tests"),
`bufLayoutBead` is now `bufLayoutEdgeBead` in `Buffer/bufschema/layout_edge_bead.go`, and
`arcLength`/`pulseSpeed` belong to the arc-length model that
[bead-lattice.md](bead-lattice.md) replaced.

## What this deletes

Measured on this branch:

- `nodes/wire/paced_wire.go` — **781 lines**, referenced by **39 files** (16 non-test):
  `Trace/Trace.go`, `nodes/wire/{ports,geometry}.go`, `nodes/input/node.go`,
  `nodes/Wiring/{gesture,topo_spec,build,stream_wiring,edge_mover,loader,node_mover,stdin_reader,input_codec,port_bindings}.go`,
  `Buffer/streamframe/edge_stream_frame.go`.
- `edgeMover` / `StreamKindEdge` touch **35 files**.

This is the largest change since the per-owner buffer split. It is not a rendering tweak.

## The consequence nobody asked for yet: per-edge streams lose their owner

`memory/feedback/architecture/bridge/feedback_no_single_writer_bridge.md` is *one goroutine = one dedicated stream*. The
per-edge fds exist because an edge is a goroutine. Remove the goroutine and the per-edge
stream has no owner — its frames have to fold into the SOURCE node's own stream (the node
that owns that edge, see open question 1).

That reaches further than the network:

- `topology/counts.json` and the fd layout (`Buffer/streamframe/stream_fds.go`, `WIREFOLD_STREAM_FDS`)
- `tools/topology-vscode/src/runCommand.ts`'s fixed fd allocation and its per-edge
  last-frame cache
- the headless harness's spawn (`headless_stream_helpers_test.go`) and
  `TestHeadlessEdgeFdDedicatedStream`, which exists precisely to prove edges reach their own
  fd — that test does not get fixed, it gets **deleted**, because the invariant it guards
  stops being true
- `Buffer/streamframe/edge_stream_frame.go` and the Edge/Bead blocks

**Decide this before writing code**, because it determines whether this is one change or two:
does the edge stream disappear in the same commit as the wire goroutine, or does the edge
keep a (now goroutine-less) stream for a transition? A stream with no goroutine writing it
contradicts the bridge invariant, so the honest answer is probably "same commit," which
makes the commit very large.

## Open model questions

1. ~~Where do the two half-sequences meet?~~ **Settled: there are no halves.** The SOURCE
   node owns the whole sequence, matching ownership the repo already has — an edge is stored
   at `topology/nodes/<source>/edges/<label>.json`, outgoing only, and "carries no `source`
   key: that is the directory it sits in" (`.claude/rules/persistence-ownership.md`). A
   midpoint split was an invention of this document, and a harmful one: the animation is
   directional, so splitting it forces the lit index to HAND OFF from one node's sequence to
   the other's mid-flight — a coordination point in the middle of the one operation that is
   supposed to be purely local. It would also put two owners on one edge's geometry where
   the persistence layout deliberately has one. Constant-time movement does not need it:
   source moves, its whole sequence rides along; target moves, the source re-aims on the
   one-hop neighbour message it already receives. Same asymmetry the stored edge file has.
2. **Who advances the lit percentage, and does delivery still work the same?** Today the wire
   times its own delivery and sends the bead over its out-channel when `ticksToCross`
   elapse. With the wire gone, the **source node** times it and sends to the destination over
   the node-to-node channel it already has. That is a smaller change than it sounds — the
   timing logic moves, it does not disappear.
3. ~~Bead birth/retirement churn.~~ **Settled: count IS length-proportional.** A drag
   changes the count, and those add/drop messages are chain maintenance on the channel — not
   anything traversing (see the opening distinction). Still worth measuring how often a drag
   changes the count, but that is a layout cost to observe, not a reason to prefer a fixed
   count.
4. ~~Uniform pulse speed.~~ **Settled, and it dissolves rather than resolves.** The dwell at
   each bead is the same as today's uniform pulse bead takes to cover that distance. Because
   the count is length-proportional (3), **spacing is constant in world distance**, so a
   constant dwell per bead IS uniform world speed — there is nothing to derive per edge:

       count = len / s            s = constant spacing
       dwell = d ticks per bead   one global constant
       speed = s / d              constant on every edge
       total = count × d = (len/s) × d = len / speed

   That last line is today's `ticksToCross = arcLength / pulseSpeed` exactly, so this
   reproduces current timing rather than approximating it, and the per-edge arc-length
   division disappears. `memory/feedback/architecture/geometry/feedback_uniform_pulse_speed.md` is satisfied structurally
   instead of by computation.

   The in-flight revision rule survives and gets simpler. MODEL.md preserves fractional
   progress `t` across a geometry edit and recomputes remaining ticks from the new arc; here
   `index = t × newCount` and `remaining = (newCount − index) × d`, which is
   `(1−t)·newArc/speed` — the same rule as index arithmetic rather than distance, which is
   the shape `memory/feedback/architecture/geometry/feedback_abc_times_constant_not_rederive.md` asks for.
5. **What happens to the Bead block?** Today `bufLayoutBead` (`Buffer/bufschema/layout.go:56`) is
   `X/Y/Z` world position + `Value`, **one row per live in-flight bead**, fed from
   `KindEdgeBead` events and read by `edge-stream-blocks.ts` off the per-edge stream — Go
   recomputes that absolute position every tick and streams it. Under this model nothing
   produces it: there is no bead with a moving position, only a fixed chain plus a lit index.
   So those three columns lose their writer. Whether the block is deleted or becomes the
   chain's node-local offsets (the Interior block's shape) is open;
   `check-no-dead-buffer-column.sh` forces the decision rather than letting them linger.

## MODEL.md is wrong the moment this lands

MODEL.md is the pinned model, and this contradicts it in at least three places. These edits
belong in the **same commit** as the behaviour, not a follow-up:

- **"Driver"** — "each wire is ALSO a goroutine — not a passive struct another goroutine
  steps" becomes false.
- **"Wire lifecycle"** — describes a lifecycle that no longer has an owner.
- **"Geometry and time"** — "Go owns the bead's PROGRESS ... AND the bead's absolute world
  position" becomes progress plus a node-local offset. `ticksToCross = arcLength /
  pulseSpeed` is no longer computed per edge: constant spacing × constant dwell yields it
  identically (open question 4). The in-flight revision rule ("preserves the bead's
  FRACTIONAL progress `t`") gets *simpler*, not harder: `t` is the only state there is, so a
  geometry edit cannot make it swing.
- **"Allowed vocabulary"** — "wire" as an active goroutine has to be re-defined or retired.
