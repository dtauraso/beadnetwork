---
paths:
  - Categories/Scene/Topology/**
  - Start/extension/runCommand.ts
  - topology/**
---

# Persistence — each owner writes its own file

The network's rule applied to disk: per-owner, no shared mutable state, no coordinator.
This is doctrine, not a plan — it describes the tree as it is.

## The layout is an adjacency list

Everything about a node lives under its own directory. There is no top-level `edges/`.

```
topology/
├── counts/nodes.bin  counts/edges.bin   one integer each — nodes = ROW COUNT
│                                          (largest node id), not a live-node count
├── constants/                             constant-r, max-index-phi, max-index-theta — read once
├── nodes/<id>/
│   ├── base/                               one value per file: type/gate/index + drag-rule/ and self-rule/ — TRACKED
│   ├── drag/                               index-phi/-theta/-r — the node's position DELTA — GITIGNORED
│   ├── data/ links/<from>To<to>/           data: init/<n>, repeat, state/<k>, send-rules/<port>; links: a links node's OUTGOING links (start, Δ, increment, lock, pick) — both TRACKED; a links node's edits write the whole link set to the GITIGNORED drag/links/, which REPLACES links/ at load, as drag/at-*.bin and drag/s.bin replace base/'s (Links/link_files.go, nodefile_place.go)
│   ├── edges/<label>/                     OUTGOING only, one value per file: wiring + delta — TRACKED
│   └── drag/edges/<label>/                that edge's accumulated geometry DELTA — GITIGNORED
└── view/
    └── camera/ overlays/ panels/ sphere/ scene/ speed.bin — one value per file
```

**One primitive per file, and the DIRECTORY HIERARCHY IS THE NESTED DICT.** A directory is an
object, a file is a leaf holding one primitive, numbered files are an ordered array
indexed by filename — sorted NUMERICALLY, since `10` follows `9` but sorts after `1`. A
non-integer name in an array directory is a load error, never a skipped element: skipping one
silently shortens the array and shifts the rest.

The `view/panels/` directory holds the overlays popover's disclosure open/closed
state (`Panel.PanelState`, `Categories/Chrome/Panels/Panel/panels_persist.go`) — its
own directory, deliberately separate from the `view/overlays/` visibility flags: a panel's
open/closed state is not an overlay visibility flag, even though the two are persisted,
streamed, and edited the same way.

**`topology/` is one of several sibling SCENES**, not the only tree. `Categories/Scene/Scenes/scenes.go`'s
`Scenes` names each sibling directory (today: `topology/`, `topology-pair/`) resolved
relative to the ANCHOR's parent — the `-topology` flag the extension host launches with is
the fixed anchor, and which sibling directory actually loads is resolved from it
(`ResolveScenePath`). Each sibling is a COMPLETE, independently loadable tree with its own
`counts/` and `view/`, laid out exactly as above. `topology/view/scene/selected.bin`
(the selected tab name, one string) is the ONE piece of state that lives at the ANCHOR rather than
inside whichever scene is loaded — it has to, since it is what says which sibling to load, and
a selection stored inside scene B would be unreachable while scene A is showing. Switching
tabs writes this file and opens the other tree in the same Go process; the file is read at
process start to choose the first scene to open.

An edge is stored under its **source** node and carries no `source` key — that is the
directory it sits in, and storing it too would be a second copy free to drift.

A node and its outgoing edges are one unit: deleting `nodes/5/` takes node 5's edges with
it, instead of leaving a dangling file in a sibling tree.

**In-edges are deliberately not local, and that costs nothing.** Nothing ever asks "what
points at node 9". Every use of an edge's target is built during a single full pass over all
edges. Do NOT "fix" this by also
recording the edge under the target — that reintroduces the duplication the layout removes.

## Base composed with drag; the delta is gitignored, not skip-worktree'd

**`base/` and `<label>/` are TRACKED and never written by the running sim.** They
hold, respectively, a node's `type`/`id`/`gate`/`drag`-rule plus its BASE position, and an
edge's wiring (`target`/`sourceHandle`/`targetHandle`/`kind`) plus its BASE geometry delta.
**Both the base position and the base delta are stored ONLY as an index** —
`base/`'s `indexPhi`/`indexTheta`/`indexR` (absolute, folded through
`polarindex.Canonical`) and `<label>/`'s `deltaIndexR`/`deltaIndexPhi`/`deltaIndexTheta`
(a relative integer step count, NEVER folded through `Canonical` — a delta is not a
position). There is no `scenePolarR`/`scenePolarPhi`/`scenePolarTheta` on a node or
`deltaPolarR`/`deltaPolarPhi`/`deltaPolarTheta` on an edge anywhere on disk any more; the
index is the sole authored quantity and `polarindex.ToPolar(idx, sc)` is the only multiply,
run at load. **A drag is stored the same way — an index delta,
`indexPhi`/`indexTheta`/`indexR` under `drag/` — and the drag value IS index × constant,
never a second stored continuous copy.** `drag/`
`drag/` and `drag/edges/<label>/` are GITIGNORED and are the only files a drag
writes; both hold exactly that one triple (the offset from `base/`'s own
`indexPhi`/`indexTheta`/`indexR`, e.g. a base `indexR` of 25 plus a 4-step drag stores `iR:
4`, not 29). There is no `dragPolarR`/`dragPolarPhi`/`dragPolarTheta` field anywhere in
either file — `polarindex.ToPolar(idx, sc)` is the ONLY place the multiply happens, and it
runs at load and at the moment a drag is measured, never stored as its own field. A
sub-step drag rounds to the nearest index (`polarindex.MeasureScalar`) and a drag that
rounds to zero is correct, not lost precision to recover. Neither file carries
`constantPhi`/`constantTheta`/`constantR` — nor does `base/` any more; those moved to
`constants/` at the scene root (read once per load, fails loudly by path on
missing/malformed input). The loaded triple (`polarindex.SceneConstants`)
is passed explicitly into every consumer, never a per-instance field or a package-level
global.

**A (base) and B (drag) are held SEPARATELY in the running program, never summed into a
stored value.**

The same split applies to an edge's target vector. `owners.Deltas` holds `baseTo` (A, seeded
once at build from each edge's `<label>/`, outgoing straight and incoming negated, never
mutated after) and `dragTo` (B, the only thing `ShiftSelfBy`/`ShiftOtherBy` mutate).
`DeltaTo` returns the derived compose; `DragDeltaTo` returns B as a `polar.Polar`.
`OutEdges.persistDelta` measures that into an index (`polarindex.MeasureScalar`) and writes
the index straight to `drag/edges/` `<label>/` — the same shape as the node path, no
continuous field survives the write.

THE DELTA (a node's standing vector to a neighbor, in `<label>/`) and THE DRAG (the
user's accumulated offset, in `drag/`) are different quantities. They are never summed into
a stored value and never substituted for one another; they combine only at the point
something is drawn, and persist always writes the DRAG, never the composed value.

On load, `Topology.ApplyDragOverlay` reads `base/`/`<label>/` into the BASE fields
and `drag/` and `drag/edges/<label>/` into separate DRAG fields on the same
spec struct — it composes NOTHING; a node or edge with no `drag/` file simply carries a zero drag
value forward. `polar.Compose`/`polar.Between` are untouched by this change; this is the
existing componentwise op, just applied at the render/persist boundary instead of at load.

Do NOT try to keep geometry in one tracked file and hide the drag write with
`git update-index --skip-worktree`. That was tried and abandoned: skip-worktree suppresses
`git status`/`git diff` reporting but NOT `git checkout`'s refusal to switch branches when a
marked file differs from the index — so the first branch switch after a drag forces the drag
output into a commit to unblock the checkout, the exact outcome the split exists to avoid. A
gitignored subdirectory has nothing in the index for `checkout` to compare against.

Edge WIRING (`target`/`sourceHandle`/`targetHandle`/`kind`/`label`) is authored data and
keeps writing to the tracked `<label>/` (`edgefile.WriteEdgeFile`) whenever it
legitimately changes. Only the geometry DELTA moved to `drag/`.

A drag touches more than the node you grabbed: the dragged node shifts its own vectors, and
each neighbour is told how far it moved and shifts its own — so a neighbour's `drag/`
`drag/` leaves and each of ITS outgoing edges' `drag/edges/<label>/` also get
written as part of that same drag, same as before.

Consequences to keep in mind:

- A fresh clone has the tracked base values only (no `drag/` directory yet — that only
  appears on the first LOCAL drag write), and loads a complete, self-consistent scene from
  them with zero deltas.
- Nothing promotes a drag's accumulated delta back into the base for a future clone; that
  would need a deliberate migration and commit (see the migration note below). Whatever is
  on disk right now under `base/`/`<label>/` is the base the NEXT clone gets.
- There is no more `*.geom.bin` to skip when scanning `nodes/*/edges/*.bin` — every file
  under an `edges/` dir is a real edge (`loadNodeEdges`, `check-no-fan-in.sh`).
- The existing `topology/`/`topology-pair/` trees were migrated by treating their
  already-committed values AS the base and starting with empty `drag/` dirs (zero deltas) —
  no attempt was made to reconstruct a pre-drag historical base, since none is recoverable.

## The owner writes, and owns the path

- A node writes its own `position/local-polars` and reads them back from one file,
  `Categories/Node/nodefile_drag_index.go` — Write and Read beside each other, so the four filenames
  and the directory are spelled once; the loader held a second reader.
- The **SOURCE NODE** owns `nodes/<source>/drag/edges/<label>.bin`, and writes it from
  `Categories/Node/node_out_edges.go`'s `persistDelta` — the same pass that derives that edge's geometry,
  since the node is what holds the vector being stored. The write is gated on the vector
  actually changing: derivation runs every tick, so an ungated write would rewrite the file
  every tick.
- Scene-level state (camera, overlays, sphere) is genuinely singular and belongs to the
  view-owner goroutine (`RunStdinReader`).

**One owner writes inside `nodes/<id>/`: that node.** Its own position/local-polars, and the
file of every edge leaving it.

Guards: `scripts/checks/source/check-persist-write-ownership.sh` (who may write which path pattern),
`Categories/Scene/Scenes/check-scene-path-resolution.sh` (who may construct a `nodes/` path).

## A topology is a directory tree, always

Node ids ARE
numbers — they are strings only because they are directory names — and **ROW ID = NODE ID -
1**: `loadTree` parses each node directory name to an int (`strconv.Atoi`) and that int,
minus one, IS the node's row directly — the directory its files live in. There is no ordering step left — a row is
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
the node-row space (the per-node value writers and the row-identity
lookup tables) must tolerate an empty row.

Per-node `edges/` file order is a PLAIN `sort.Strings` — those names are labels, not
numbers, and must stay lexicographic. Port order is alphabetical by port name — neither is
authored order, because a tree has no array.

There is no migration from any pre-split form: a tree holding only a legacy sidecar loses
its camera pose, overlay flags and sphere on load and falls back to defaults.
