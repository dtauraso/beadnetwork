# Visual editor — real-world session log

Append-only log of friction surfaced while driving the visual editor. Newest first.

## Entry format

```
## YYYY-MM-DD — <short task description>

**Observation:** what the user noticed driving the editor.

**Hypothesis / scope:** quick read on what's likely going on.

**Decision:** start a task branch, log-only, defer, etc.

**Outcome:** what changed (or "logged only").
```

---

## 2026-06-27 — Shipped summary (folded from the retired handoff.md "What shipped")

Durable record of work that reached `main`, relocated when `handoff.md` was retired in
favor of git-derived state (`tools/next.sh` + branch descriptions):

- **Viewpoint nav math moved TS → Go; camera is Go-owned and POLAR.** Go holds
  `(pivot, r, pos, up)` (`nodes/Wiring/viewpoint.go`) and does angle-only spherical trig
  (`spherical.go`: `dir{θ,φ}`, `rot{axis,angle}`, `rotateDir`/`arcBetween`/`angleAboutAxis`).
  Trig is **epsilon-free** (great-circle bearing form, `atan2` of two unnormalized terms —
  no `/sinθ`, no pole special-case). Four gestures route through Go: `edit` op `viewpoint`
  (`kind` = set/orbit/orbit-locked/zoom/pan) in, `camera` trace event out. TS does only the
  two edge conversions (pointer px→angles; polar→three.js, quaternion at draw) —
  `viewpoint-bridge.ts`, `CameraFromStore.tsx`. Persists as `cameraPolar` in scene.json.
  Only Cartesian in Go is `pivot` (translated, never rotated). Zoom-to-cursor is a dolly = pan.
- **Pick resolution** by `userData.nodeId`/`body` across all pick paths; z-blind proximity
  fallback gone; handholds excluded from node picks.
- **Port-move** projects pointer onto the node's own ring plane (`z = nodeCenter.z`), not `z=0`.
- **Dynamic port auto-aim** (`AimedPortRegistry`, `aimed_ports.go` + loader.go): edges 1→2,
  1→6, 1→8, 2→3, 2→7 aim source port at child and child input back — radial spokes from LOAD.
  Node 8 `FeedbackOut` (8→1) stays ring-anchored and manually movable.
- **θ-lock** (`thetaLock`, lock.go): nodes 2 & 6 on 1, and 3 & 7 on 2, share θ, each keeps its
  own φ. Registered by id in loader.go (stopgap, like the chord lock).
- **Node kind `Excitatory` → `Pulse`** (pure rename; package `nodes/pulse`; nodes 6,7 `Pulse`).
- **HoldFlip (node 4)** mirrors Pulse: main loop drains input to LATEST + updates interior bead
  immediately; drive goroutine continuously pulses the flip (`1-held`). Continuous-drive output;
  before any input it drives the `-1` placeholder.
- **WindowAndGate (node 5)** discards `-1` placeholders, re-samples each side to most-recent real bead.
- **Trace** serializes stdout writes (`drain()` holds the mutex across the sink write).
- **Polar frame markers + scene-tori "rings" toggle** (theta-lock-diag keepers): camera-
  independent +y/+x/+z axis markers + labels in NavGuides.tsx; Go-owned scene-tori show/hide.

---

## 2026-06-12 — post-redesign follow-ups: prebuilt-binary runner + zombie-bead reset (2 branches)

**Observation:** After the persistence redesign merged, two friction points surfaced
driving the editor: (a) every launch re-linked via `go run .`; (b) a bead in-flight when
STOP killed Go reappeared as a zombie in the next run (seen on `2To3`).

**Decision:** two task branches, both merged + deleted.

**Outcome:**
- **Prebuilt-binary runner + `.go` watcher + orphan reap** (merge `6c8a1f31`,
  `task/prebuilt-binary-runner`): editor spawns a prebuilt
  `<repoRoot>/.wirefold-cache/wirefold` (gitignored) instead of `go run .`. Lazy staleness
  check (`ensureBinaryBuilt`, `runCommand.ts`) + eager `**/*.go` `FileSystemWatcher`
  (`extension.ts`, 250ms debounce) both rebuild via shared `buildBinary()` (`goBuild.ts`,
  module-level `building` guard = wait-free coalesce). `killOrphanedSims()` SIGKILLs
  leftover sims from crashed sessions on launch. First launch after fresh checkout does a
  one-time `go build`; reused until a `.go` changes.
- **Zombie-bead-on-restart fix** (merge `f40260b8`, `task/clear-pulses-on-restart`): added
  `clearAllPulses()` (swaps in a fresh empty `Map` in `webview/three/pulse-state.ts`),
  called at the TOP of `store.load()`. Go emits its startup spec → `load` on every restart,
  so each run wipes prior transient beads; pause doesn't route through `load()` so beads
  correctly persist across pause. Pure render-state reset at the run boundary — no change to
  wire timing or the bead model. Clear-on-bare-stop declined (beads linger until next run by
  choice).

