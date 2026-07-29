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

But the code that writes them lives in 8 `nodes/Wiring/*_persist.go` files, and the
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

**What it costs: nothing.** The textbook objection to an adjacency list is that in-edges are
expensive — but nothing here ever asks for them. Every use of an edge's target is built
during a SINGLE full pass over all edges (`buildEdgeMaps`' `inbound` map, `loader_layout`'s
`neighbors`); there is no query of the form "find the edges pointing at node 9". Walking
every node's `edges/` yields the identical complete edge list in one sweep, and those maps
get built from it exactly as they are today.

That is a consequence of the model, not luck: each goroutine takes its own state from its
own place in the tree. Nothing traverses the graph hunting for relationships — an edge
belongs to its source, and the target learns of it when the source's edge is wired.

Do **not** "fix" a non-problem by also recording the edge under the target node: that is the
same fact in two places, which is the drift this layout exists to remove.

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

Two supported shapes double the cost of every later step, and the tree form has won. This is
a larger and cleaner deletion than it first looks — all of it is already dead or vestigial:

- `nodePosPersister` exists *only* so tests can ask which form a `MoveDispatch` loaded from;
  its own comment says it "schedules no writes of its own."
- `sceneTreeRoot`'s monolithic branching and `sceneJSONPath`'s true-monolithic fallback in
  `scene_paths.go` — the `IsDir()` probing and the `return ""` path.
- `LoadTopologyFromJSON` (`loader.go:76`), an exported entry point that "builds a topology
  from an in-memory JSON blob, bypassing the tree". **Zero callers** — production or test.

Verified there is nothing to migrate: no `topology.json` and no second `meta.json` tree
exists anywhere in the repo, and no test constructs a monolithic topology.

**NOT in this step: `sceneCameraPath`.** Its own comment claims "all call sites have been
migrated", and that comment is wrong — `scene_camera.go:82`,
`scene_overlays_persist.go:166` and `scene_sphere_persist.go:49` all still call it. It backs
a DIFFERENT legacy: the pre-split `view/scene.json`, read best-effort when
`camera.json`/`overlays.json`/`sphere.json` are absent ("Legacy fallback: pre-split topology
only has scene.json's cameraPolar key", `scene_camera.go:80`).

Two legacies are in play and must not be conflated: the monolithic TOPOLOGY form (a
`topology.json` file instead of a tree — this step's target) and the pre-split SCENE file
(`view/scene.json` — untouched here). No `view/scene.json` exists anywhere in the repo, so
that fallback is dead data-wise but live code-wise; removing it would drop the camera,
overlays and sphere of anyone holding an old one on disk. Its own decision, not this one.

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

### 6. Stop TS counting the tree — store the counts

`tools/topology-vscode/src/runCommand.ts` has `countNodes` and `countEdges`, which walk the
tree to size the stdio fd array. They exist because the extension host is the process that
SPAWNS Go, and Node's `spawn()` takes the stdio array up front: with one dedicated pipe per
emitting goroutine, the pipe count is a function of graph size and must be known before the
child exists. Go cannot answer, because Go is not running yet.

The cost is that the on-disk layout becomes knowledge held in TWO languages with nothing
enforcing agreement. Step 2 was exactly that near-miss: the files moved and `countEdges` had
to move with them, or it would have returned 0, allocated no dedicated edge streams, and
failed as a bridge problem rather than a test failure.

**The counts are stored, not walked.** They are fixed for a given topology and change only
when the editor adds or removes a node or an edge — the operation that changes them already
knows the delta, so it writes them. TS reads two numbers and stops knowing the layout
entirely; `countNodes` and `countEdges` both delete.

**Nobody re-derives them — not TS, and not Go either.** An earlier draft of this step had Go
compare the declared counts against a walk of the tree at load. That was the same mistake in
a different language: it re-derives a fixed number in order to check a fixed number. The
count is written once by the operation that changes it, and read. That is the whole design.

Go's load already iterates `nodes/` and each node's `edges/`, but that is LOADING — reading
each node's data to build the graph — not counting. Do not add a counting pass alongside it.

The correctness argument is single-writer, not verification: exactly one operation creates or
deletes a node or edge, and that operation updates the count. Nothing else writes it, so
there is nothing to reconcile.

Alternative considered and rejected: a `--count` pre-pass (spawn Go once to print the
numbers, then spawn for real). It needs no file and no writer, but costs a process spawn per
run and leaves the counts underivable without running a binary. Storing them keeps the
topology self-describing.

## Explicitly out of scope

**Loading.** A node cannot read its own file before it exists; something must scan the tree
and decide what to create. That is the node-registry question, not this one. Writing
decentralizes cleanly because the owner already exists by the time a write happens; reading
is the harder half, and bundling the two would hide which one broke.

Note what loading now looks like, since it changed under this plan's feet (see below):
`RegisterBuilder(kind, ports, build)` takes the port manifest as an explicit `[]PortSpec`
argument, so a kind's ports are known at registration, before any node is constructed. The
loader still walks the tree, reads each `meta.json`'s `type`, and looks it up in `Registry`.
Under the adjacency layout that walk also picks up each node's outgoing edges from
`nodes/<id>/edges/` instead of a separate `topology/edges/` pass — a change to WHERE the
loader reads, not to who does the reading.

## Construction rework — already landed

Construction was decentralized separately and is **merged** (`2ca6f7ce` … `39ae89b4`): every
kind now constructs itself through a `BuildArgs` seam, and the central reflection pipeline
(`reflectBuild`, `reflectPorts`) is deleted. This plan was written against the old code, so
its claims were re-verified against the post-merge tree — `nodePosPersister`,
`md.persist.overlays`/`md.persist.sphere`, and `scene_paths.go`'s 7 path constructors are
all still exactly as described. That rework touched construction only; persistence was left
untouched, which is why this plan survives it unchanged.

One piece of loose drift it left behind, worth folding into step 1 or fixing separately: 21
comments under `nodes/` still reference `reflectBuild`/`reflectPorts`, neither of which
exists (e.g. `nodes/selectright/node.go:20`, `nodes/TimeStart/node.go:171`) — despite
`39ae89b4` being titled "retire the reflectBuild claims the deletion left behind". Adding
both tokens to `DEAD_COMMENT_TOKENS` in `tools/check-comment-vocab.sh` after fixing them
prevents a third round.
