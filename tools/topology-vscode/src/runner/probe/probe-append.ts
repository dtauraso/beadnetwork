import * as fs from "fs";
import { decodeBufferLog, decodeStreamFrameEvents } from "../../buffer-log";
import { decodeNodeStreamFrame } from "../../webview/three/decode/buffer-decode-node";
import { decodeEdgeStreamFrame } from "../../webview/three/decode/buffer-decode-edge";
import { decodeInteriorStreamFrame } from "../../webview/three/decode/buffer-decode-interior";

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

export function appendEdgeProbe(probeFile: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  if (!probeFile) return;
  const decoded = decodeEdgeStreamFrame(row, ab);
  if (!decoded || decoded.eventCount === 0) return;
  appendLines(probeFile, decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !probeTrace));
}

export function appendNodeProbe(probeFile: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  if (!probeFile) return;
  const decoded = decodeNodeStreamFrame(row, ab);
  if (!decoded || decoded.eventCount === 0) return;
  appendLines(probeFile, decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !probeTrace));
}

export function appendInteriorProbe(probeFile: string | undefined, row: number, ab: ArrayBuffer, probeTrace: boolean): void {
  if (!probeFile) return;
  const decoded = decodeInteriorStreamFrame(row, ab);
  if (!decoded || decoded.eventCount === 0) return;
  appendLines(probeFile, decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !probeTrace));
}
