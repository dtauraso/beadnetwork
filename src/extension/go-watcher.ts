import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "../runCommand";
import { buildBinary } from "../goBuild";
import { shouldRestartAfterBuild, TrailingDebouncer } from "../hotRestart";

export function armGoWatcher(
  repoRoot: string | undefined,
  runner: BuildAndRunRunner,
  panel: vscode.WebviewPanel,
): vscode.FileSystemWatcher | undefined {
  if (!repoRoot) return undefined;
  const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
  const goErrorsFile = path.join(repoRoot, ".probe", "go-errors.jsonl");
  const goChannel = vscode.window.createOutputChannel("topology go-build");
  const goWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(repoRoot, "**/*.go"),
  );
  const debouncer = new TrailingDebouncer(250);
  const rebuild = () => {
    debouncer.schedule(() => {
      const res = buildBinary(repoRoot, binPath);
      if (shouldRestartAfterBuild(res)) {
        goChannel.appendLine("[go] rebuilt wirefold");

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
        } catch { /* eslint-disable-line no-empty */ }
      }

    });
  };
  goWatcher.onDidChange(rebuild);
  goWatcher.onDidCreate(rebuild);
  goWatcher.onDidDelete(rebuild);

  panel.onDidDispose(() => {
    debouncer.dispose();
    goChannel.dispose();
  });
  return goWatcher;
}
