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
topology/edges/<A>To<B>.json
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

## The model

**The goroutine that owns the state writes the file, and owns the path.**

- A `nodeMover` writes `topology/nodes/<its own id>/`. Nothing else constructs or touches
  that path.
- An `edgeMover` writes its own `topology/edges/<A>To<B>.json`.
- Scene-level state (camera, overlays, sphere) is genuinely singular, so it stays a single
  file — but with one named owning goroutine instead of a shared persister bag.

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

### 2. Give paths to their owners

Move node-path construction into the node mover and edge-path construction into the edge
mover. `scene_paths.go` keeps only scene-level paths and root resolution.

This step is what makes the rest possible. Until a path can only be built by its owner,
"the owner writes it" is a convention rather than a property.

### 3. Move each writer to its owner — one persister per commit

Start with `quant_offset_persist.go`. It is already the flat per-node model (one integer
triple `(iTheta,iPhi,iR)` per node, every node independent — see
`scene_node_pos_persist.go`'s header), so it is closest to owned already and is the
smallest diff.

### 4. Guard it

A check that `os.WriteFile` under `topology/nodes/` appears only in the node mover, and
under `topology/edges/` only in the edge mover. Same shape as
`tools/check-no-network-locks.sh`. This is what stops step 3 from silently regressing —
without it, the next central writer is one convenient refactor away.

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
