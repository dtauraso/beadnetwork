# topology-vscode — architecture map

One-screen orientation. Read this before grepping into the source tree. The
full model (bead, bead line, node goroutine, clock, bridge) lives in
[MODEL.md](MODEL.md) and [CLAUDE.md](CLAUDE.md) —
this file is only the file-layout map for this package.

## Two sides

```
extension host (Node)                webview (browser)
─────────────────────                ─────────────────
  Start/extension.ts                ◄──►   Start/main.tsx
  Start/extension/runCommand.ts                  Start/extension/webview/scene/ThreeView.tsx
  Start/extension/handle-message.ts    Start/extension/webview/scene/scene-root.tsx
  Start/extension/html.ts                          Categories/Node/node-label.ts
  Start/extension/goBuild.ts                     Categories/Scene/scene-leaves.ts
  src/* (shared)
```

esbuild bundles each side separately ([esbuild.mjs](esbuild.mjs)).
Communication is `panel.webview.postMessage` ↔ `vscode.postMessage`, wired in
`extension.ts` `panel.webview.onDidReceiveMessage`.

## Message protocol (single source of truth)

`Start/extension/messages.ts` is the shared discriminated-union source for both sides.
`WebviewToHostMsg` includes `ready` and the binary bridge envelope (a fully
encoded editor→Go record built by the concern that owns the edit — `Categories/Overlay/encode.ts`,
`Categories/Scene/encode.ts`, `Categories/Node/encode.ts` and their siblings, each beside that concern's
Go decoder — and written FRAMED to Go's stdin by `runCommand.ts`). There is no `HostToWebviewMsg` —
the host has nothing to say to the webview, because everything Go tells the
renderer is a file the renderer reads. Extension-side dispatch is
`Start/extension/handle-message.ts`. Per CLAUDE.md, Go → TS is the block
FILES and nothing else; TS → Go is framed binary records
(addressed `edit` ops, or the bare `save` command) — see CLAUDE.md
for the full bridge-surface model, not duplicated here.

**Do not restate the kind list here.** Each vocabulary is declared by the concern
whose records use it, and the TS half is GENERATED from that Go declaration, so
neither can drift and there is no second hand-kept copy to compare:

- the record kind byte — with the record it names: `Categories/Input/Drag/kinds.go` for
  raw-input, `Categories/Scene/scenerun/dispatch_edit.go` for edit-update, and
  `Categories/Scene/inputactor/stdin_reader.go` for save. Removed kind bytes are
  preserved as GAPS and never renumbered.
- the update-kind ordering — `Categories/Scene/scenerun/dispatch_edit.go`, beside the
  table that dispatches on it.
- update attributes — each concern's own `update_attrs.go` (`Node`, `Node/Edge`,
  `Scene`, `Overlay`, `Chrome/Panels/Panel`, `Speed`). An attr byte is read only
  inside the decoder its kind selected, so each concern numbers its own from zero.
- event and hit kinds — `Categories/Input/Drag/kinds.go`.
- overlay and panel flags — `Categories/Overlay/flags.ts` and
  `Categories/Chrome/Panels/Panel/flags.ts`, which both sides derive from.

Read the declaration to learn the current surface; prose copied into this file cannot
fail and so cannot be trusted.

## Extension side — what lives where

| File | Owns |
|---|---|
| `extension.ts` | `topology.openEditor` command → `createWebviewPanel`; message dispatch |
| `Start/extension/handle-message.ts` | Routes `WebviewToHostMsg` to the Go process / disk |
| `Start/extension/html.ts` | Webview HTML shell + CSP |
| `runCommand.ts` | Spawns the Go process and frames stdin records. Nothing streams back — Go inherits three stdio slots and only stderr carries anything |
| `goBuild.ts` | Compiles the Go binary; invoked automatically on `ready`, not by a button |
| registries | There is no `schema/` and no `Buffer/`: each registry lives with its concern — `node-defs.ts` in `Categories/NodeKinds/`, `wire-defs.ts` in `Categories/Scene/loadspec/`, the input codec in `Categories/Input/`, the trace events with each owner |

## Webview side

The webview is React Three Fiber (R3F) — a single 3D canvas. There are no
per-kind render components; `scene-root.tsx` draws every node/edge/bead
generically from the block files each component reads, keyed off `NODE_DEFS`
(`Categories/NodeKinds/node-defs.ts`).

| File | Role |
|---|---|
| `Start/main.tsx` | Entry point and mount. Receives NO messages — the host→webview direction is empty; everything Go says arrives as files |
| `Categories/Scene/scene-leaves.ts` | Polls the scene's binary leaves, including `spawn`, whose change means Go was replaced |
| `Start/extension/webview/scene/scene-root.tsx` | Composition root of the render tree; each component it assembles reads its own block files |
| `Start/extension/webview/scene/ThreeView.tsx` | R3F `<Canvas>` root. Holds NO gesture state — raw pointer/wheel events forward verbatim to Go's FSM (`Categories/Input/Gesture` package) |
| `Categories/Input/Drag/raw-input-build.ts` | Raw pointer/wheel + raycast hit → binary `raw-input` record to Go |
| `Categories/Overlay/overlay-flags.ts` | Read-only reflection of Go-owned overlay-toggle state (`useSyncExternalStore`; no store) |
| `webview/log/*` | Crash listeners, error boundary, log posting to the extension host |

There is no JSON-trace render path, no `pump.ts`, and no zustand/Redux-style
store — the TS layer is render + forward only (guard:
`Start/extension/webview/check-no-webview-state.sh`).

## Spec vs viewer state

- **The `topology/` tree** — read directly by the Go loader (`Categories/Scene/scenebuild/load.go`,
  `loader_tree.go`) at startup; every field maps to live wiring. Edited through `edit`
  messages. The live form is a directory tree — `nodes/<id>/base.json`, `data.json`,
  `inputs/`, `outputs/`, and `edges/*.json` (adjacency layout: an edge lives under its
  source node, `nodes/<source>/edges/<label>.json`, no top-level `edges/` dir). The
  earlier monolithic single-file `topology.json` form was deleted; the tree is the only
  supported form.
- **`<tree-root>/view/{camera,overlays,sphere}.json`** — one file per writer, for
  camera/view state not affecting generated Go. Paths computed in
  `Categories/Scene/scenepaths/scene_paths.go`. (An earlier shared sidecar under that same `view/`
  directory, a single `scene.json`, held all three in one document; it and its
  best-effort read fallback were removed once the split landed — no such file exists in
  this repo's tree.)

If a field affects generated Go, it belongs in the spec. Otherwise the sidecar.

## Build

`npm run build` → `out/extension.js` (Node CJS) + `out/webview.js` (browser
IIFE) + `out/webview.css`. Watch mode via `npm run watch`.
