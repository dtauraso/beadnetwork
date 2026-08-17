# docs/ — index

Entry point for the files under `docs/`. Grouped by topic; one line each.

**Reading the `.html` explainers:** `scripts/block-open-html-hook.py` blocks opening them
in a browser. Read them as text, or use the editor's HTML preview. They are self-contained
static pages (no external assets).

**Nothing here is a historical record.** Docs that described removed code, resolved
investigations, or superseded models were deleted rather than annotated — git history is the
history, and the repo holds what is current. If a page here is wrong, fix it or delete it.

**Planning docs are branch-local** (`.claude/rules/planning-docs.md`): a doc under
`docs/planning/` carries a `branch:` frontmatter and is stripped before merge by
`tools/strip-branch-local-docs.sh`, so on `main` that directory is empty between changes.
An untagged doc there escaped a merge — delete it.

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

## Pair node

| Doc | What it covers |
|---|---|
| [pair-node/index.html](pair-node/index.html) | The pair node — anatomy, channels, cycle, and the tilt rules. |
| [math/framework/index.html](pair-node/math/framework/index.html) | The tilt rules as one block of modular arithmetic, with the ring diagrams. |

## Process

| Doc | What it covers |
|---|---|
| [drift-checklist.md](process/drift-checklist.md) | Drift checklist — periodic agent/model-health audit. |

## Planning

`docs/planning/` is empty on `main` by rule — a plan lives on its branch and leaves with the
merge. Screenshots taken during a change go under `docs/planning/visual-editor/screenshots/`
with a date-prefixed kebab name (`tools/repo-hygiene/hooks/check-stray-screenshots.sh`), and
are referenced from whatever memory file the work lands in.
