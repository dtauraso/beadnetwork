# wirefold

## The vision

An animated network in three dimensions, modeling `1 * 2 = 2 * 1`.

Arithmetic is normally written on a line and evaluated a step at a time, or drawn as a flat diagram that holds still. Here each side of the equation is a network in 3D space, where multiplication is expressed through geometry and timing rather than as a step to execute. Running the network plays out the arithmetic. Equality is a structure too. Its own nodes connect both sides into one network.

## What this is

Two things in one repo:

1. **A concurrent dataflow runtime in Go.** Behavior emerges from how nodes are wired together, not from procedural code. Goroutines and channels replace conventional control flow.

2. **A visual editor** (vscode webview, Three.js / React Three Fiber). The diagram is the spec for **topology/wiring** — interpreted data, no codegen step on that path: the editor writes a directory tree of `topology/nodes/<id>/base.json`, `inputs|outputs/*.json`, and `topology/nodes/<id>/edges/*.json` (an adjacency list — an edge lives under its source node, no top-level `edges/` dir), which the runtime loader reads directly at startup. (Node-kind behavior and the content-buffer schema are a *separate*, code-generated axis — `Categories/NodeKinds/*/SPEC.md` and the block-file layouts drive `gen-node-defs`, staleness-guarded by `check-generated.sh`.) The directory tree is the only supported form — the earlier monolithic `topology.json` form was deleted.

## Running it

Everything starts in [Categories/Start/](Categories/Start/) — the Go binary, the
extension host, and the webview all have their entry point there.

```bash
go build ./...                 # compile every package
go run ./Categories/Start      # run the network
```

```bash
npm run build       # out/extension.js and out/webview.js
npm run watch       # rebuild both on change
```

## Layout

`Categories/` holds the code. Each directory under it is one category, holding
everything about that one thing — the Go that runs it and the TS that draws it,
side by side.

```
Categories/
  Start/        the entry points: main.go, extension.ts, main.tsx
  Node/         nodes, their beads, edges, geometry
  Scene/        the scene: its spec on disk, assembling it, running it
  Chrome/       the UI that is not the diagram
  extension/    the VS Code integration, including webview/
  Camera/  Clock/  Input/  NodeKinds/  Overlay/  Polar/  Ring/  RingPoint/

scripts/        what serves the repo rather than one category
docs/  memory/  and the tool configs stay at the root, where each
                toolchain looks for its own
```

[EDITOR.md](EDITOR.md) has the vscode extension build/run instructions.

## License

See [LICENSE](LICENSE).
