---
branch: task/decentralize-persistence
---

# Decentralized persistence — each owner writes its own file

Status: plan, not started. Branch `task/decentralize-persistence`.

## The observation

The **data** is already decentralized. The **writers** are not.

Per-entity files exist today:

```
topology/nodes/<id>/meta.json
topology/nodes/<id>/position.json
topology/nodes/<id>/local-polars.json
topology/nodes/<id>/cascade-edges.json
topology/nodes/<id>/inputs/<PortName>.json     port geometry: {name, anchorId, portR}
topology/nodes/<id>/outputs/<PortName>.json    port geometry: {name, anchorId, portR}
topology/edges/<A>To<B>.json                   {source, sourceHandle, target, targetHandle, kind, label}
```

But the code that writes them lives in ~10 `nodes/Wiring/*_persist.go` files, and the
persisters hang off `MoveDispatch` (`md.persist.overlays`, `md.persist.sphere` —
`stdin_reader.go:284,288`). The node whose position changed does not write its own file; a
central object writes on its behalf.

`scene_paths.go` is the other half of the problem: one file knows every path in the tree
(`sceneTreeRoot`, `sceneJSONPath`, `cameraFilePath`, `overlaysFilePath`, `sphereFilePath`,
…). While any goroutine can construct any path, ownership is unenforceable — there is
nothing to violate.

There is also a legacy fork: `nodePosPersister.root == ""` means a monolithic
`topology.json` rather than the directory tree. Two topology forms are still supported.

## The layout: adjacency list

Everything about a node lives under `topology/nodes/<id>/`. The top-level `edges/`
directory goes away; an edge is stored under its **source** node, which is exactly an
adjacency list — a node's directory names the node and the nodes it points at.

```
topology/
├── nodes/
│   └── <id>/
│       ├── meta.json                     type, polar position, localPolars
│       ├── position.json
│       ├── data.json
│       ├── local-polars.json
│       ├── cascade-edges.json
│       ├── inputs/<PortName>.json        port geometry
│       ├── outputs/<PortName>.json       port geometry
│       └── edges/<label>.json            OUTGOING only: {sourceHandle, target,
│                                         targetHandle, kind, label}
└── view/
    ├── camera.json
    ├── overlays.json
    └── sphere.json
```

`source` drops out of the edge record — it is the directory the file is in. Storing it
would be a second copy of the same fact, free to drift.

**Why this is better than a sibling `edges/` directory:** a node and its outgoing edges
become one unit. Deleting `nodes/5/` deletes node 5's edges with it, instead of leaving
`edges/2To5.json` dangling as it does today. Reading one directory answers "what is this
node, and what does it point at?" without consulting a second tree.

**What it costs:** in-edges are not local. Finding what points *at* node 9 means scanning
every node's `edges/`. That is inherent to an adjacency list, and it is affordable here —
fan-in is rejected at parse (`validateNoFanIn`), so an input port has at most one incoming
edge, and load already walks every node directory anyway. Do **not** fix this by also
recording the edge under the target node: that is the same fact in two places, which is the
drift this layout exists to remove.

## The model

**The goroutine that owns the state writes the file, and owns the path.**

- A `nodeMover` writes its own `meta/position/data/local-polars/cascade-edges` and its
  `inputs/` and `outputs/` port geometry.
- An `edgeMover` writes its own `nodes/<source>/edges/<label>.json`.
- Scene-level state (camera, overlays, sphere) is genuinely singular, so it stays a single
  file — but with one named owning goroutine instead of a shared persister bag.

Note the consequence of the adjacency layout: **ownership is per-file-pattern, not per
directory.** Two goroutines write inside `nodes/1/` — node 1's mover, and the edgeMover of
each edge leaving node 1. The directory is an addressing scheme; the owner is determined by
the path pattern within it. The step-5 guard must therefore match paths
(`nodes/*/edges/*.json` → edge movers only), not directories.

The alternative — routing edge writes through the source node's mover — would make one
goroutine write another's state on request, which is the coordination this model exists to
avoid.

This is the rule the network already runs on, applied to disk: per-owner, no shared mutable
state, no coordinator. It is also what
`memory/project_lock_persistence_survives_respawn.md` already learned the hard way — each
follower must self-persist, or a respawn reloads stale positions.

## Order of work

### 1. Delete the monolithic form

Two supported shapes double the cost of every later step, and the tree form has won.
`nodePosPersister` currently exists *only* so tests can ask which form a `MoveDispatch`
loaded from — its own comment says it "schedules no writes of its own." Removing the fork
deletes the struct outright.

### 2. Move edges under their source node

`topology/edges/<A>To<B>.json` → `topology/nodes/<A>/edges/<label>.json`, dropping the now
redundant `source` key. Touches the edge reader, the edge writer, `scene_paths.go`, and the
committed `topology/` fixture; `nodes/<id>/cascade-edges.json` already proves per-node edge
data works.

Do this BEFORE moving writers around: it changes what the paths ARE, and step 3 is about
who may construct them. Reversing the order means doing step 3 twice.

There is no backward compatibility to keep — step 1 has already established that the
directory tree is the only supported form, and the tree in `topology/` is the only instance
of it.

### 3. Give paths to their owners

Move node-path construction into the node mover and edge-path construction into the edge
mover. `scene_paths.go` keeps only scene-level paths and root resolution.

This step is what makes the rest possible. Until a path can only be built by its owner,
"the owner writes it" is a convention rather than a property.

### 4. Move each writer to its owner — one persister per commit

Start with `quant_offset_persist.go`. It is already the flat per-node model (one integer
triple `(iTheta,iPhi,iR)` per node, every node independent — see
`scene_node_pos_persist.go`'s header), so it is closest to owned already and is the
smallest diff.

### 5. Guard it

A check matching PATH PATTERNS, not directories (see "The model" — after step 2 two
goroutines legitimately write inside `nodes/<id>/`):

- `nodes/*/edges/*.json` — written only by the edge mover
- everything else under `nodes/*/` — written only by the node mover
- `view/*.json` — written only by the scene owner

Same shape as `tools/check-no-network-locks.sh`. This is what stops steps 3–4 from silently
regressing — without it, the next central writer is one convenient refactor away.

## Explicitly out of scope

**Loading.** A node cannot read its own file before it exists; something must scan the tree
and decide what to create. That is the node-registry question and belongs with
`task/decentralize-node-build`, not here. Writing decentralizes cleanly because the owner
already exists by the time a write happens; reading is the harder half and bundling the two
would hide which one broke.

## Coordination note

`task/decentralize-node-build` has concurrent work on it (a `BuildArgs` /
`RegisterBuilder` seam, `Time` migrated first). That branch changes construction; this plan
changes persistence. They should not collide, but step 2 touches code adjacent to
`port_wiring.go` — check that branch before starting.
