// Minimal `vscode` module stub for unit tests. The extension host normally gets
// `vscode` injected by the VS Code runtime; under vitest it is unresolvable, so the
// vitest config aliases `vscode` to this file. State is mutable so a test can point
// workspaceFolders at a temp dir and inspect the output-channel calls.

export type FakeOutputChannel = {
  clear: (...a: unknown[]) => void;
  show: (...a: unknown[]) => void;
  append: (...a: unknown[]) => void;
  appendLine: (...a: unknown[]) => void;
  dispose: (...a: unknown[]) => void;
};

// Overridable factory so a test can capture channel output if it wants to.
export const window = {
  createOutputChannel(_name: string): FakeOutputChannel {
    return {
      clear() {},
      show() {},
      append() {},
      appendLine() {},
      dispose() {},
    };
  },
};

export const workspace: {
  workspaceFolders: Array<{ uri: { fsPath: string } }> | undefined;
  getConfiguration: (section?: string) => { get<T>(key: string, defaultValue?: T): T | undefined };
} = {
  workspaceFolders: undefined,
  // Minimal config stub: no setting is ever actually configured, so it always falls
  // through to the CALLER's defaultValue argument — same as real vscode.WorkspaceConfiguration
  // when a key is unset. Returning a hardcoded `undefined` here (the prior behaviour) was
  // wrong for any caller relying on a non-undefined default (e.g. wirefold.reloadOnHostBuild's
  // default-true), and happened to read as correct only for callers whose default was
  // itself falsy (wirefold.probe.trace's default-false).
  getConfiguration: () => ({ get: <T>(_key: string, defaultValue?: T): T | undefined => defaultValue }),
};

export const Uri = {
  file(p: string): { fsPath: string } {
    return { fsPath: p };
  },
};
