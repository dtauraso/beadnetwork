import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "./runCommand";
import { buildBinary } from "./goBuild";
import { shouldRestartAfterBuild, TrailingDebouncer } from "./hotRestart";
import { hashBundle, isHostReloadEnabled, shouldReloadHost } from "./hostReload";
import type { HostToWebviewMsg } from "./messages";
import { buildWebviewHtml } from "./extension/html";
import { handleMessage } from "./extension/handle-message";
import { PROBE_FILES } from "./probe-files";

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.commands.registerCommand("topology.openEditor", (uri?: vscode.Uri) => {
      openTopologyEditor(context, uri);
    }),
  );
  armHostReloadWatcher(context);
}

// armHostReloadWatcher watches the BUILT extension-host bundle (out/extension.js — the
// esbuild `extension` output, see esbuild.mjs's outfile; distinct from out/webview.js,
// which bundleWatcher below already handles per-panel) and self-reloads the window when
// it changes. This closes the last of the editor's three refresh paths: webview bundle
// rebuilt -> tab refreshes itself (bundleWatcher), Go rebuilt -> live sim hot-restarts
// (goWatcher/runner.restart()), and now extension-host rebuilt -> window reloads itself.
// Without this a rebuilt host runs stale code until "Developer: Reload Window" is run by
// hand, which produces WRONG DEBUGGING: Go can emit something the stale host's fd
// allocation or message handling doesn't know about, and the symptom is silence —
// indistinguishable from "the code never ran" (memory/feedback_two_process_editor_reload.md).
//
// Deliberately armed once in activate(), not per-panel like bundleWatcher/goWatcher: the
// host bundle is a property of the WHOLE extension process, not of any one editor tab, so
// one watcher for the extension's lifetime is the right scope (and avoids N watchers/N
// reload commands if the user has multiple topology tabs open).
function armHostReloadWatcher(context: vscode.ExtensionContext): void {
  const hostBundlePath = path.join(context.extensionPath, "out", "extension.js");
  const hostChannel = vscode.window.createOutputChannel("topology host-reload");
  context.subscriptions.push(hostChannel);
  // Baseline: the hash of the bundle THIS instance actually loaded. Captured once, here,
  // never rewritten except by a fresh activate() after a real reload — see
  // shouldReloadHost's doc comment for why that makes a self-triggered loop impossible.
  let loadedHash: string | undefined;
  try {
    loadedHash = hashBundle(fs.readFileSync(hostBundlePath));
  } catch {
    // Bundle unreadable at activation (shouldn't happen once the extension is running
    // from it, but a race with a build is possible) — shouldReloadHost treats undefined
    // as "no baseline, never reload" rather than guessing.
  }
  const hostWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "extension.js",
    ),
  );
  context.subscriptions.push(hostWatcher);
  const debouncer = new TrailingDebouncer(250);
  context.subscriptions.push({ dispose: () => debouncer.dispose() });
  // reloading latches once a reload has been requested, so a second debounced firing in
  // the window before "workbench.action.reloadWindow" actually tears the process down
  // (VS Code takes a beat) can't queue a second reload command.
  let reloading = false;
  const maybeReload = () => {
    debouncer.schedule(() => {
      if (reloading) return;
      // Read at point of use (not cached) so toggling the setting takes effect without
      // itself needing a reload — see isHostReloadEnabled's doc comment.
      if (!isHostReloadEnabled()) return;
      let newHash: string;
      try {
        newHash = hashBundle(fs.readFileSync(hostBundlePath));
      } catch {
        return; // Build in progress / file mid-write — the next fs event will retry.
      }
      if (!shouldReloadHost(loadedHash, newHash)) return;
      // NOTE (requirement 5): there is no cheap "a gesture is in flight" signal
      // reachable from the extension host to defer this against. Pointer/drag state
      // lives entirely in Go's gesture FSM (nodes/Wiring/gesture.go), reached only via
      // fire-and-forget raw-input on Go's stdin (CLAUDE.md's bridge surface) — the host
      // never learns whether a drag is mid-flight, so there is nothing here to poll or
      // wait on. Documented as a known rough edge rather than inventing a signal that
      // doesn't exist: a reload during an active drag is possible and will feel abrupt.
      reloading = true;
      hostChannel.appendLine("[topology] extension host bundle changed — reloading window");
      vscode.window.setStatusBarMessage(
        "Topology: extension updated, reloading window…",
        3000,
      );
      // Fire-and-forget: reloadWindow tears this process down: there is nothing
      // meaningful to await from the code that is about to stop existing.
      void vscode.commands.executeCommand("workbench.action.reloadWindow");
    });
  };
  hostWatcher.onDidChange(maybeReload);
  hostWatcher.onDidCreate(maybeReload);
}

