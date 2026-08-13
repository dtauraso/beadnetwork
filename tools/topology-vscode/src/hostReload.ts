import * as crypto from "crypto";
import * as vscode from "vscode";

export const HOST_RELOAD_SETTING_SECTION = "wirefold";
export const HOST_RELOAD_SETTING_KEY = "reloadOnHostBuild";

export function isHostReloadEnabled(): boolean {
  return vscode.workspace
    .getConfiguration(HOST_RELOAD_SETTING_SECTION)
    // Default OFF, matching package.json. An automatic reload seizes the window
    // on every rebuild; the stale-code risk it existed for is now caught by the
    // layout handshake in the VIEW frame, which reports instead of reloading.
    .get<boolean>(HOST_RELOAD_SETTING_KEY, false);
}

export function hashBundle(bytes: Buffer): string {
  return crypto.createHash("sha256").update(bytes).digest("hex");
}

export function shouldReloadHost(loadedHash: string | undefined, newHash: string): boolean {
  return loadedHash !== undefined && newHash !== loadedHash;
}
