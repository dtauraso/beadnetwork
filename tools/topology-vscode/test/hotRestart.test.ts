// Decision-logic tests for the .go-change hot-restart feature (single-actor: pure
// functions / one debouncer instance, no cross-process communication — see
// docs/testing-shape.md).
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shouldRestartAfterBuild, TrailingDebouncer } from "../src/hotRestart";

describe("shouldRestartAfterBuild", () => {
  it("restarts on a genuinely completed successful build", () => {
    expect(shouldRestartAfterBuild({ ok: true })).toBe(true);
    expect(shouldRestartAfterBuild({ ok: true, busy: false })).toBe(true);
  });

  it("never restarts on a failed build — a broken compile must leave a live sim alone", () => {
    expect(shouldRestartAfterBuild({ ok: false, error: "boom" })).toBe(false);
  });

  it("does not restart on a coalesced (busy) result — it did not cause a build itself", () => {
    expect(shouldRestartAfterBuild({ ok: true, busy: true })).toBe(false);
  });
});

describe("TrailingDebouncer", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("coalesces a burst of schedule() calls into a single trailing invocation", () => {
    const fn = vi.fn();
    const d = new TrailingDebouncer(250);
    // Simulate several fs-watcher events for one save, or hundreds from a git checkout —
    // each schedule() call arrives before the delay elapses and must not add its own timer.
    for (let i = 0; i < 200; i++) {
      d.schedule(fn);
      vi.advanceTimersByTime(50); // well under the 250ms window
    }
    expect(fn).not.toHaveBeenCalled();
    vi.advanceTimersByTime(250);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("fires once per quiet period, not once per schedule(), across two separate bursts", () => {
    const fn = vi.fn();
    const d = new TrailingDebouncer(250);
    d.schedule(fn);
    vi.advanceTimersByTime(250);
    expect(fn).toHaveBeenCalledTimes(1);

    d.schedule(fn);
    vi.advanceTimersByTime(250);
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it("dispose() cancels a pending trailing call", () => {
    const fn = vi.fn();
    const d = new TrailingDebouncer(250);
    d.schedule(fn);
    d.dispose();
    vi.advanceTimersByTime(1000);
    expect(fn).not.toHaveBeenCalled();
  });
});
