import * as path from "path";
import * as vscode from "vscode";
import { realPath } from "./html";
import { sceneRoots } from "./scene-roots";

export function webviewOptions(extensionPath: string, anchorPath: string): vscode.WebviewOptions {
  return {
    enableScripts: true,
    localResourceRoots: [
      vscode.Uri.file(path.join(realPath(extensionPath), "out")),
      vscode.Uri.file(path.join(realPath(extensionPath), "Categories")),
      ...sceneRoots(anchorPath).map((dir) => vscode.Uri.file(dir)),
    ],
  };
}
