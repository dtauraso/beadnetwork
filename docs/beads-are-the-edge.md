# Beads are the edge — plan

The edge stops being a thing that owns geometry and timing. A node owns a **sequence of
placeholder beads** toward each neighbour; the sequence *is* what an edge looks like. The
wire goroutine is removed and its animation logic moves into the node. Traversal is no
longer a bead moving along a wire — it is successive placeholder beads **lighting up** at
their fixed percentages.

Model as stated (three parts, all from David):

1. **Beads replace the edge.** `len(edge) × x%` placeholder beads, arranged so the sequence
   looks like an edge. The bead sequence is part of the node, so each node is "next to" its
   neighbours. Moving a node is constant time.
2. **Nodes coordinate over the channels they already have.** Node-to-node messages say which
   node makes another bead and which node adjusts to an x%-farther distance. There is no
   edge entity mediating it.
3. **The wire goroutine is removed.** Its animation logic is owned by the node.

## Why this is not the rejected chain model

`memory/project_wire_is_straight_line_not_chain.md` reverted a bead-chain wire and says not
to re-propose it. That rejection was specifically about **neighbour-midpoint relaxation** —
a bead's position depending on its neighbours' positions, making straightness a diffusion
process that follows a drag in O(N²) (measured ~1.5s at N≈40). The same memory names the
one design that escapes it:

> The only escape is giving each bead the two anchors directly (lerp on the anchor line →
> dependency depth 1), which means broadcasting endpoint data to every bead.

A bead at a **fixed percentage** is exactly that: dependency depth 1, no follow, no
relaxation, no born-to-hold-spacing settling. The ban does not apply, and the escape it
names is what this is.

**The line not to cross:** no bead position may depend on another bead's position. The
moment spacing is maintained by looking at the bead next door, this becomes the reverted
design. Positions come from the node (and its neighbour's), never from adjacent beads.

## Why moving a node is constant time

Only if bead offsets are **node-local**, which the buffer already does elsewhere. The
Interior block stores `OX/OY/OZ` as "the Go-owned NODE-LOCAL slot offset (relative to the
node center — the renderer adds the node center to get the world position)"
(`Buffer/layout.go`). Go owns the offsets, the renderer does one add. That is not TS owning
positions and it ships today.

So a node moving rewrites its own centre and nothing else — its whole bead sequence rides
along for free. What still has to happen is **re-aiming when the neighbour moves**, and that
is the existing one-hop neighbour message (`neighborSetC`), not a new mechanism.

Without node-local offsets there is no constant time: Go owns absolute bead positions today
(MODEL.md "Geometry and time"), so a move would cost `degree × N` position writes — worse
than the status quo, not better. **Node-local offsets are load-bearing, not an optimisation.**

## What this deletes

Measured on this branch:

- `nodes/wire/paced_wire.go` — **781 lines**, referenced by **39 files** (16 non-test):
  `Trace/Trace.go`, `nodes/wire/{ports,geometry}.go`, `nodes/input/node.go`,
  `nodes/Wiring/{gesture,topo_spec,build,stream_wiring,edge_mover,loader,node_mover,stdin_reader,input_codec,port_bindings}.go`,
  `Buffer/edge_stream_frame.go`.
- `edgeMover` / `StreamKindEdge` touch **35 files**.

This is the largest change since the per-owner buffer split. It is not a rendering tweak.

## The consequence nobody asked for yet: per-edge streams lose their owner

`memory/feedback_no_single_writer_bridge.md` is *one goroutine = one dedicated stream*. The
per-edge fds exist because an edge is a goroutine. Remove the goroutine and the per-edge
stream has no owner — its frames have to fold into the two endpoint nodes' own streams.

That reaches further than the network:

- `topology/counts.json` and the fd layout (`Buffer/stream_fds.go`, `WIREFOLD_STREAM_FDS`)
- `tools/topology-vscode/src/runCommand.ts`'s fixed fd allocation and its per-edge
  last-frame cache
- the headless harness's spawn (`headless_stream_helpers_test.go`) and
  `TestHeadlessEdgeFdDedicatedStream`, which exists precisely to prove edges reach their own
  fd — that test does not get fixed, it gets **deleted**, because the invariant it guards
  stops being true
- `Buffer/edge_stream_frame.go` and the Edge/Bead blocks

