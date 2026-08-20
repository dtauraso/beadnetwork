import * as fs from "fs";
import { decodeBufferLog, decodeStreamFrameEvents } from "../../buffer-log";
import { decodeNodeStreamFrame } from "../../../Node/buffer-decode-node";
import { decodeEdgeStreamFrame } from "../../../Node/Edge/buffer-decode-edge";
import { decodeBeadStreamFrame } from "../../../Node/BeadAnimation/buffer-decode-bead";
import { decodeInteriorStreamFrame } from "../../../Node/Interior/buffer-decode-interior";
import { probeOwnerFile, type ProbeOwner } from "../../probe-files";
import type { DecodedEvents } from "../../../webview/decode/buffer-decode-shared";

function appendLines(probeFile: string, lines: string): void {
  if (lines.length === 0) return;
  try {
    fs.appendFileSync(probeFile, lines, "utf8");
  } catch { /* eslint-disable-line no-empty */ }
}

export function appendViewProbe(probeFile: string | undefined, ab: ArrayBuffer, probeTrace: boolean): void {
  if (!probeFile) return;
  appendLines(probeFile, decodeBufferLog(ab, !probeTrace));
}

function appendOwnerProbe(
  probeDir: string | undefined, owner: ProbeOwner, row: number,
  decoded: { events: DecodedEvents } | null,
  probeTrace: boolean,
): void {
  if (!probeDir || !decoded) return;
  appendLines(
    probeOwnerFile(probeDir, owner, row),
    decodeStreamFrameEvents(decoded.events, !probeTrace),
  );
}

export function appendEdgeProbe(probeDir: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  appendOwnerProbe(probeDir, "edge", row, decodeEdgeStreamFrame(row, ab), probeTrace);
}

export function appendNodeProbe(probeDir: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  appendOwnerProbe(probeDir, "node", row, decodeNodeStreamFrame(row, ab), probeTrace);
}

export function appendBeadProbe(probeDir: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  appendOwnerProbe(probeDir, "bead", row, decodeBeadStreamFrame(row, ab), probeTrace);
}

export function appendInteriorProbe(probeDir: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  appendOwnerProbe(probeDir, "interior", row, decodeInteriorStreamFrame(row, ab), probeTrace);
}