---

## 2026-06-12 — topology-tree / Go-owns-persistence redesign: live bringup (task/persist-geometry-from-go-stream)

**Observation:** The 4-phase redesign (tree reader, tree writer, command-launched panel,
flowToSpec retirement) was built and statically verified last session but had not been
exercised in the live editor.

**Scope:** Bring the redesign up in the live editor; confirm the original branch goal
(port anchor survives reload) end to end.

**Five real fixes needed after the build:**
1. Startup load-race — Go's one-shot spec line beat the webview listener; fixed by caching
   `lastSpec` in the extension host and replaying `"load"` on webview `"ready"`.
2. Dead document-gate — `handle-message.ts` gated webview-log on the removed custom-editor
   `document`; fixed by threading a `logUri` into `MessageCtx` and removing the gates.
3. parseSpec schema mismatch — tree `EmitSpecLine` dropped port `kind` and edge `id` that
   `parseSpec` required; fixed by emitting edge `id`, applying `json:"-"` to
   `specNode.Position`, defaulting `kind` in `parsePort`, and adding a load-error breadcrumb
   to surface silent `store.load` throws to `ts-errors.jsonl`.
4. Auto-fit not firing / wrong framing — `CameraFitter` bailed before Go geometry was
   present; fixed to fit once-per-epoch gated on full Go geometry; dropped y-negation to
   match manual Fit.
5. `tree_writer` now emits compact single-line JSON matching the fixtures.

**Outcome:** Live-verified persistence round-trip: dragging a node wrote
`topology/view/nodes/<id>.json`; dragging a port anchor wrote
`topology/nodes/<id>/inputs/<port>.json` with an `"anchor"` field; on reload the spec
stream carried both the moved position AND the port anchor. The original branch goal
confirmed. Redesign complete; merging to `main` this session.

---

## 2026-06-01 — InhibitRightGate window verified live (no task branch)

**Observation:** handoff open-item #2 asked whether InhibitRightGate's coincidence window had ever been validated against actual input alignment in the live ring, or whether the ring was merely "not starved."

**Finding 1 — existing coverage understated:** the open-item wording was stale. InhibitRightGate already has direct unit coverage in `nodes/inhibitrightgate/firing_rule_test.go` (TestWindowFire / TestWindowClear). ReadGate's window (commit 48749fd) explicitly mirrors it. "Not independently verified" understated what was already in the test suite.

**Finding 2 — ring cannot run headlessly:** `go run . -duration=20s` builds and starts but deadlocks after the first hop — only bootstrap_rg + in08 fire, then nothing. Cause: poll-and-hold delivery. A Send marks a bead inFlight; the value only enters the destination slot when NotifyDelivered fires, which is driven by the visual layer (webview pulse-completion → stdin reader). No editor = no delivered messages = ring stalls. By design, not a bug. To exercise the ring you need the live editor (.probe relay) or a headless delivery driver.

**Finding 3 — live measurement:** cleared .probe, ran once in the editor. 12.1 s, 18 fires across 6 nodes, 0 errors. Per gate: inhibitRight0 = 2 fires / 0 window_clear; readGate1 = 4 fires / 0 window_clear. The gate's inputs genuinely coincide within W in the live ring — it is not surviving via lucky non-starvation. Caveat: small sample (2 fires / 12 s); a 60 s run would strengthen the ratio, but the qualitative signal (zero clears, zero errors) is unambiguous.

**Decision:** log-only (no task branch). Open-item #2 resolved.

**Outcome:** open-item #2 closed. Handoff updated.

---

<!-- source: session-log/2026-05-03-industry-standard-pattern-review-visual-editor.md -->

## 2026-05-03 — industry-standard-pattern review (visual editor)

