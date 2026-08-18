# A panel is a goroutine; TS renders it

## The target

Each panel is a goroutine that owns its own state and streams what to draw. TS positions
and paints it and nothing else: it reads no domain state, derives no rows, and holds no
panel state.

A panel is 2D and camera-independent. Go emits its geometry in screen terms, so there is no
projection to do and no camera action that can move it.

## Why

The panels are the last place the render tree reads domain state in order to decide what to
show. `readNodeRuleRows` walks every node AND every edge to build the rules panel's rows;
`readTiltVectorRows` walks every node. That is TS deriving the model, which the drift rule
excludes, and it is the only reason `useSyncExternalStore` (11 call sites), the frame ticker
and two module-level derived caches exist at all.

## The precedent to copy

Node labels already work this way. Go emits `LabelAnchor` columns, `BufferLabelProjector`
turns them into pixels, and `BufferLabelOverlay` renders absolutely-positioned DOM. TS
decides neither the text nor which labels exist.

A panel is the same shape minus the projection, because its coordinates do not depend on the
camera. **Text stays DOM.** Nothing here requires canvas text rendering; if a panel is later
drawn in the canvas that is a separate change with its own cost.

## What a panel goroutine owns

- its open/closed state — already Go's (`viewstate.PanelState`, streamed on the Panel block)
- its rows: which exist, their labels, their values, their enabled/checked state
- its screen rect, and each row's rect

One panel is one owner, so one stream, with one column per thing and run-valued columns
where the row count varies — the shape the tilt arrows and edge beads already use.

## The model decision to settle BEFORE any code

Today a click knows WHICH control it hit because TS built the DOM, and each panel sends an
addressed edit (`encodeClockSpeed`, `encodeOverlaysToggle`, …). Once Go owns the layout, TS
knows only WHERE the pointer went. So panel input becomes raw-input plus hit-testing in Go,
like node picking — not an addressed edit per control.

That is a real change to the TS → Go vocabulary in `.claude/rules/bridge-surface.md`, and it
should be agreed before the first panel moves, not discovered during it.

## Order

1. **SpeedSlider** — one value, one control, no rows. Establishes the per-panel stream, the
   screen-rect columns and the hit-test path end to end, on the smallest possible surface.
2. **TiltVectorButtons** — rows that vary with node count; the first run-valued panel columns.
3. **NodeRulesPanel** — the largest, and the one whose rows are derived from nodes AND edges
   today; deletes `readNodeRuleRows` and its cache.
4. **AngleDropdown**, **NodesDropdown**, **OverlaysDropdown**.
5. Delete what is then unused: the 11 hooks, `src/webview/frame-tick.ts`, both derived
   caches, and the `check-no-webview-state.sh` allowances that name those files.

Stop after 1 and judge it. Six panels is a lot of surface to move on an unproven shape.

## Verification

- **Headless**, per panel: drive the real binary and read the panel's columns off the wire —
  its rows and rects — the same way the bead and tilt columns were checked. A silent column
  and a clean build look identical to the guards.
- **In the editor**: the panel renders, does not move under orbit or zoom, and a click
  reaches Go — a breadcrumb on the panel goroutine naming the row it hit.

## What breaks

- Panel input stops being an addressed edit per control (see the model decision above).
- `check-no-webview-state.sh` currently ALLOWS `useSyncExternalStore` in the named panel
  files. As each panel moves, its name comes off that list, and the guard should fail if a
  hook reappears.
- The Panel block already streams open/closed. A per-panel stream must not carry a second
  copy of it.
- Panel goroutines are a new kind of owner — not a node, edge or view. Stream fds are sized
  from `counts/` before Go starts; panels are a FIXED set, so they need their own base fd
  decided once rather than a count read from disk.

## Risks

- **Coordinate mismatch.** Go hit-tests against rects TS renders. If the two disagree about
  device pixels vs CSS pixels, or about the canvas origin, clicks land on the wrong row and
  it will look like a logic bug.
- **A panel is interactive state mid-gesture** — a slider being dragged, a dropdown open with
  the pointer down. That state belongs to the panel goroutine, and the round trip is now over
  the bridge; a slider that lags its own drag will be the first thing anyone notices.
- **Six panels, one shape.** The value has to be proven by the first one.
