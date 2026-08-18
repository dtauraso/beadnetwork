---
paths:
  - nodes/Wiring/**
  - tools/topology-vscode/src/runCommand.ts
  - topology/**
---

# Persistence — each owner writes its own file

The network's rule applied to disk: per-owner, no shared mutable state, no coordinator.
This is doctrine, not a plan — it describes the tree as it is.

## The layout is an adjacency list

Everything about a node lives under its own directory. There is no top-level `edges/`.

```
topology/
├── counts/nodes.json  counts/edges.json   one integer each — nodes = ROW COUNT
│                                          (largest node id), not a live-node count
├── constants.json                        {"constantR", "constantPhi", "constantTheta"} — scene-level, read once
├── nodes/<id>/
│   ├── base.json                          type/id/gate/drag-rule + the BASE position — TRACKED, see below
│   ├── drag/self.json                 the node's accumulated position DELTA — GITIGNORED, see below
│   ├── data.json  local-polars.json
│   ├── edges/<label>.json                OUTGOING only — wiring + the BASE geometry delta — TRACKED
│   ├── drag/edges/<label>.json            that edge's accumulated geometry DELTA — GITIGNORED
│   └── (no *.geom.json — folded into <label>.json)
└── view/
    └── camera.json  overlays.json  panels.json  sphere.json  scene.json
```

The panels.json file under view/ holds the overlays popover's disclosure open/closed
state (`viewstate.PanelState`, `tools/topology-vscode/OverlaysDropdown/panels_persist.go`) — its
own file, deliberately separate from the overlays.json overlay-visibility file: a panel's
open/closed state is not an overlay visibility flag, even though the two are persisted,
streamed, and edited the same way.

**`topology/` is one of several sibling SCENES**, not the only tree. `nodes/Wiring/scene/scene.go`'s
`Scenes` names each sibling directory (today: `topology/`, `topology-pair/`) resolved
relative to the ANCHOR's parent — the `-topology` flag the extension host launches with is
the fixed anchor, and which sibling directory actually loads is resolved from it
(`ResolveScenePath`). Each sibling is a COMPLETE, independently loadable tree with its own
`counts/` and `view/`, laid out exactly as above. `topology/view/scene.json`
(`{"selected": "<tab name>"}`) is the ONE piece of state that lives at the ANCHOR rather than
inside whichever scene is loaded — it has to, since it is what says which sibling to load, and
a selection stored inside scene B would be unreachable while scene A is showing. Switching
tabs writes this file and ends the Go process; the extension host's already-looping runner
respawns it, and the respawn re-reads the selection and loads the other tree.

An edge is stored under its **source** node and carries no `source` key — that is the
directory it sits in, and storing it too would be a second copy free to drift.

A node and its outgoing edges are one unit: deleting `nodes/5/` takes node 5's edges with
it, instead of leaving a dangling file in a sibling tree.

**In-edges are deliberately not local, and that costs nothing.** Nothing ever asks "what
points at node 9". Every use of an edge's target is built during a single full pass over all
edges (`buildEdgeMaps`' `inbound`, `loader_layout`'s `neighbors`). Do NOT "fix" this by also
recording the edge under the target — that reintroduces the duplication the layout removes.

## Base composed with drag; the delta is gitignored, not skip-worktree'd

**`base.json` and `<label>.json` are TRACKED and never written by the running sim.** They
hold, respectively, a node's `type`/`id`/`gate`/`drag`-rule plus its BASE position, and an
edge's wiring (`target`/`sourceHandle`/`targetHandle`/`kind`) plus its BASE geometry delta.
**Both the base position and the base delta are stored ONLY as an index** —
`base.json`'s `indexPhi`/`indexTheta`/`indexR` (absolute, folded through
`polarindex.Canonical`) and `<label>.json`'s `deltaIndexR`/`deltaIndexPhi`/`deltaIndexTheta`
(a relative integer step count, NEVER folded through `Canonical` — a delta is not a
position). There is no `scenePolarR`/`scenePolarPhi`/`scenePolarTheta` on a node or
`deltaPolarR`/`deltaPolarPhi`/`deltaPolarTheta` on an edge anywhere on disk any more; the
index is the sole authored quantity and `polarindex.ToPolar(idx, sc)` is the only multiply,
run at load. **A drag is stored the same way — an index delta,
`indexPhi`/`indexTheta`/`indexR` under `drag/` — and the drag value IS index × constant,
never a second stored continuous copy.** `drag/`
`self.json` and `drag/edges/` `<label>.json` are GITIGNORED and are the only files a drag
writes; both hold exactly that one triple (the offset from `base.json`'s own
`indexPhi`/`indexTheta`/`indexR`, e.g. a base `indexR` of 25 plus a 4-step drag stores `iR:
4`, not 29). There is no `dragPolarR`/`dragPolarPhi`/`dragPolarTheta` field anywhere in
either file — `polarindex.ToPolar(idx, sc)` is the ONLY place the multiply happens, and it
runs at load and at the moment a drag is measured, never stored as its own field. A
sub-step drag rounds to the nearest index (`polarindex.MeasureScalar`) and a drag that
rounds to zero is correct, not lost precision to recover. Neither file carries
`constantPhi`/`constantTheta`/`constantR` — nor does `base.json` any more; those moved to
`constants.json` at the scene root (read once per load, fails loudly by path on
missing/malformed input, and asserts `constantR == lattice.BeadStepR` — the radial grid
must match the bead lattice's own step size). The loaded triple (`polarindex.SceneConstants`)
is passed explicitly into every consumer, never a per-instance field or a package-level
global.

**A (base) and B (drag) are held SEPARATELY in the running program, never summed into a
stored value.** The quantized index rides this split, in `owners.Quant`: `base` (A, loaded
once from `base.json`'s `indexPhi`/`indexTheta`/`indexR`) and `drag` (B, set by
`CommitQuantOffset` as `polarindex.Delta(measured, base)`, never reconstructed by
subtracting later). `Quant.Composed()` returns `polarindex.Compose(base, drag)`, computed at
each call site and held nowhere else; `nodegeom.ScenePolarOf`/`polar.Compose` derive the
continuous on-screen position from that composed index the same way, at each call site that
needs it, held in no field that outlives that call. `persistQuantOffset` writes `drag`
straight to `drag/` `self.json`'s `indexPhi`/`indexTheta`/`indexR`.

The same split applies to an edge's target vector. `owners.Deltas` holds `baseTo` (A, seeded
once at build from each edge's `<label>.json`, outgoing straight and incoming negated, never
mutated after) and `dragTo` (B, the only thing `ShiftSelfBy`/`ShiftOtherBy` mutate).
`DeltaTo` returns the derived compose; `DragDeltaTo` returns B as a `polar.Polar`.
`OutEdges.persistDelta` measures that into an index (`polarindex.MeasureScalar`) and writes
the index straight to `drag/edges/` `<label>.json` — the same shape as the node path, no
continuous field survives the write.

THE DELTA (a node's standing vector to a neighbor, in `<label>.json`) and THE DRAG (the
user's accumulated offset, in `drag/`) are different quantities. They are never summed into
a stored value and never substituted for one another; they combine only at the point
something is drawn, and persist always writes the DRAG, never the composed value.

On load, `loadspec.ApplyDragOverlay` reads `base.json`/`<label>.json` into the BASE fields
and `drag/` `self.json`/`drag/edges/` `<label>.json` into separate DRAG fields on the same
spec struct (`DragScenePolarR/Phi/Theta` on a node, `DragDeltaPolarR/Phi/Theta` on an
edge) — it composes NOTHING; a node or edge with no `drag/` file simply carries a zero drag
value forward. `polar.Compose`/`polar.Between` are untouched by this change; this is the
existing componentwise op, just applied at the render/persist boundary instead of at load.

Do NOT try to keep geometry in one tracked file and hide the drag write with
`git update-index --skip-worktree`. That was tried and abandoned: skip-worktree suppresses
`git status`/`git diff` reporting but NOT `git checkout`'s refusal to switch branches when a
marked file differs from the index — so the first branch switch after a drag forces the drag
output into a commit to unblock the checkout, the exact outcome the split exists to avoid. A
gitignored subdirectory has nothing in the index for `checkout` to compare against.

Edge WIRING (`target`/`sourceHandle`/`targetHandle`/`kind`/`label`) is authored data and
keeps writing to the tracked `<label>.json` (`edgefile.WriteEdgeFile`) whenever it
legitimately changes. Only the geometry DELTA moved to `drag/`.

A drag touches more than the node you grabbed: the dragged node shifts its own vectors, and
each neighbour is told how far it moved and shifts its own — so a neighbour's `drag/`
`self.json` and each of ITS outgoing edges' `drag/edges/` `<label>.json` also get
written as part of that same drag, same as before.

Consequences to keep in mind:

- A fresh clone has the tracked base values only (no `drag/` directory yet — that only
  appears on the first LOCAL drag write), and loads a complete, self-consistent scene from
  them with zero deltas.
- Nothing promotes a drag's accumulated delta back into the base for a future clone; that
  would need a deliberate migration and commit (see the migration note below). Whatever is
  on disk right now under `base.json`/`<label>.json` is the base the NEXT clone gets.
- There is no more `*.geom.json` to skip when scanning `nodes/*/edges/*.json` — every file
  under an `edges/` dir is a real edge (`loadNodeEdges`, `check-no-fan-in.sh`).
- The existing `topology/`/`topology-pair/` trees were migrated by treating their
  already-committed values AS the base and starting with empty `drag/` dirs (zero deltas) —
  no attempt was made to reconstruct a pre-drag historical base, since none is recoverable.

## The owner writes, and owns the path

- A node writes its own `position/local-polars` (path construction in
  `dragfile/drag_file.go`). There is no longer a separate `inputs/`/`outputs/` port-geometry
  file — port geometry was removed with the port model (edges attach on the bead lattice,
  `nodes/bead/lattice/bead_lattice.go`); this bullet used to list it as a second thing the mover writes.
- The **SOURCE NODE** owns `nodes/<source>/drag/edges/<label>.json`, and writes it from
  `nodes/Wiring/nodeactor/owners/out_edges.go`'s `persistDelta` — the same pass that derives that edge's geometry,
  since the node is what holds the vector being stored. The write is gated on the vector
  actually changing: derivation runs every tick, so an ungated write would rewrite the file
  every tick. (This bullet used to claim "an `edgeMover` owns" it and "no Go writer exists
  yet — edges are editor-authored". Both were wrong: `edge_delta_persist.go` had been the
  writer all along, and it rewrote every edge file on startup, which is why a plain run
  used to dirty the whole tree.)
- Scene-level state (camera, overlays, sphere) is genuinely singular and belongs to the
  view-owner goroutine (`RunStdinReader`).

**One owner writes inside `nodes/<id>/`: that node.** Its own position/local-polars, and the
file of every edge leaving it.

This REVERSES what this section used to say. The old text had two owners per directory —
the node's mover and the edgeMover of each edge leaving it — and warned that "routing edge
writes through the source node's mover would make one goroutine write another's state on
request". That warning assumed the edge's delta was the EDGE's state. It is not: the vector
from a node to its neighbour is the node's own state, held in its `owners.Deltas` and
maintained by composing each neighbour's reported move into it. The edgeMover was computing
a second copy of that vector from mirrored absolute positions and persisting the copy. So
the node writing the file is a goroutine writing what it already owns — not a request, and
one fewer representation of the same quantity.

Guards: `tools/network/persist/check-persist-write-ownership.sh` (who may write which path pattern),
`tools/network/persist/check-scene-path-resolution.sh` (who may construct a `nodes/` path).

## A topology is a directory tree, always

The monolithic single-file form is gone. `readSpec` rejects a non-directory. Node ids ARE
numbers — they are strings only because they are directory names — and **ROW ID = NODE ID -
1**: `loadTree` parses each node directory name to an int (`strconv.Atoi`) and that int,
minus one, IS the node's buffer row directly. There is no ordering step left — a row is
declared by the id, never derived by sorting or by position in a list, so there is nothing
for a "10" vs "2" comparison to get wrong any more. A node directory name that fails to
parse as a number, an id below 1 (ids are 1-based), or a duplicate parsed id is each a LOAD
ERROR, loud and naming the offending directory, like a missing/malformed count file
below — never a silent fallback. The id itself stays a `string` everywhere downstream (it is
a map key across the codebase); only the row derivation parses it to an int.

The row space (`topoSpec.RowCount`) is sized by the **largest id found**, not by the node
count: deleting `nodes/5/` leaves row 4 empty rather than shifting nodes 6.. down to fill
it — that shift is precisely the silent renaming this model exists to remove (node 6's
geometry used to arrive on node 5's row the moment node 5 was deleted). Every consumer of
the node-row space (the buffer packer, the per-node fd/stream wiring, the row-identity
lookup tables) must tolerate an empty row.

Per-node `edges/` file order is a PLAIN `sort.Strings` — those names are labels, not
numbers, and must stay lexicographic. Port order is alphabetical by port name — neither is
authored order, because a tree has no array.

The pre-split scene sidecar (a single `scene.json` under the view dir, holding what is now
split across camera/overlays/sphere) and its best-effort read fallback (`sceneCameraPath`/
`sceneJSONPath`) were REMOVED — no such file existed anywhere in this repo, and nothing
wrote it once the one-file-per-writer split landed. A topology directory holding only that
legacy sidecar now loses its camera pose, overlay flags and scene sphere on load (falls back
to defaults) instead of migrating them forward. That was a DIFFERENT legacy from the
monolithic topology form covered above.

## Counts are stored, never re-derived

`topology/counts/nodes.json` and `topology/counts/edges.json` exist because the extension host SPAWNS Go, and Node's `spawn()` takes the
stdio array up front — with one dedicated pipe per emitting goroutine, the pipe count must
be known before the child exists, and Go cannot answer because Go is not running yet.

**`nodes` means the ROW COUNT** (the largest node id in the tree, `topoSpec.RowCount`),
**not** how many node directories exist. Under ROW ID = NODE ID - 1 a gap in the id space
(a deleted node) still needs a dedicated fd allocated for its now-empty row — the row space
doesn't shrink just because one row went empty — so sizing the stdio array from a live-node
count would under-allocate the moment any id is missing from the middle of the range.
`edges` is unaffected by this — edges have no id space to gap, so it stays a plain count.

**Nobody re-derives it — not TS, and not Go.** Go's own load (`loadTree`) walks `nodes/`
and computes its own `RowCount` independently, from the SAME rule (largest id), but that is
LOADING (reading each node's data to build the graph and its own row space), not reading
those files — Go never opens them at all; it exists solely so something can size the
stdio array BEFORE Go is running to answer. Correctness is single-writer: the one operation
that creates, deletes, or renumbers a node, or creates/deletes an edge, updates
both count files. Nothing else writes them.

A missing or malformed count file must fail LOUDLY. Returning 0 allocates no dedicated
streams and degrades the bridge invisibly — the behaviour the old `countEdges` had. The
extension host reader (`tools/topology-vscode/src/runCommand.ts`'s `readCounts`) and the Go
headless test harness (`headless_stream_helpers_test.go`) must fail the same way if the
stored `nodes` value disagrees with the tree's own largest id — not just on a missing file.

No Go writer exists today, so the file is hand-maintained alongside the tree. The headless
harness sizes its spawn from this same file and fails if it disagrees with the tree's own
computed row count, which is the only drift check that does not re-derive at runtime.
