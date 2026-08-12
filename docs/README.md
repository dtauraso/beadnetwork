# docs/ — index

Entry point for the 30 files under `docs/`. Grouped by topic; one line each.

**Reading the `.html` explainers:** `scripts/block-open-html-hook.py` blocks opening them
in a browser. Read them as text, or use the editor's HTML preview. They are self-contained
static pages (no external assets).

**Planning docs are branch-local going forward** (`.claude/rules/planning-docs.md` →
"Planning docs are branch-local"): new docs under `docs/planning/` carry a `branch:` frontmatter and are
stripped before merge. The existing untagged ones below predate that rule and stay until
individually judged.

## Concurrency & lock architecture

The mutex-removal work: each `sync.Mutex`/`Cond` replaced by single-owner state. Start with
`framings.md` for the overview, then the per-lock pages.

| Doc | What it covers |
|---|---|
| [framings.md](concurrency/framings.md) | The framing ledger — what replaced what, and the architecture built for the old model. No locks remain. |
| [concurrency-map.html](concurrency/concurrency-map.html) | Map of the concurrency model — goroutines, channels, who owns what. |
| [mutex-architecture.html](concurrency/mutex-architecture.html) | Overview of the mutex architecture (and its removal). |
| [outbox-architecture.html](concurrency/outbox-architecture.html) | `outbox.mu` resolved — per-direction channels replaced the shared move queue. |
| [trace-mutex-architecture.html](concurrency/trace-mutex-architecture.html) | `Trace.mu` resolved — events ride each owner's own stream. |
| [debounced-persister-architecture.html](concurrency/debounced-persister-architecture.html) | `debouncedPersister.mu` resolved — inline per-caller writes, no shared timer. |
| [scene-persist-architecture.html](concurrency/scene-persist-architecture.html) | `scene_persist` — the last unexamined locks; per-writer file ownership. |
| [node1-fanout-goroutines.html](concurrency/node1-fanout-goroutines.html) | Node-1 fan-out — one node driving several outgoing edges. |

## Investigations

| Doc | What it covers |
|---|---|
| [backpressure-investigation-order.md](investigations/backpressure-investigation-order.md) | Recommended order for the 7 backpressure/concurrency investigation branches (the branch docs themselves are branch-local). |
| [interior-stream-framing.md](investigations/interior-stream-framing.md) | Interior-stream framing corruption — investigation and reproduction. |
| [which-lattice-a-node-lives-on.md](investigations/which-lattice-a-node-lives-on.md) | Which lattice a node lives on — resolved; kept as history. |
| [audit-baseline.md](investigations/audit-baseline.md) | Audit baseline — settled findings audit subagents must not re-report. |

## Design specs & audits

| Doc | What it covers |
|---|---|
| [go-authoritative-clock/index.html](go-authoritative-clock/index.html) | Go-authoritative clock design spec — Go owns the one clock. |
| [level4-audit/index.html](level4-audit/index.html) | Level-4 audit of the system. |

## Polar geometry & layout

| Doc | What it covers |
|---|---|
| [polar-sphere.html](polar-geometry/polar-sphere.html) | The polar coordinate system for a sphere. |
| [pole-singularity.html](polar-geometry/pole-singularity.html) | The layout pole singularity — φ grid vs great-circle bearing. |

## Bead / edge model

| Doc | What it covers |
|---|---|
| [beads-are-the-edge.md](bead-model/beads-are-the-edge.md) | Beads are the edge — the node-owned chain of placeholder beads (superseded on the length model; chain description still current). |
| [bead-lattice.md](bead-model/bead-lattice.md) | The bead lattice — an edge is one integer; supersedes the arc-length model. |
| [arc-from-local-polar.md](bead-model/arc-from-local-polar.md) | One integer per edge — the arc comes from the stored LocalPolar (superseded by bead-lattice.md). |
| [channels-not-ports.md](bead-model/channels-not-ports.md) | A port is a ROLE, not a place — agreed model narrowing MODEL.md. |

## Process

| Doc | What it covers |
|---|---|
| [drift-checklist.md](process/drift-checklist.md) | Drift checklist — periodic agent/model-health audit. |

## Visual-editor planning (`docs/planning/visual-editor/`)

Planning/spec (untagged, predate the branch-local rule):

| Doc | What it covers |
|---|---|
| [camera-navigation.html](planning/visual-editor/camera-navigation.html) | 3D camera navigation model. |
| [edit-hop-audit.html](planning/visual-editor/edit-hop-audit.html) | Edit round-trip audit — why 12 hops. |
| [node-edges-goroutine-spec.html](planning/visual-editor/node-edges-goroutine-spec.html) | A node runs its own outgoing edges. |
| [sphere-chain-layout-spec.html](planning/visual-editor/sphere-chain-layout-spec.html) | Sphere-chain node layout. |
| [timing-spec.html](planning/visual-editor/timing-spec.html) | Wirefold timing spec. |
| [timing-window.html](planning/visual-editor/timing-window.html) | Timing-window spec. |
| [animation-drag-issues.md](planning/visual-editor/animation-drag-issues.md) | Live-observed open issues in animation & dragging. |
| [double-link-polar-model.md](planning/visual-editor/double-link-polar-model.md) | Double-link polar movement model. |
| [existing-lock-system-record.md](planning/visual-editor/existing-lock-system-record.md) | Lock-system record kept before the double-link rewrite. |
| [layout-on-domain-network.md](planning/visual-editor/layout-on-domain-network.md) | Layout on the domain network (rebuild). |
| [one-clock-sleep-only.md](planning/visual-editor/one-clock-sleep-only.md) | One clock, sleep-only pacing (decided model). |
| [polar-frame-rewrite.md](planning/visual-editor/polar-frame-rewrite.md) | Polar-frame rewrite plan (preserved from a deleted task branch). |

Screenshots: `planning/visual-editor/screenshots/` — panguide triangle-drift (2026-06-17,
×2) and saturated pulse-wires (2026-07-14).
