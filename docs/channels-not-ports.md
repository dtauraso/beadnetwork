# A port is a ROLE, not a place

Agreed model. Narrows MODEL.md's "Input port" bullet; MODEL.md is updated when this lands.

## Most of this is already how the code works

That is the headline, and it only surfaced by reading the real structs after several rounds
of designing something much larger. `SelectRight` embeds `gatecommon.GateNode` and reads
`FromLeft`/`FromRight` — role names that say nothing about `Pulse` or `PulseLeft`. The
binding already happens at LOAD, by name, with no generator:

	Wiring.RegisterBuilder("SelectRight",
	    []Wiring.PortSpec{{Name: "FromLeft", Dir: Wiring.PortIn}, ...},
	    func(a Wiring.BuildArgs) (wire.Node, error) {
	        n.FromLeft = a.In("FromLeft")   // channel bound to role — runtime data
	    })

So "declare roles, bind channels to them at load, keep kinds out of node.go" is not a new
design. It is `PortSpec` plus `a.In(...)`, already shipped. The node structs do not move.

## Ideas considered and rejected

Recorded because each was reached by a plausible route and none should be re-derived.

- **A Go struct field per channel** (`PulseLeftToSelectRight chan Bead`). A field name is an
  identifier, so wiring would live in CODE — dragging an edge in the editor would mean
  regenerating Go, rebuilding, and restarting, with in-flight beads and node-local held
  state reset on every rewire. The whole "spec regenerates Go files, editor auto-refreshes"
  loop existed only to serve this. Dropping it removes the generator, the compile cycle and
  the state reset in one stroke: a channel name is DATA, never a Go symbol.
- **Two mirror-image names per channel** (`ToTimeStartFromInput` on the source side,
  `FromInputToTimeStart` on the target side). One channel with two identifiers cannot be
  found by one grep — search one spelling and you silently miss the other end — and every
  hop between ends is a four-token reversal with a plausible wrong answer. The direction
  words also duplicate what the pair order already says, and a redundant encoding that can
  be written two ways is the shape that confuses its own author.
- **Numbered generic input names** (`in0`, `in1`) with an append-only rule so a number's
  meaning is never repointed. The append-only rule is sound, but it patches a problem
  numbering creates: a number carries no meaning, so nothing but convention pins `in0`'s
  semantics, and repointing it leaves the firing rule compiling and WRONG. A named role
  pins its own meaning — `FromLeft` means the left input whatever feeds it — so the rule is
  unnecessary and roles stay freely re-bindable, which is the point.
- **Removing the port outright.** It does three jobs; two die, one is load-bearing.

## The three jobs

| Job | Verdict |
|---|---|
| Geometry — ring anchor, port radius, `port ∈ torus` | **Dies.** Carries no control, and it is what produces the end-bead gap. |
| Channel identity — "which connection is this" | **Dies.** The sender answers it: `PulseLeftToSelectRight`. |
| The node's declared input ROLE | **Survives.** The firing rule needs it and nothing else can supply it. |

Port-as-geometry never carried control at all. Across the whole bead path — placement,
drive, timing, lighting — the port appears at exactly ONE control point: delivery, where
the wire's out-channel was the destination's input-port channel.

## What actually changes

- **Ports lose their geometry.** No ring anchor, no port radius, no `port ∈ torus` gesture,
  no port rows or port geometry columns in the buffer, no `nodes/<id>/inputs/` or
  `outputs/`. Edges attach at node surfaces on the bead lattice (docs/bead-lattice.md,
  which lands with task/arc-from-local-polar). This is the entire fix for the end-bead gap: the
  chain measured node-torus to node-torus while the port sat proud of (or inside) that
  surface, so the first and last bead were off by the port's own radius while interior
  spacing stayed correct — which is exactly how the defect presented.
- **The channel is named `<FromStruct>To<ToStruct>`, as DATA** — the edge label, the edge
  filename, the wire's identity, the log line, the sentence in an agent's report. One
  string per channel, identical from either end, found by one grep, direction given by pair
  order. Never a Go identifier.
- **The edge file loses `sourceHandle`.** Today node 1's edge to node 2 is stored under
  `edges/` as `1To2.json`, holding a `sourceHandle`, a `target`, a `targetHandle` and a
  `label`. It becomes a file named for the channel — `InputToTimeStart` — holding only
  `target` and `targetRole`. The sender is the directory the file sits in;
  storing it again would be a second copy free to drift
  (`.claude/rules/persistence-ownership.md`).
- **`Target + "." + TargetRole` stays** as the uniqueness key. One edge per role is still
  the rule and still what we want.
- **Wrong role names are fixed** (below). Nothing else in `node.go` moves.

A two-struct channel name is unique only while no two nodes share a kind. Checked against
the live graph before adopting it: all 9 nodes are distinct kinds and all 10 edges have a
unique kind pair. A second `Pulse` would make `PulseToSelectRight` name two channels — and
it fails LOUDLY, because both edges want the same path under `nodes/<source>/edges/`.

## Role naming rules

A role says WHAT THIS INPUT DOES to the node declaring it. It is read by that node's own
firing rule and by nobody else, so it must be legible without knowing the graph.

1. **Never name a kind.** `FromPrevTimeNode` and `FromInput` name the `Time` and `input`
   kinds; the moment a different kind is wired there the name lies — and free rewiring is
   the whole point.
2. **Never number.** `Out2` says nothing about what the second output is for — the same
   defect as `in1`, rejected above.
3. **Say the job.** `FromLeft` / `FromRight` / `ToPassed` are already right: they describe
   the node's own view of its inputs and outputs.
4. **A sole input may be `In`.** With exactly one, there is nothing to distinguish.