**Decide this before writing code**, because it determines whether this is one change or two:
does the edge stream disappear in the same commit as the wire goroutine, or does the edge
keep a (now goroutine-less) stream for a transition? A stream with no goroutine writing it
contradicts the bridge invariant, so the honest answer is probably "same commit," which
makes the commit very large.

## Open model questions

1. **Where do the two half-sequences meet?** If each node owns beads out to the midpoint,
   every edge is two node-local sequences and no shared geometry exists anywhere — cleanest,
   and it makes "each node is next to the other nodes" literally true. If one node owns the
   whole run, the far node moving re-aims someone else's beads, which is a worse ownership
   story. Assume midpoint unless David says otherwise.
2. **Who advances the lit percentage, and does delivery still work the same?** Today the wire
   times its own delivery and sends the bead over its out-channel when `ticksToCross`
   elapse. With the wire gone, the **source node** times it and sends to the destination over
   the node-to-node channel it already has. That is a smaller change than it sounds — the
   timing logic moves, it does not disappear.
3. **Bead birth/retirement churn.** Count is `f(length)`, so a drag continuously creates and
   retires beads. The reverted chain model also had "born/retired to hold spacing" — that is
   not itself why it was rejected, but it is worth measuring the churn rate during a drag
   before committing to length-proportional counts. A fixed count per edge with variable
   spacing avoids the churn entirely and should be priced against it.
4. **Uniform pulse speed.** `memory/feedback_uniform_pulse_speed.md` says speed is uniform
   across all wires. The lit-percentage advance rate must therefore be derived from arc
   length (`ticksToCross = arcLength / pulseSpeed`), not be a per-edge constant — otherwise a
   long edge lights up at the same rate as a short one and world-speed stops being uniform.
5. **What happens to the Bead block?** If beads become node-local offsets plus a lit index,
   the Bead block's absolute `X/Y/Z` are dead and it collapses toward the Interior block's
   shape. `check-no-dead-buffer-column.sh` will force that decision rather than let the old
   columns linger.

## MODEL.md is wrong the moment this lands

MODEL.md is the pinned model, and this contradicts it in at least three places. These edits
belong in the **same commit** as the behaviour, not a follow-up:

- **"Driver"** — "each wire is ALSO a goroutine — not a passive struct another goroutine
  steps" becomes false.
- **"Wire lifecycle"** — describes a lifecycle that no longer has an owner.
- **"Geometry and time"** — "Go owns the bead's PROGRESS ... AND the bead's absolute world
  position" becomes progress plus a node-local offset. The in-flight revision rule
  ("preserves the bead's FRACTIONAL progress `t`") gets *simpler*, not harder: `t` is the
  only state there is, so a geometry edit cannot make it swing.
- **"Allowed vocabulary"** — "wire" as an active goroutine has to be re-defined or retired.

## Staging

The size argues against one commit, but the bridge invariant argues against a transition
state. Proposed order, each verified with `bash scripts/stop-checks.sh`:

1. **Node-local bead offsets, alongside the existing wire.** Add the node-owned sequence and
   stream it on the node's own frame, with the wire goroutine still running and still
   authoritative. Nothing renders from it yet. This proves the offsets and the constant-time
   move in isolation, and is revertible on its own.
2. **Render from the sequence; stop rendering the moving bead.** The visual switch. The wire
   goroutine still runs and still owns delivery timing.
3. **Move delivery timing into the source node; delete the wire goroutine**, the per-edge
   streams, `Buffer/edge_stream_frame.go`, the Edge/Bead blocks that die with them, and
   `TestHeadlessEdgeFdDedicatedStream`. Update MODEL.md in this commit.
4. **Update `counts.json` / fd layout / `runCommand.ts` / headless harness** for the removed
   edge fds — possibly unavoidable inside step 3 rather than after it.

Step 1 is the single concrete next step and the only one worth starting before the open
questions above are answered.

## Risk worth naming

Steps 1 and 2 leave the system with two representations of the same thing (a moving bead and
a lit sequence) and one of them authoritative. That is a drift shape this repo has been
bitten by before — `memory/feedback_reflect_dont_create_store.md`, and the whole reason
`check-no-webview-state.sh` exists. Keep the wire authoritative and the sequence purely
derived until step 3 flips it in one move; do not let both be true at once.
