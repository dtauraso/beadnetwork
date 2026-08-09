# Node palette: drag to create, delete to remove — plan

Delete this document when the change lands; git history is its archive.

**What it is.** A palette in the pair tab. Drag a node kind onto the scene and it is created
where it was dropped, connected to the NEAREST existing node. Select a node, press delete,
and it goes along with every edge touching it. A link that cannot exist is refused rather
than made. One new kind ships with it: a node that takes normals from two nodes and holds
their total.

There WAS a palette once (`NodePalette.tsx`, `bb9819cd`) and it is not this one. That was
the React-Flow era: TS owned the graph, a drop resolved to flow coordinates, and creation was
a store mutation. Go owns the topology now, so nothing of that survives but the gesture.

## The constraint that shapes everything

**Per-node and per-edge streams are dedicated fds, allocated by the HOST at spawn** from
`counts.json`, because Node's `spawn()` takes its stdio array up front and Go is not running
yet to be asked (`.claude/rules/persistence-ownership.md`, "Counts are stored, never
re-derived"). A node created in-process therefore has no stream and Go cannot emit its
frames.

So **create and delete are persist-then-respawn**, exactly like the scene-tab switch: write
the tree, end the process, and the host's already-looping runner spawns a new one that reads
the changed tree (`scene_tabs.go`'s `SelectScene`, "HOW THE SWITCH HAPPENS"). Agreed with
David. The alternative — dynamic fd allocation — is a much larger change to the bridge and
buys nothing a respawn does not already give.

## The reversals, named

1. **`create` and `delete` come back**, as ATTRIBUTES on the scene entity
   (`edit-update kind="scene" attr="create"` / `attr="delete"`), not as ops. `messages.ts`
   records that the create/delete OPS were removed end-to-end; the one-op rule stays intact
   because these are attributes, per CLAUDE.md's "new capability is a new entity kind or
   attribute, NOT a new op". Names agreed with David.
2. **A node's row space gains a hole.** Deleting a middle id leaves its row EMPTY rather
   than renumbering — renumbering is the silent-rename bug ROW ID = NODE ID - 1 exists to
   prevent. Every consumer of the row space already tolerates an empty row; this makes it
   routine rather than theoretical.

## Ripple list, in landing order

Each step builds and verifies on its own.

**1 — Bridge vocabulary.** `messages.ts` `EditMsg`: two new scene attributes. `create`
carries the kind id and the drop point; `delete` carries the target node's buffer ROW (never
an id — no sidecar). `input-layout.ts` encode/decode + the `updateAttrs` token in
`INPUT_LAYOUT_FINGERPRINT` (Go's `input_codec.go` is the source; the TS mirror is generated).
`stdin_reader.go`'s `applyUpdate` grows the two branches. Guards that will speak up:
`check-edit-op-parity`, `check-message-kind-parity`, `input_fixture_test.go` (regenerate via
`npm run gen:input-fixture`).

**2 — Create, in Go.** Resolve the nearest existing node to the drop point from Go's own
node geometry (TS never measures proximity). Allocate `largest id + 1`. Write
`nodes/<id>/meta.json` + `position.json`, and the edge file under the SOURCE node
(`nodes/<src>/edges/<label>.json`, outgoing only, no `source` key). Update `counts.json`
(`nodes` = largest id, `edges` = count + 1). Then end the run.

**3 — Delete, in Go.** Remove `nodes/<row's id>/` — its out-edges go with it. Walk every
other node's `edges/` and remove any whose target is the deleted node: in-edges are
deliberately not local, so this is a pass, and the loader already does the same walk. Update
`counts.json`. Then end the run.

**4 — Refuse impossible links.** A kind declares its ports as an explicit `[]PortSpec` at
`RegisterBuilder`. A link is refusable for two reasons and both are Go's to judge: the target
kind declares NO input port at all, or every input it declares is already bound by an
existing edge. Go performs no partial work in that case — nothing is written, nothing is
respawned, and no node appears.

**The drop is REFUSED WITH A VISIBLE SIGNAL** (David). Nothing is created, and the refusal
shows: a drop that silently does nothing is a bug indistinguishable from a broken build. The
signal is Go's to raise, since the judgement is Go's — it rides the same stream everything
else does, and TS renders it.

**5 — The new kind.** Four things in one commit (CLAUDE.md's primitive landing rule): a Go
package `nodes/<Kind>/` with its logic in `node.go` plus `SPEC.md`, the `NODE_DEFS` entry,
nothing else in the schema dir, and `go run ./tools/gen-node-defs`. Two input ports taking a
normal each, one held value that is their total, and an output carrying it.

**The total is DRAWN** (David), so it needs a buffer column and a render path — a
tilt-vector-style arrow from the node's centre along the summed direction, alongside the
arrows TiltVectors already draws.

**6 — The palette + the delete key, in TS.** A drag source per kind, listed from `NODE_DEFS`
(the single registry). The drop resolves a world point through the raycast that already runs
— an empty-space hit — and fires one addressed edit, fire-and-forget like every other
(`check-no-await-on-bridge`). Delete key: Go owns selection, so TS sends the SELECTED node's
row and nothing else. Both are render+forward; no domain state (`check-no-webview-state`).

**7 — Pair only.** A `SceneTab` field beside `DistanceGroups`/`UpAxis`/`ClockDivisor`, so
"which scenes can be edited" is a scene property Go owns rather than a TS branch on a name.

## Verification

- **Persistence through a real reload** is the one that matters
  (`memory/feedback_headless_repro_verifies_persistence.md`): drive the real binary
  headlessly, create a node, and assert the BYTES on disk — the new node directory, the edge
  under its source, `counts.json`. Then delete it and assert every edge that named it is gone
  from every other node's `edges/`. Green unit tests have hidden live persistence failures
  three times in this repo; this is the check that would not have.
- Per-goroutine tests for the refusal rule (one goroutine's own decision, no delivery).
- `bash scripts/stop-checks.sh` clean at each step, and `go run ./tools/gen-node-defs` after
  the kind lands.

## Risks

- **The respawn is user-visible**: in-flight beads are gone and the run restarts. Acceptable
  for an explicit structural edit, and identical to what a tab switch already does.
- **`counts.json` is hand-maintained today.** Step 2 makes Go a writer of it — the first one.
  It must stay single-writer: the operation that creates or deletes a node is the operation
  that updates it, and nothing else touches it.
- **A drop with no existing node to connect to** (an empty scene) has no nearest node. The
  node is created unconnected; that is not an error.
