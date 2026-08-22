# Topology Editor (VS Code extension)

Editor for the `topology/` directory tree. Drag nodes around; saves write back
through `WorkspaceEdit` and the runtime loader picks up the new spec on next run.

It is a **command-activated webview panel**, not a custom editor bound to a file
type — opening a `.json` file will not launch it.

---

# Setup from a brand-new Mac

If you already have Go, Node, VS Code and a clone of the repo, skip to
[Build](#build). Otherwise see [SETUP.md](SETUP.md) — no prior setup assumed.

---

# Build

From the repo root (`~/Documents/wirefold`):

```sh
npm install
npm run build
```

`npm install` takes a minute or two and prints a lot of text; that is normal.

The build step regenerates node definitions using the Go toolchain, so it fails
with `go: command not found` if Go is missing — that error means the prerequisite
is absent, not that the build is broken.

## Run it

```sh
code .
```

That opens this folder in VS Code. Press **F5**. A second VS Code window opens
(the Extension Development Host). In that new window:

1. It already has this repo open as its folder — the launch config passes
   `${workspaceFolder}` for exactly that. The extension resolves everything
   relative to the FIRST workspace folder, so if you open a subfolder instead it
   fails to find the topology tree.
2. In the file list on the left, right-click the `topology` folder →
   **Topology: Open Editor**. (Or `Cmd`+`Shift`+`P` → **Topology: Open Editor**,
   which resolves `topology/` from the workspace root.)

The panel compiles the Go binary on open and starts streaming. First open is
slower because it compiles from scratch.

## Rebuilding while you work

`npm run watch` rebuilds the webview bundle on save. What you do after a rebuild
depends on which side changed:

- **webview code** (`extension/webview/`) — close and reopen the panel.
- **extension-host code** (everything else in `src/`) — `Cmd`+`Shift`+`P` →
  **Developer: Reload Window** in the Extension Development Host.

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for the file map (extension side,
webview side, message protocol, spec-vs-viewer split).

## Packaging a .vsix

To install into your normal VS Code instead of running the dev host. The
`package` script calls `vsce`, which is not in `devDependencies`, so run it
through `npx`:

```sh
npx --yes @vscode/vsce package   # → topology-vscode-0.0.1.vsix
code --install-extension topology-vscode-0.0.1.vsix
```

The installed extension still needs Go on your PATH and the repo open as the
workspace folder.
