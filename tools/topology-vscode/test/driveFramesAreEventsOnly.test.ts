// driveFramesAreEventsOnly.test.ts — a drive-slot frame must never become a node's
// interior state.
//
// THE BUG THIS PINS. Reported from the editor: a pulse node's held bead appears, is there
// for about half a second, and is then gone with nothing in its place.
//
// A node's interior slots are authored by ONE goroutine — its own Update loop, via
// emitHeldBead / emitNodeBeads / emitInputBeads. A DriveHeld goroutine writes on a separate
// fd (one goroutine, one fd) whose frames are interior-SHAPED but author no slot state:
// interiorStream.WriteEvents reuses THAT stream's own lastPresent, which nothing ever sets,
// so every drive frame carries an all-absent snapshot. The host relayed those frames under
// the same node row as the real interior frames, and the webview's last-writer-wins cell
// took them — erasing the held bead a fraction of a second after it appeared, because the
// drive goroutines pulse at the wire cadence while a held value changes every few seconds.
//
// The frames are still decoded and probe-logged; they are simply not interior state.

import { describe, it, expect } from "vitest";
import { BuildAndRunRunner } from "../src/runCommand";

/** Reach the private relay tail directly: this is a unit test of the routing decision, not
 *  of process spawning, and driving a real child would test the OS rather than this rule. */
function relay(runner: BuildAndRunRunner, row: number, frames: ArrayBuffer[], assertsSlots: boolean) {
  (runner as unknown as {
    processInteriorLikeFrames(row: number, frames: ArrayBuffer[], assertsSlots: boolean): void;
  }).processInteriorLikeFrames(row, frames, assertsSlots);
}

const frame = (marker: number) => new Uint8Array([marker, 0, 0, 0]).buffer;

describe("drive-slot frames are events only", () => {
  it("does not post a drive frame to the webview as interior state", () => {
    const posted: unknown[] = [];
    const runner = new BuildAndRunRunner((m) => posted.push(m));

    relay(runner, 4, [frame(1)], false); // a DriveHeld goroutine's frame

    expect(posted).toEqual([]);
  });

  it("still posts the node's own interior frame", () => {
    const posted: unknown[] = [];
    const runner = new BuildAndRunRunner((m) => posted.push(m));

    relay(runner, 4, [frame(2)], true); // the node's own Update-loop frame

    expect(posted).toHaveLength(1);
  });

  // The live sequence: the node emits its held bead, then a drive goroutine pulses. The
  // bead must survive the pulse — this is the half-second disappearance, in miniature.
  it("leaves a held bead intact when a drive pulse follows it", () => {
    const posted: unknown[] = [];
    const runner = new BuildAndRunRunner((m) => posted.push(m));

    relay(runner, 4, [frame(2)], true); // held bead appears
    relay(runner, 4, [frame(1)], false); // drive pulse, moments later

    // Exactly one frame reached the webview: the held bead. Nothing overwrote it.
    expect(posted).toHaveLength(1);
  });

  // The replay cache feeds a remounting webview. A drive frame in it would blank the node
  // on reopen, which is the same defect one reload later.
  it("keeps drive frames out of the replay cache", () => {
    const runner = new BuildAndRunRunner(() => {});

    relay(runner, 4, [frame(2)], true);
    relay(runner, 4, [frame(1)], false);

    const cached = runner.getLastInteriorFrames();
    expect(cached).toHaveLength(1);
    expect(new Uint8Array(cached[0]!.buffer)[0]).toBe(2);
  });
});
