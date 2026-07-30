// Decision logic for the extension-host-bundle hot-reload feature (extension.ts's
// hostWatcher). Split out from extension.ts for the same reason hotRestart.ts is split
// out (docs/testing-shape.md): this is single-actor decision logic (a hash comparison, a
// settings read), not two goroutines/processes communicating — extension.ts itself is
// wiring that cannot be driven headlessly (it needs a live vscode.ExtensionContext and a
// real reload command).
import * as crypto from "crypto";
import * as vscode from "vscode";

// Single source of truth for the settings key gating this feature, same pattern as
// probe-files.ts's PROBE_TRACE_SETTING_SECTION/KEY — no other file should embed the
// "reloadOnHostBuild" (or "wirefold") literal.
export const HOST_RELOAD_SETTING_SECTION = "wirefold";
export const HOST_RELOAD_SETTING_KEY = "reloadOnHostBuild";

/** Whether the window should self-reload when the built extension-host bundle
 *  (out/extension.js) changes. Default TRUE, unlike wirefold.probe.trace's default-off:
 *  the user asked for this behaviour, and a stale host silently drops frames from a
 *  rebuilt Go/TS pair in a way that reads as "the code never ran" (memory/
 *  feedback_two_process_editor_reload.md) — the failure mode this feature exists to
 *  remove is worse than an occasional unwanted reload. Read at the point of use (not
 *  cached at activation) so toggling the setting takes effect without itself needing a
 *  reload. */
export function isHostReloadEnabled(): boolean {
  return vscode.workspace
    .getConfiguration(HOST_RELOAD_SETTING_SECTION)
    .get<boolean>(HOST_RELOAD_SETTING_KEY, true);
}

/** Content hash of a built bundle, used instead of mtime so a watcher firing on a
 *  write that doesn't change the bytes (e.g. a `touch`, or a build tool rewriting the
 *  file with identical output) doesn't trigger a reload — see shouldReloadHost. */
export function hashBundle(bytes: Buffer): string {
  return crypto.createHash("sha256").update(bytes).digest("hex");
}

// shouldReloadHost decides whether a just-rebuilt host bundle warrants reloading the
// window. `loadedHash` is the hash of the bundle THIS running extension-host instance
// was actually loaded from (captured once, at activation); `newHash` is the hash of the
// bundle on disk right now. This is also the loop guard (requirement 4): a reload always
// respawns the extension host, which re-activates and re-captures loadedHash from
// whatever is on disk at THAT moment — so the new instance's own baseline already equals
// the bundle it is running, and nothing in activation writes out/extension.js itself, so
// there is no content change left for the new instance to react to until a REAL rebuild
// happens. A loop would require activation to write its own bundle, which it does not.
//
// loadedHash undefined means activation couldn't hash its own bundle (e.g. a read raced
// a build) — with no trustworthy baseline to compare against, this returns false rather
// than guessing, since a false positive here means reloading against nothing new.
export function shouldReloadHost(loadedHash: string | undefined, newHash: string): boolean {
  return loadedHash !== undefined && newHash !== loadedHash;
}
