# docs/ — index

Entry point for the files under `docs/`. Grouped by topic; one line each.

**Reading the `.html` explainers:** `scripts/block-open-html-hook.py` blocks opening them
in a browser. Read them as text, or use the editor's HTML preview. They are self-contained
static pages (no external assets).

**Nothing here is a historical record.** Docs that described removed code, resolved
investigations, or superseded models were deleted rather than annotated — git history is the
history, and the repo holds what is current. If a page here is wrong, fix it or delete it.

**Planning docs are branch-local going forward** (`.claude/rules/planning-docs.md` →
"Planning docs are branch-local"): new docs under `docs/planning/` carry a `branch:`
frontmatter and are stripped before merge by `tools/strip-branch-local-docs.sh`. The
untagged ones below predate that rule and stay until individually judged.

## Concurrency

No locks remain in the network; each `sync.Mutex`/`Cond` was replaced by single-owner state.

| Doc | What it covers |
|---|---|
| [concurrency-map/index.html](concurrency/concurrency-map/index.html) | Map of the concurrency model — goroutines, channels, who owns what. |
| [node1-fanout/index.html](concurrency/node1-fanout/index.html) | Node-1 fan-out — one node driving several outgoing edges. |

## Investigations

| Doc | What it covers |
|---|---|
| [audit-baseline.md](investigations/audit-baseline.md) | Audit baseline — settled findings audit subagents must not re-report. |

## The model

| Doc | What it covers |
|---|---|
| [entities.md](model/entities.md) | The entities: bead, wire, node goroutine, input port, clock. |
| [editor-surface.md](model/editor-surface.md) | The editor surface — streams, blocks, who writes what. |
| [timing.md](model/timing.md) | Timing — ticks, dwell, in-flight revision. |
| [lifecycle.md](model/lifecycle.md) | Lifecycle — load, run, respawn. |
| [polar-model.md](model/polar-model.md) | The polar model — index × constant, base composed with drag. |
| [polar-model-drag.md](model/polar-model-drag.md) | The drag half of the polar model. |
| [scenes.md](model/scenes.md) | Scenes — a scene declares its own tab and its own forks. |

## Polar geometry & layout

| Doc | What it covers |
|---|---|
| [pole-singularity/index.html](polar-geometry/pole-singularity/index.html) | The layout pole singularity — φ grid vs great-circle bearing. |

## Bead / edge model

| Doc | What it covers |
|---|---|
| [bead-lattice.md](bead-model/bead-lattice.md) | The bead lattice — an edge is one integer. |
| [bead-count.md](bead-model/bead-count.md) | How the source node computes the bead count, and where the rounding is. |
| [channels-not-ports.md](bead-model/channels-not-ports.md) | A port is a ROLE, not a place — agreed model, not landed. |

## Pair node

| Doc | What it covers |
|---|---|
| [pair-node/index.html](pair-node/index.html) | The pair node — anatomy, channels, cycle, and the tilt rules. |
| [math/framework/index.html](pair-node/math/framework/index.html) | The tilt rules as one block of modular arithmetic, with the ring diagrams. |

## Process

| Doc | What it covers |
|---|---|
| [drift-checklist.md](process/drift-checklist.md) | Drift checklist — periodic agent/model-health audit. |

## Visual-editor planning (`docs/planning/visual-editor/`)

Planning/spec (untagged, predate the branch-local rule):

| Doc | What it covers |
|---|---|
| [edit-hop-audit/index.html](planning/visual-editor/edit-hop-audit/index.html) | Edit round-trip audit — why 12 hops. |
| [node-edges/index.html](planning/visual-editor/node-edges/index.html) | A node runs its own outgoing edges. |
| [sphere-chain/index.html](planning/visual-editor/sphere-chain/index.html) | Sphere-chain node layout. |
| [timing-spec/index.html](planning/visual-editor/timing-spec/index.html) | Wirefold timing spec. |
| [timing-window/index.html](planning/visual-editor/timing-window/index.html) | Timing-window spec. |
| [double-link-polar-model.md](planning/visual-editor/double-link-polar-model.md) | Double-link polar movement model. |
| [polar-frame-rewrite.md](planning/visual-editor/polar-frame-rewrite.md) | Polar-frame rewrite plan (preserved from a deleted task branch). |

Screenshots: `planning/visual-editor/screenshots/` — panguide triangle-drift (2026-06-17,
×2), saturated pulse-wires (2026-07-14), pair normals mirrored (2026-08-06).
