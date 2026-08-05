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
├── counts.json                           {"nodes": 9, "edges": 10}  — nodes = ROW COUNT
│                                          (largest node id), not a live-node count
├── nodes/<id>/
│   ├── meta.json                         type, polar position, localPolars
│   ├── position.json  data.json  local-polars.json  cascade-edges.json
│   └── edges/<label>.json                OUTGOING only
└── view/
    └── camera.json  overlays.json  sphere.json  scene.json
```

**`topology/` is one of several sibling SCENES**, not the only tree. `nodes/Wiring/scene_tabs.go`'s
`SceneTabs` names each sibling directory (today: `topology/`, `topology-pair/`) resolved
relative to the ANCHOR's parent — the `-topology` flag the extension host launches with is
the fixed anchor, and which sibling directory actually loads is resolved from it
(`ResolveScenePath`). Each sibling is a COMPLETE, independently loadable tree with its own
`counts.json` and `view/`, laid out exactly as above. `topology/view/scene.json`
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

## The owner writes, and owns the path

- A `nodeMover` writes its own `position/local-polars/cascade-edges`, and constructs those
  paths (`node_mover.go`). There is no longer a separate `inputs/`/`outputs/` port-geometry
  file — port geometry was removed with the port model (edges attach on the bead lattice,
  docs/bead-lattice.md); this bullet used to list it as a second thing the mover writes.
- An `edgeMover` owns `nodes/<source>/edges/<label>.json`. No Go writer exists yet — edges
  are editor-authored — but when one is added, its path construction belongs there.
- Scene-level state (camera, overlays, sphere) is genuinely singular and belongs to the
  view-owner goroutine (`RunStdinReader`).

**Ownership is per-file-pattern, not per-directory.** Two owners legitimately write inside
`nodes/<id>/`: that node's mover, and the edgeMover of each edge leaving it. Routing edge
writes through the source node's mover would make one goroutine write another's state on
request — the coordination this model exists to avoid.

Guards: `tools/check-persist-write-ownership.sh` (who may write which path pattern),
`tools/check-scene-path-resolution.sh` (who may construct a `nodes/` path).

## A topology is a directory tree, always

The monolithic single-file form is gone. `readSpec` rejects a non-directory. Node ids ARE
numbers — they are strings only because they are directory names — and **ROW ID = NODE ID -
1**: `loadTree` parses each node directory name to an int (`strconv.Atoi`) and that int,
minus one, IS the node's buffer row directly. There is no ordering step left — a row is
declared by the id, never derived by sorting or by position in a list, so there is nothing
for a "10" vs "2" comparison to get wrong any more. A node directory name that fails to
parse as a number, an id below 1 (ids are 1-based), or a duplicate parsed id is each a LOAD
ERROR, loud and naming the offending directory, like a missing/malformed `counts.json`
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

`counts.json` exists because the extension host SPAWNS Go, and Node's `spawn()` takes the
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
`counts.json` — Go never opens that file at all; it exists solely so something can size the
stdio array BEFORE Go is running to answer. Correctness is single-writer: the one operation
that creates, deletes, or renumbers a node, or creates/deletes an edge, updates
`counts.json`. Nothing else writes it.

A missing or malformed `counts.json` must fail LOUDLY. Returning 0 allocates no dedicated
streams and degrades the bridge invisibly — the behaviour the old `countEdges` had. The
extension host reader (`tools/topology-vscode/src/runCommand.ts`'s `readCounts`) and the Go
headless test harness (`headless_stream_helpers_test.go`) must fail the same way if the
stored `nodes` value disagrees with the tree's own largest id — not just on a missing file.

No Go writer exists today, so the file is hand-maintained alongside the tree. The headless
harness sizes its spawn from this same file and fails if it disagrees with the tree's own
computed row count, which is the only drift check that does not re-derive at runtime.
