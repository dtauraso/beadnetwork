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
├── counts.json                           {"nodes": 9, "edges": 10}
├── nodes/<id>/
│   ├── meta.json                         type, polar position, localPolars
│   ├── position.json  data.json  local-polars.json  cascade-edges.json
│   ├── inputs/<PortName>.json            port geometry
│   ├── outputs/<PortName>.json           port geometry
│   └── edges/<label>.json                OUTGOING only
└── view/
    └── camera.json  overlays.json  sphere.json
```

An edge is stored under its **source** node and carries no `source` key — that is the
directory it sits in, and storing it too would be a second copy free to drift.

A node and its outgoing edges are one unit: deleting `nodes/5/` takes node 5's edges with
it, instead of leaving a dangling file in a sibling tree.

**In-edges are deliberately not local, and that costs nothing.** Nothing ever asks "what
points at node 9". Every use of an edge's target is built during a single full pass over all
edges (`buildEdgeMaps`' `inbound`, `loader_layout`'s `neighbors`). Do NOT "fix" this by also
recording the edge under the target — that reintroduces the duplication the layout removes.

## The owner writes, and owns the path

- A `nodeMover` writes its own `position/local-polars/cascade-edges` and its `inputs/` and
  `outputs/` port geometry, and constructs those paths (`node_mover.go`).
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

The monolithic single-file form is gone. `readSpec` rejects a non-directory. Node row order
is directory-sorted (alphabetical by id); port order is alphabetical by port name — neither
is authored order, because a tree has no array.

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

**Nobody re-derives it — not TS, and not Go.** Go's load iterates `nodes/` and each node's
`edges/`, but that is LOADING (reading each node's data to build the graph), not counting;
do not add a counting pass beside it. Correctness is single-writer: the one operation that
creates or deletes a node or edge updates `counts.json`. Nothing else writes it.

A missing or malformed `counts.json` must fail LOUDLY. Returning 0 allocates no dedicated
streams and degrades the bridge invisibly — the behaviour the old `countEdges` had.

No Go writer exists today, so the file is hand-maintained alongside the tree. The headless
harness (`headless_stream_helpers_test.go`) sizes its spawn from this same file and fails if
it disagrees with the tree, which is the only drift check that does not re-derive at runtime.