**Branch:** task/industry-pattern-review
**Mode:** AI-driven audit (CLAUDE.md "what did the rest of the world
converge on?" rule). No implementation this session — output is a
coverage matrix and triage of which gaps merit task branches.

Reference set: yEd, draw.io (mxGraph), ELK, the React Flow ecosystem
(incl. xyflow Pro examples), JointJS. Patterns surveyed are the
ones a typical graph-editor user expects on first contact.

### Coverage matrix

| Pattern | Have it? | Where / gap | Rough effort |
|---|---|---|---|
| Pan / zoom / fit-view on load | Yes | [app.tsx:892](../../../tools/topology-vscode/src/webview/rf/app.tsx#L892) (`fitView`), `minZoom 0.1`, `maxZoom 4` | — |
| Snap-to-grid | Yes | [app.tsx:879-880](../../../tools/topology-vscode/src/webview/rf/app.tsx#L879-L880) (`GRID=24`) | — |
| Alignment guides during drag | Partial | [app.tsx:681-707](../../../tools/topology-vscode/src/webview/rf/app.tsx#L681-L707) — single-node only; multi-node selection drag clears guides intentionally | S — extend to bbox of selection |
| Marquee / lasso selection | Yes | `selectionOnDrag`, `SelectionMode.Partial`, `panOnDrag={[1]}` ([app.tsx:895-897](../../../tools/topology-vscode/src/webview/rf/app.tsx#L895-L897)) | — |
| Multi-select drag | Yes (RF default) | — | — |
| Port-anchored handles | Yes | `sourceHandle`/`targetHandle`; 1-to-1 input invariant enforced ([app.tsx:474-482](../../../tools/topology-vscode/src/webview/rf/app.tsx#L474-L482)) | — |
| Edge reroute (drag endpoint) | Yes | `onEdgeUpdate*` handlers | — |
| Orthogonal routing | Partial | `EdgeRoute = "line"\|"snake"\|"below"` ([schema.ts:62](../../../tools/topology-vscode/src/schema.ts#L62)); snake is orthogonal but **sharp corners** ([AnimatedEdge.tsx:155-167](../../../tools/topology-vscode/src/webview/rf/AnimatedEdge.tsx#L155-L167)) | S — replace `L` with rounded-corner arcs / `Q` |
| Rounded corners on orthogonal edges | **No** | sharp 90° corners only | folded into above |
| Auto-routing (avoid node overlaps) | **No** | corridor offset is fixed `+40`; no obstacle awareness | L — adopt a router (ELK, libavoid-js) or accept `route` field as authoritative |
| Edge labels | Partial | edges have a `label` (Go identifier, not display label); not rendered on canvas | M — render display label on `BaseEdge`, anchor near midpoint |
| Edge-label collision avoidance | **No** | n/a until labels render | M (after labels) |
| MiniMap / overview | **No** | no `MiniMap` import; only `Controls` | XS — drop in `<MiniMap />` |
| Zoom-to-fit shortcut | Partial | bridge exposes `fitNodes(ids)` ([app.tsx:255-261](../../../tools/topology-vscode/src/webview/rf/app.tsx#L255-L261)) but no global `f` / `cmd-1` keybinding | XS — wire keybinding |
| Zoom-to-selection | Partial | `fitNodes` works on selection via bridge, no keyboard hook | XS |
| Undo / redo | Yes | scoped stacks (spec / viewer), gesture-aware via `data-undo-scope` ([app.tsx:140-236](../../../tools/topology-vscode/src/webview/rf/app.tsx#L140-L236)) | — |
| Undo grouping at gesture level | Partial | `mutateBoth` groups spec+viewer for delete; multi-node drag pushes one history entry per node-drag-stop (not coalesced) | S — coalesce within a single selection-drag gesture |
| Copy / paste | **No** | no clipboard handlers | M — serialize selection subgraph, regen ids on paste |
| Duplicate (cmd-D) | **No** | — | S (after copy/paste plumbing) |
| Keyboard nav (arrows nudge, tab through nodes) | **No** | only delete + cmd-Z + space (onion swap) | S for arrow-nudge by GRID; M for tab cycle |
| Lane / swimlane containment | Partial | `Fold` placeholder collapses N nodes into one, **not** an open container holding children visually like a draw.io swimlane | L — different abstraction; would need parent-node support |
| Group / ungroup | Partial via folds | — | — |
| Node search / quick-jump (cmd-K palette) | **No** | — | S |
| Context menus | Partial | edge-kind and fold menus exist; no general node menu (rename/duplicate/etc.) | S |
| Keybinding cheatsheet / discoverability | **No** | — | XS — static panel |
| Touch / trackpad pan | Yes | `panOnScroll={true}` | — |
| Connect-validation feedback | Partial | port-conflict logged to console, no UI cue ([app.tsx:478-481](../../../tools/topology-vscode/src/webview/rf/app.tsx#L478-L481)) | XS — toast or red handle flash |
| Diff / compare view | Yes (project-specific) | A-live / A-other / B-onion modes | — beyond category baseline |
| Auto-layout (one-shot) | **No** | manual placement only; ELK / dagre are canonical drop-ins | M |

### Triage — which gaps deserve a task branch

**High value, low effort (open branches when next friction surfaces):**
1. **MiniMap** — XS, drop-in `<MiniMap />`. Standard expectation; perceptual win for "where am I in this graph."
2. **Zoom-to-fit / zoom-to-selection keybindings** — XS, function already exists; bind `f` and `shift-f`.
3. **Rounded corners on `snake` route** — S, single edit in [AnimatedEdge.tsx:155-167](../../../tools/topology-vscode/src/webview/rf/AnimatedEdge.tsx#L155-L167). Visual polish every other tool has.
4. **Connect-validation UI cue** — XS, replace `console.warn` for port-already-wired with a transient red flash on the rejected target handle.

**High value, medium effort (branch when friction is logged):**
5. **Copy / paste / duplicate** — M, baseline expectation. Subgraph serialization with id regeneration; hooks into existing `mutateSpec` + `scheduleSave`.
6. **Edge display labels** — M; channel-name `label` is currently invisible, surprising once more than a handful of edges are wired.
7. **Multi-node alignment guides** — S, extend matcher to use selection-bbox center.
8. **Drag-stop undo coalescing** — S, one history entry per multi-node drag gesture.

**High value, high effort (hold until friction insists):**
9. **Auto-routing with obstacle avoidance** — L. Biggest gap vs. yEd/draw.io. Defer until "edges crossing through nodes" gets logged. Canonical answer is ELK or libavoid-js; prefer adopting over rolling.
10. **Auto-layout (dagre / ELK one-shot)** — M-L. Defer until a 50+ node spec generates complaints about hand-placement.

**Low priority / different shape:**
- Swimlanes — fold abstraction already covers "collapse a region"; container-style lanes are a different mental model.
- Keyboard tab-through-nodes — diminishing return for graphs of this size.

**Branch-opening recommendations (proposed, not started):**
- `task/fix-minimap-add` (item 1).
- `task/fix-zoom-keybindings` (item 2).
- `task/fix-snake-rounded-corners` (item 3).
- Bundle items 1–4 into a single `task/industry-quick-wins` branch if landed together (≥$5 cost-marker territory).
- Items 5–8 wait for explicit friction logged in this session-log
  before opening branches.
- Items 9–10 stay dormant per post-v0 friction-driven posture.

### Addendum — patterns missed in the first pass

Surveyed after the quick-wins shipped; not in the original matrix.

| Pattern | Have it? | Notes | Effort |
|---|---|---|---|
| Export to PNG / SVG | **No** | Universal in yEd/draw.io/RF examples; `react-flow` has `toPng`/`toSvg` helpers | XS |
| Tooltips on hover | **No** | Long ids / truncated sublabels have no hover reveal | XS |
| Bend points / waypoints on orthogonal edges | **No** | draw.io's signature gesture; our `route` is one of three presets, no per-edge waypoints | M-L |
| Node resizing handles | **No** (intentional) | Sizes encode node role; deviation from yEd/draw.io is probably correct here | — |
| Snap to other nodes' edges (not just centers) | **No** | Guides match centers within `ALIGN_TOL`; edge-flush snapping is common | S |
| Outline / structure panel | **No** | yEd-style tree of nodes; probably overkill at our scale | — |
| Z-order controls (send to front/back) | **No** | Not needed until nodes overlap meaningfully | — |
| Properties inspector sidebar | **No** | Editing arbitrary `props` is piecemeal (rename, sublabel only) | M |

**Triage:** export and tooltips are the only clean "everyone has
this, it's cheap" gaps. Holding both per friction-driven posture;
neither has caused observed pain yet.



### 2026-05-03 — Implementation-pattern audit (different axis)

Reframed the industry-pattern review from *missing user features* to
*hand-rolled code that duplicates library primitives*. Scanned
`tools/topology-vscode/src/` and produced an industry-pattern audit: 19
"reimplemented" items (R1-R19) with canonical replacements + 7
"missing" react-flow/ecosystem features (M1-M7). Out-of-scope items
(Yjs, Storybook, telemetry, mobile, react-query) explicitly listed
and excluded.

Key cross-references with the deferrals memo:
- R14 (elkjs/libavoid-js routing) subsumes the deferred *auto-routing
  with obstacle avoidance* item.
- M1 (`isValidConnection`) subsumes the reject-flash quick win — no
  flash needed if the drag never starts.
- M3 (react-flow `EdgeLabelRenderer`) is a prerequisite for the
  deferred *edge display labels* item.
- R19 coordinates with deferred *snap to other nodes' edges* and
  *multi-node alignment guides*.

Audit doc is the spec for a future session; nothing landed here.
Suggested cluster order: state→zustand (R1-R3) → panels→React
(R4-R7) → geometry/routing (R14-R18, blocks on lib choice).



---

## 2026-05-27 — Load-transport collapse + 3D camera persistence (task/collapse-load-transport)

**Observation:** Two separate correctness flaws (H1 and H3) both verified fixed and working in the editor.

**H3 — Single-message load transport (order-fragility gone):** The old two-message protocol (`load` + `view-load`) let `view-load` arrive before `load`, silently dropping the view (the `_lastSpec` reorder cache only partially mitigated this). Collapsed into one `load` message and one `load(text)` store action that parses spec + `topology.json#view` together and builds flow once. Deleted `loadView`, `_lastSpec`, `view-load-noop` branch, `view-load` message variant, and host-side `sendView`. On-disk representations were already merged (single `topology.json` with a `view` key) by a prior effort; this collapsed only the in-memory transport.

**H1 — 3D camera persistence:** The old `viewerState.camera` was RF pan/zoom only — Three.js PerspectiveCamera state was never persisted. Added `Camera3D` (position + quaternion) to viewer-state schema and parser; committed on orbit/dolly/pan/roll gesture-end via `scheduleViewSave`; restored on mount, skipping auto-fit when a saved camera exists. A follow-up fixed a React effect-deps timing bug: `camera3d` arrives async after first render, so `initialCamera3d` was added to `CameraRefBridge` effect deps + `updateMatrixWorld` called to force the matrix before the skip-auto-fit guard ran.

**Outcome:** Both verified working in the editor — load preserves node positions/fade; rotate+reload restores orientation; topologies with no saved camera still auto-fit. Branch ready to merge pending sign-off.

---


## 2026-05-14 — Integration test suite (task/integrated-go-tests)

Implemented the integration test plan from `diagrams/test-plan/`. Created harness
utilities (`_fixtures.ts`, `_harness.ts`) and 5 new test files covering:
- IRG modes A5–A8 (left-alone, right-only, both-filled)
- CI fan-out B1–B2 (lockstep fan-out, seed-blocked CI)
- Lateral cascade C1–C2 (inhibit drain verified; see blocker below)
- Backpressure D1–D3 (queue holding, consume release, partial join)
- Misc E1, F1, D3-ext (sequential drain, wire seed, 3-input partial gate)

**Blocker found:** C1 single-winner mutual exclusion is not achievable with
CI.inhibitOut → IRG.right. The right-only path drains the inhibit signal, but
a subsequent left delivery fires anyway. Mutual exclusion requires inhibit
upstream of CI's own firing decision. Documented in test comment; needs design
decision.

**Go finding:** relay fires only on input fill (fill→onRun). Sequential
drain via relay requires timer advancement; E1 test uses direct input→readgate
to observe canAccept-triggered sequential delivery.

125 tests passing, all green.

---

## 2026-06-17 — pan-guide triangle: right angle on the view-aligned torus

The pan-guide right angle wasn't landing on the visible tori intersection: after the tori
were made view-aligned (horizontal-torus normal = camera up), the disk∩torus intersection and
the triangle base still used world Y, so they sat on the world equator instead of the visible
torus. Briefly tried a Thales triangle (right angle guaranteed on the circle) but the
diameter-hypotenuse made it span the whole sphere and drift off-view
([drift 1](screenshots/2026-06-17-panguide-triangle-drift-1.png),
[drift 2](screenshots/2026-06-17-panguide-triangle-drift-2.png)). Fix: use the camera-up as
the pole so the green intersection line and the compact C–Q–P right triangle (right angle at
the foot Q) land on the visible horizontal torus.

---

## 2026-07-02 — Agnostic content buffer: one binary representation both ways (merged to main)

Designed + shipped the biggest architecture change since v0: **input → Go (owns all state) →
one language-agnostic binary buffer → TS renders**. TS holds no domain state; it decodes the
buffer and forwards raw input. Built the new system behind a flag alongside the old one, then
**erased the old system 100%** in one revertible commit (23 files, ~1700 lines: `pump.ts`,
the render/spec/camera/cursor stores, the old scene-graph/beads render, `interaction-handlers`,
the `wirefold.newSystem` flag). The erase found **no hidden compile-time deps** — a clean `tsc`
with the old code gone was the proof the new system stood alone.

Brought every feature to parity on the buffer (nodes/colors/labels, beads, edges, double-links,
sphere-rings, selection incl. on-surface + edge, ports + edge-create + handhold, overlays, fade,
hover, camera). Then made **both bridges binary**: Go→TS is the content buffer (no sidecar —
label rides a buffer section, kind is a numeric column, identity is row-index); TS→Go is framed
binary records on stdin (no JSON on either wire). Logs now **decode from the buffer's EVENT
block** (Go's JSON emitter removed) — the spec's "one representation including logs."

**The friction (and the lesson):** the live-testing round exposed three persistence bugs that
GREEN UNIT TESTS HID — (1) `EnableEditPersist` computed the tree root as `""` for a FILE-form
`topologyPath` so node/anchor/overlay persisters silently no-op'd; (2) `main.tsx` still pushed
overlays to Go on `load` (empty→defaults), clobbering Go's persisted state; (3) `LoadOverlays`
skipped emitting when scene.json had no overlay keys, so the buffer streamed all-off. Each fix
"passed its test" while failing live. The reliable check turned out to be a **headless
disk-repro** — drive the built binary via framed stdin, read the actual `scene.json`/`meta.json`
bytes, two runs to prove write THEN reload-restore. Compounded by the two-process reload gap
(webview-only reload keeps the stale Go process) and the cost-overrun pattern (each blind fix was
speculative tooling on an unverified diagnosis). Captured as
`feedback_headless_repro_verifies_persistence`.

**Hardening after merge (PR #5):** turned the pain into guardrails — unified all five persisters
onto one `sceneTreeRoot`/`sceneJSONPath` resolver + a `check-scene-path-resolution` guard (the
path-drift bug class is now unrepresentable); wired a real **Go debug breadcrumb channel** to
`.probe/go-debug.jsonl` (the JSON-emitter removal had left Go-side debugging to scattered
`Fprintf`); audited `memory/` against the post-erase code (51/55 valid; rewrote the 2 the erase
invalidated). Trimmed dead surface: the `overlays attr=set` bridge record (no sender post
load-push removal) and the unexercised missed-bead/red-torus status viz (buffer columns + trace
event + render), preserving HoldNewSendOld's discard behavior (a mismatched bead is still
consumed-and-dropped, now silently — consistent with the no-back-pressure node model).

**Outcome:** merged to `main` (fast-forward), then hygiene/trims via PR #5. Both bridges binary,
no JSON anywhere, no sidecar, logs from the buffer. Verified live in the editor across the whole
feature set; persistence proven on-disk.

## 2026-07-28 — Reload gap: extension activation exonerated, respawn still slow

Measured with `tools/reload-gap.sh` because the friction was ambiguous: opening the
extension felt back to normal, but "Developer: Reload Window" did not.

The extension-host log splits those two things cleanly, and they disagree:

```
01:10:13.937  host 99435 started
01:10:14.102  Eager extensions activated   <- 165ms, healthy
01:10:20.050  host 99435 exiting with code 0
01:10:23.942  host 99617 started           <- 3.9s dead gap
```

Across five starts in that window, activation was **150–310ms** (healthy) while the
exit→start gap was **3.9s / 4.9s / 4.4s** against a **1.8s** baseline. So the cost is
entirely VS Code process respawn — nothing of ours runs in that window. Activation was
never the cause, and this is now a second, independent confirmation that the earlier
.probe-log-size / file-watcher theory is dead (watcherExclude and per-run log rotation
had already landed and did not move the number).

**A VS Code relaunch does NOT fix it.** The window measured above belongs to a VS Code
launched 12 minutes earlier (log session `20260728T010956`), and it was already at 4s.
That kills the "accumulated editor/session state" half of the standing suspicion in
`reload-gap.sh`'s header comment; the remaining suspects are machine-level:

- uptime **44 days**
- swap **2.83 GB of 4 GB** used
- load average **2.46 / 2.83 / 3.23**

**Next test: a full macOS restart**, then re-run `tools/reload-gap.sh`. If the gap
returns to ~1.8s, the regression was host memory/process pressure and there is nothing
to fix in this repo — the script's baseline note should be amended to say so.

**Baseline caveat — why this entry exists.** Of the 12 VS Code log sessions on disk,
only one still carries any exthost start/exit lines; the healthy-period logs have been
pruned. The 1.8s figure now survives *only* as a comment in `tools/reload-gap.sh`, so
there is no longer a log to diff against. These numbers are recorded here as the
pre-reboot half of the comparison.

## 2026-07-28 — Reboot fixed the reload gap; the measurement itself was wrong

Ran the restart the previous entry called for. Reload is now perceptibly instant
("less than half a second each time" across three reloads), and the gap is back to
baseline:

```
2529 exiting 01:22:16.268 -> 2677 started 01:22:17.906   1.6s
2739 exited  01:22:20.938 -> 2765 started 01:22:22.672   1.7s
```

against 3.9 / 4.9 / 4.4s pre-reboot on the same machine. Swap went **2.83 GB of 4 GB
→ 0.00 MB**. So the regression was host memory/process pressure, exactly as the
pre-reboot entry's remaining suspect list predicted, and there is nothing to fix in
this repo. `tools/reload-gap.sh`'s header now says reboot the machine, and explicitly
records that a VS Code quit+relaunch was measured and does NOT fix it.

**The script was over-reporting, and that is the durable lesson.** A short-lived
extension host can exit before flushing anything, leaving *no* `started` and *no*
`exiting` line in `exthost.log` — absent from the file entirely, not merely partial.
Pid 2739 lived 01:22:19→01:22:20 and appears only in `main.log`. The old pairing
(exthost's own `exiting` → next `started`) therefore spanned the invisible host and
printed **3.5s** for what were two ~1.7s reloads: two reloads reported as one slow
one. Fixed by taking exits from the session's `main.log`, which logs an
`exited with code` line for *every* pid, and pairing each `started` with the latest
exit before it. A gap over 60s is a window that sat closed, not a respawn, and is
dropped. The pre-reboot 3.9/4.9/4.4s figures are unchanged under the new pairing, so
the regression itself was real — but the measurement could have inflated it, and for
one row it did.

Worth noting the failure shape: a diagnostic tool that reads a log has to know which
records that log is allowed to omit. `exthost.log` is written *by* the process being
measured, so the fastest-dying instances are exactly the ones it cannot record — the
observer is blind to one end of the very distribution it exists to measure.
`main.log` is written by the supervising process and has no such gap.

## 2026-07-28 — .probe trace logs were unconditional; gated behind a setting

**Observation:** `.probe/go-edge.jsonl` reached 1.1 GB in a long session, ~30x every other
log combined.

**What it turned out to be.** There was no gate of any kind: `probePathsFor` mkdir'd
`.probe/` and armed all seven paths on every run, and every stream handler appended
unconditionally. No setting, no env var, no debug flag — diagnostic instrumentation that
shipped permanently on. Compounding it, opening the topology panel spawns Go immediately
(`extension.ts:184`) with no duration bound, so an editor left open writes hundreds of MB
with nobody watching. The volume was never a cost of *driving* the editor.

**Two corrections to the original framing**, both from measuring rather than reasoning:

1. *The rate is not a constant.* The 734 KB/s headline did not reproduce — a later
   session measured 405 lines/sec = 75 KB/s, 10x lower, same code. Volume is
   `beads_in_flight x 50/sec x ~186 bytes`, so 258 MB/hr is a saturated-ring figure.
2. *The cadence was never the problem.* One bead sampled at 20.3 ms (~50 Hz), matching
   the ~16 ms `Trace/Trace.go` documents. The multiplier is bead LIFETIME: a bead is in
   flight ~20 s, so it alone produces ~1000 lines. "Per-tick firehose" overstated it.

**The decisive finding** was that nothing consumes these events. Grepping `webview/` for
`edge-bead` returns zero hits; the renderer reads the frame's Bead block via
`edgeStream.beads(row)` (`BeadInstances.tsx:42`). The event duplicates x/y/z/value at 67
bytes against the Bead block's 16, adding only `gen` and fractional `t`, and its sole path
is `appendFileSync` to the log. So this did NOT violate "Go only emits a frame when
something changes" — a bead in flight really does change every frame. What was
unjustified was the second, larger copy that only a log file read.

**Outcome:** `wirefold.probe.trace`, default off, gating the five trace logs; error logs
always on. Verified live: 4+ minutes of running sim, zero bytes written, and `.probe/`
showing only the three error files rotated — which also proved the new code was the code
running, since the old path rotated all eight.

**The near-miss worth remembering.** The first version of the gate silently broke the
documented Go debugging channel. Breadcrumbs are not a separate file — since the
binary-buffer move they ride the per-owner streams as `kind=="breadcrumb"`, and
`probe-merge.sh --debug` greps them out of the exact four files the setting turned off. A
breadcrumb would have fired and produced nothing, with no error, so the natural conclusion
would have been "my breadcrumb is broken" rather than "the gate ate it". Fixed by making
breadcrumb rows unconditional and gating only the non-breadcrumb bulk; confirmed live with
47 real breadcrumbs surfacing through `--debug` while tracing is off.

Generalizable: before gating a channel by volume, enumerate what else rides it. The
firehose and the debug channel shared a file, and the cheap fix would have taken both.

**Declined:** removing the Go-side `KindEdgeBead` emission (proposed, then withdrawn —
~27 KB/s of discarded pipe traffic at typical bead counts, and `gen`/`t` are exactly what
you want when you *do* opt into the trace). Also declined: defaulting the sim to off. Go
streams the whole scene and TS holds no domain state (guarded), so no Go process means an
empty viewport, not a static graph.

## 2026-07-28 — Explicit upper bounds: one real leak, four ceilings, two silent failures

Follow-on from the 1.1 GB edge-stream case above, which is what motivated the sweep.

**A real unbounded leak, found by inventory and confirmed before fixing.**
`PacedWire.pending` grew forever whenever no per-edge stream was wired. `emitArrive`
appended per arrival gated on `beadPlacement.streams()` = `bp.Node != ""` — a property of
the placement's *geometry*, saying nothing about whether anyone would ever read the
events — while the only drain sat behind `writeStreamFrame`'s early return for a nil
`streamOut`. The two conditions are independent, so appends continued while nothing
drained. Measured contrast: stream unwired, 40 deliveries left all 40 queued; wired, zero.
Wider than first thought — `newEdgeMover` never sets those fields, so it covered **every
headless run**, not just the >256-edge path.

Fixed by gating production on a consumer existing (`StreamsActive`, set once at wiring
time before launch), **not** by draining unconditionally. Draining would have kept
building an event per arrival and discarding it — the same produce-for-nobody pattern
removed from the edge-bead path that morning — and would have left "who consumes this?"
hidden, which is what caused the bug.

**Ceilings added**, each with its at-the-bound decision written down rather than applied
as a policy: `maxPendingEvents`, `maxInflightBeads`, `maxPendingSends` (all panic),
`moverInboxDepth` (named but deliberately *not* asserted), `MAX_FRAME_BYTES` (+ a parity
guard against Go's `maxFrameBytes`).

Three lessons the work itself taught:

- **A diagnostic that names the wrong cause is worse than a vague one.** The first
  `maxInflightBeads` message blamed a destination that stopped draining `outCh`. True, but
  a *source outpacing the wire* reaches the same bound — that message would send a reader
  hunting a consumer that is working fine. Both messages now name every cause that
  applies, and `maxPendingSends`'s test asserts it does *not* name unresolvable
  destinations, since those are dropped rather than retained.
- **Panic is not the answer everywhere; the dividing line is who caused it.** The three
  queue ceilings panic because reaching them means a code bug. Filling a mover inbox is
  the *designed* backpressure, so it is unasserted. `MAX_EDGE_STREAMS` overflow reports
  loudly without crashing, because a large topology is legitimate input.
- **"Derive it from topology" is right for counts and wrong for depths.** The mover
  inboxes' *count* is structural (one triple per edge); their *depth* is a gesture-rate
  queue and is not in any saved file. The plan promised a derivation it could not deliver,
  and the constant now says outright that 8 is chosen.

**Two silent failures fixed, and one left deliberately.** `splitFrames` read `frameLen`
off the wire with no maximum while Go enforced one on the same protocol — a bound present
on only one side of a two-sided protocol, now guarded against re-diverging. The
`MAX_EDGE_STREAMS` overflow silently set the stream count to 0, disabling all dedicated
streams with no message; the clamp stays, the silence does not. Left alone:
`waitForCenterSettle` returns silently on a 200 ms timeout with no signal it never
settled — bounded by time, so not a missing-bound case; its defect deserves its own
decision.

**That decision, 2026-07-31: `waitForCenterSettle` STAYS AS IS.** It is not open work; do
not re-surface it. The removal was costed and declined. It could be deleted by measuring
every pair's centers ONCE up front and computing all target positions from that snapshot
(the wait exists only because `ApplyDistanceGroupTarget` re-reads centers mid-loop at
`distance_groups.go:110-111`, so a pair aims from a node an earlier pair just moved — in
"time", node 4 is the target of 2→4 and the source of 4→7/4→6). But that changes two
user-visible behaviours for one silent-timeout signal: chained pairs in "time"/"select"
stop compounding down the chain on a ▲/▼ click, and the distance panel's number would have
to come from the length dispatch REQUESTED rather than one it measures back, because
`emitViewFrame` (line 156) reads live centers and would then run before the movers commit.
Not worth it. The real deletion still rides with the standing redesign documented at line
145 — an edge carrying its own length on its own EDGE frame — which removes the ordered
loop and the wait together, and answers the VIEW-frame constraint properly.

**Guard note worth remembering:** `check-test-integrity`'s `[allow-test-weakening]` marker
is **branch-wide, not commit-scoped** — it greps every commit message in `base..HEAD`. Two
legitimate uses (both `recover()` asserting a panic, which its regex cannot distinguish
from `recover()` hiding a failure) disabled that guard for the rest of the branch. Worth
narrowing so the escape hatch closes behind itself.

**DONE — `9c886227` "Scope the test-weakening escape hatch to the files it names"
(2026-07-28), on main.** The marker now exempts only the paths it lists; a path-less
`[allow-test-weakening]` is a hard failure, and so is one naming a file the branch never
changed (a typo'd path read identically to a granted exemption). Nothing open here.

## Speed slider: the tick labels, and three wrong fixes before the right one

**2026-08-05, `task/speed-four-ticks`.** The slider grew from three positions to six
(0, ¼, ½, ¾, 1, 2) with a label under each. Getting those labels readable took four
attempts, and only the last one was aimed at the actual defect.

A screenshot of the tick row is what finally settled it (not kept — the finding below is
the durable part). Three earlier fixes were all guesses made without one:

- **"They're barely visible"** — the labels were `#ddd` at 0.5 opacity, the DARK
  overlay-panel palette, but the slider portals into `.toolbar`, which is
  `background: #fff`. Near-white on white. Check which surface a widget sits on before
  picking a colour; the two palettes in this webview are not interchangeable.
- **"A few px different, not completely different glyphs"** — the fix rebuilt each
  fraction from full-size digits around a slash, which changed how big everything looked.
  The ask was a gap, and only a gap.
- **"The ticks are the same"** — twice. Both times the deployed code WAS current (bundle
  fresh, cache-busted on mtime); the changes were simply too small to see. The useful move
  was not another adjustment but asking what was actually on screen, which established in
  one question that the new code was live and the problem was sizing.

**The real defect, from the screenshot:** ¼/½/¾ render stacked over a bar in this font,
with numerator and denominator almost touching. A precomposed glyph's internal spacing is
not adjustable, so the fraction has to be COMPOSED to make that gap settable — but at
`FRAC_EM`, the size the glyph draws its own digits at, so nothing changes size and only the
gap moves. `FRAC_GAP` is the one number the change exists to set.

**Lesson worth keeping:** a repeated "I see no change" is not a signal to adjust harder. It
is a signal to verify the change is reaching the screen, and then to ask what IS on screen.
Two rounds of pixel-guessing were spent before that question got asked.

**Guard note:** `tools/check-no-shell-source-edits.sh` (added the same day) blocked the
`cat >> session-log.md` that first tried to write this entry. Working as intended, on its
author, within the hour.