// Truncate probe logs on each editor open so each session's trace is clean (cross-session accumulation was misleading).
function resetProbeLogs(repoRoot: string): void {
  try {
    const probeDir = path.join(repoRoot, ".probe");
    fs.mkdirSync(probeDir, { recursive: true });
    // Iterate the canonical registry (not a hand-typed list) so a newly-added probe file is
    // reset automatically — the omission that let go-debug.jsonl accumulate across sessions
    // cannot recur. ALL files reset unconditionally, including the four Go trace files:
    // they always receive DEBUG BREADCRUMB rows regardless of wirefold.probe.trace (see
    // buffer-log.ts's breadcrumbsOnly filtering), so they are live logs either way and must
    // reset like every other probe file.
    for (const name of Object.values(PROBE_FILES)) {
      fs.writeFileSync(path.join(probeDir, name), "");
    }
  } catch {
    // Swallow: logging reset must never block opening the editor.
  }
}

function openTopologyEditor(context: vscode.ExtensionContext, folderUri?: vscode.Uri): void {
  // Resolve topology folder path. Command can be invoked from explorer context
  // menu (folderUri is the topology/ dir) or command palette (no uri).
  let topologyPath: string | undefined;
  if (folderUri) {
    topologyPath = folderUri.fsPath;
  } else {
    // Fallback: find topology/ dir in workspace root
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (folder) {
      const candidate = path.join(folder.uri.fsPath, "topology");
      if (fs.existsSync(candidate)) topologyPath = candidate;
    }
  }

  // Reset probe logs early: same workspace root the runner (.probe/go*.jsonl) and
  // appendWebviewLog (.probe/ts*.jsonl) write to, before any log can be appended.
  const probeRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  if (probeRoot) resetProbeLogs(probeRoot);

  const panel = vscode.window.createWebviewPanel(
    "topology.editor",
    "Topology Editor",
    vscode.ViewColumn.One,
    {
      enableScripts: true,
      retainContextWhenHidden: true,
      localResourceRoots: [vscode.Uri.file(path.join(context.extensionPath, "out"))],
    },
  );
  panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath);

  // Fire-and-forget host→webview send (bridge doctrine: no await, no Promise
  // chain — see check-no-await-on-bridge). `void` discards the postMessage
  // Thenable so this returns void and can be passed where VS Code expects a
  // void-returning callback, without floating the promise.
  const post = (msg: HostToWebviewMsg): void => {
    void panel.webview.postMessage(msg);
  };
  const runner = new BuildAndRunRunner(
    // buffer-snapshot frames: forward each to the webview verbatim. Without this
    // wiring the runner decodes each per-owner stream (handleViewFd/handleEdgeFd/
    // handleNodeFd/handleInteriorFd) but drops every frame, so the new-system
    // BufferScene (which polls getLatestSnapshot each frame) never receives node/edge/
    // camera/bead geometry and renders nothing. Fire-and-forget host→webview send.
    (snapshot) => post(snapshot),
  );


  // Hot-reload of the webview bundle. Armed in every extension mode, not just
  // Development: gating it on extensionMode was self-defeating. In a real
  // install out/webview.js never changes, so an always-armed watcher costs one
  // idle inotify handle and fires never; but the case that actually matters —
  // a developer rebuilding the bundle while the editor tab is open, INCLUDING
  // against an installed extension rather than an F5 dev host — is precisely
  // what the gate suppressed, forcing a manual tab reload. This is safe
  // because the webview holds no domain state (render-and-forward-only seam,
  // guard: check-no-webview-state.sh): Go re-streams and main.tsx re-posts
  // "ready" on remount, so a refreshed tab re-learns everything.
  const bundleWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "webview.js",
    ),
  );
  {
    console.log("[topology] bundleWatcher armed for", path.join(context.extensionPath, "out", "webview.js"));
    let pending: NodeJS.Timeout | undefined;
    const reload = (kind: string) => () => {
      console.log("[topology] bundleWatcher fired:", kind);
      if (pending) clearTimeout(pending);
      pending = setTimeout(() => {
        console.log("[topology] hot-reload: re-rendering webview.html");
        panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath);
      }, 150);
    };
    bundleWatcher.onDidChange(reload("change"));
    bundleWatcher.onDidCreate(reload("create"));
  }

  // Eager Go-binary watcher: rebuild the prebuilt binary the moment a .go file is saved so
  // launches stay instant (the lazy ensureBinaryBuilt in runner.run() remains the safety
  // net for missed events). If a sim is LIVE when a rebuild SUCCEEDS, it also hot-restarts
  // that sim (runner.restart(), which is a no-op — requirement 1 — if nothing is running)
  // so the new geometry/behaviour is on screen with no window reload and no user action.
  // Debounced (TrailingDebouncer) so one save, or a multi-file edit/checkout touching
  // hundreds of .go files, produces at most one rebuild and one restart, not one per event.
  const repoRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  let goWatcher: vscode.FileSystemWatcher | undefined;
  if (repoRoot) {
    const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
    const goErrorsFile = path.join(repoRoot, ".probe", "go-errors.jsonl");
    const goChannel = vscode.window.createOutputChannel("topology go-build");
    goWatcher = vscode.workspace.createFileSystemWatcher(
      new vscode.RelativePattern(repoRoot, "**/*.go"),
    );
    const debouncer = new TrailingDebouncer(250);
    const rebuild = () => {
      debouncer.schedule(() => {
        const res = buildBinary(repoRoot, binPath);
        if (shouldRestartAfterBuild(res)) {
          goChannel.appendLine("[go] rebuilt wirefold");
          // Only restarts a LIVE sim (runner.restart() no-ops otherwise — requirement 1);
          // reuses the topology path the live run was already started with (runner owns
          // that, restart() never takes one — requirement 2). Told to the user here so a
          // sim that silently changes under someone mid-drag isn't confusing (requirement 7).
          if (runner.restart()) {
            goChannel.appendLine("[go] hot-restarting sim");
          }
        } else if (!res.ok) {
          goChannel.appendLine(`[go] build error: ${res.error}`);
          try {
            fs.mkdirSync(path.dirname(goErrorsFile), { recursive: true });
            fs.appendFileSync(
              goErrorsFile,
              JSON.stringify({ ts_ms: Date.now(), src: "go", kind: "error", message: res.error }) + "\n",
              "utf8",
            );
          } catch { /* swallow */ }
        }
        // else: res.ok && res.busy — coalesced against another in-flight build (see
        // shouldRestartAfterBuild's doc comment: this caller did not cause a build, so
        // its result says nothing about whether THIS edit's changes are in the binary
        // yet). Nothing to report; skip the restart rather than restart against a binary
        // that might still be one edit stale.
      });
    };
    goWatcher.onDidChange(rebuild);
    goWatcher.onDidCreate(rebuild);
    goWatcher.onDidDelete(rebuild);
    // goWatcher/goChannel/debouncer track THIS panel's lifetime, so the panel is their
    // single disposal owner (onDidDispose below). Deliberately NOT pushed into
    // context.subscriptions — mirrors the bundleWatcher single-owner contract and
    // avoids a double-dispose across the two owners.
    panel.onDidDispose(() => {
      debouncer.dispose();
      goChannel.dispose();
    });
  }

  context.subscriptions.push(runner);
  // bundleWatcher tracks THIS panel's lifetime, so the panel is its single disposal
  // owner (onDidDispose below). It is deliberately NOT pushed into
  // context.subscriptions to avoid a muddled double-owner contract.

  panel.onDidDispose(() => {
    bundleWatcher?.dispose();
    goWatcher?.dispose();
    runner.dispose();
  });

  panel.webview.onDidReceiveMessage((raw) => {
    const workspaceFolder = folderUri ? vscode.workspace.getWorkspaceFolder(folderUri) : undefined;
    // Final fallback is undefined (no real workspace) — appendWebviewLog skips the
    // write rather than misdirecting .probe/ logs to an arbitrary cwd.
    const logUri = workspaceFolder?.uri ?? folderUri ?? vscode.workspace.workspaceFolders?.[0]?.uri;
    void handleMessage(raw, { logUri, runner, post }).catch((err: unknown) => {
      console.error("topology: handleMessage failed", err);
    });
  });

  // Spawn Go immediately; the render path is buffer-only (buffer-snapshot on
  // fd3) so there is nothing else to send on "ready".
  // The one USER-started spawn, so the only one that reveals the output panel.
  runner.run(topologyPath, { reveal: true });
}
