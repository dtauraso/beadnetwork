// Decision-logic tests for the extension-host-bundle reload feature (single-actor: pure
// hash comparison + a settings read, no cross-process communication — see
// docs/testing-shape.md). Matches hotRestart.test.ts's shape.
import { describe, it, expect, afterEach } from "vitest";
import { workspace } from "vscode";
import { hashBundle, isHostReloadEnabled, shouldReloadHost } from "../src/hostReload";

describe("shouldReloadHost", () => {
  it("reloads when the on-disk bundle hash differs from the loaded baseline", () => {
    expect(shouldReloadHost("aaa", "bbb")).toBe(true);
  });

  it("does not reload when the hash is unchanged (a write that didn't change bytes)", () => {
    expect(shouldReloadHost("aaa", "aaa")).toBe(false);
  });

  it("does not reload with no baseline — activation couldn't hash its own bundle", () => {
    expect(shouldReloadHost(undefined, "bbb")).toBe(false);
  });
});

describe("hashBundle", () => {
  it("is deterministic for identical bytes and differs for different bytes", () => {
    const a = Buffer.from("same content");
    const b = Buffer.from("same content");
    const c = Buffer.from("different content");
    expect(hashBundle(a)).toBe(hashBundle(b));
    expect(hashBundle(a)).not.toBe(hashBundle(c));
  });
});

describe("isHostReloadEnabled", () => {
  const original = workspace.getConfiguration;
  afterEach(() => {
    workspace.getConfiguration = original;
  });

  it("defaults to true when the setting is unset", () => {
    // The vscode test stub's default getConfiguration().get() returns undefined for
    // every key; isHostReloadEnabled must supply true as ITS OWN default (unlike
    // wirefold.probe.trace, which defaults off) rather than inheriting the stub's
    // undefined-means-false shape.
    expect(isHostReloadEnabled()).toBe(true);
  });

  it("reads an explicit false from settings", () => {
    workspace.getConfiguration = () => ({ get: () => false }) as ReturnType<typeof workspace.getConfiguration>;
    expect(isHostReloadEnabled()).toBe(false);
  });

  it("reads an explicit true from settings", () => {
    workspace.getConfiguration = () => ({ get: () => true }) as ReturnType<typeof workspace.getConfiguration>;
    expect(isHostReloadEnabled()).toBe(true);
  });
});
