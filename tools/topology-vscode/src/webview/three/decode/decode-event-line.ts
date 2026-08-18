import { TRACE_EVENT_KINDS, BREADCRUMB_LABELS } from "../../../../Trace/trace-kinds";
import { nodeLabel } from "./buffer-decode-node";
import {
  readEventKind, readEventNodeRow, readEventPortRow, readEventTargetRow, readEventTargetPortRow,
  readEventEdgeRow, readEventSlot, readEventValue, readEventBead,
  readEventBeadSteps, readEventX, readEventY, readEventZ,
  readEventLabel, readEventDebug, readEventTextOff, readEventTextLen,
} from "../../../../Buffer/buffer-layout";

export type Line = Record<string, unknown>;

const EVENT_TEXT_DECODER = new TextDecoder();

export function decodeEventLine(ev: DataView, eventTextView: DataView, i: number): Line | null {
  const kindId = readEventKind(ev, i);
  const kind = TRACE_EVENT_KINDS[kindId];
  if (kind === undefined) return null;
  const nodeRow = readEventNodeRow(ev, i);
  const portRow = readEventPortRow(ev, i);
  const targetRow = readEventTargetRow(ev, i);
  const targetPortRow = readEventTargetPortRow(ev, i);
  const edgeRow = readEventEdgeRow(ev, i);
  const value = readEventValue(ev, i);
  const bead = readEventBead(ev, i);
  const node = nodeRow >= 0 ? nodeLabel(nodeRow) : "";

  const port = "";

  if (kind === "breadcrumb") {

    const labelId = readEventLabel(ev, i);
    const label = BREADCRUMB_LABELS[labelId] ?? String(labelId);
    const textOff = readEventTextOff(ev, i);
    const textLen = readEventTextLen(ev, i);
    const text = textLen > 0 && eventTextView.byteLength >= textOff + textLen
      ? EVENT_TEXT_DECODER.decode(new Uint8Array(eventTextView.buffer, eventTextView.byteOffset + textOff, textLen))
      : "";
    const t = targetRow >= 0 ? nodeLabel(targetRow) : "";
    const line: Line = {
      kind, label, debug: readEventDebug(ev, i) === 1,
      node, port, value,
      x: readEventX(ev, i), y: readEventY(ev, i), z: readEventZ(ev, i),
      nodeRow, portRow, targetRow, targetPortRow, edgeRow, slot: readEventSlot(ev, i),
    };
    if (t) line.target = t;
    if (text) line.text = text;
    return line;
  }

  switch (kind) {
    case "recv":
      return { kind, node, port, value };
    case "fire":
      return { kind, node };
    case "send": {
      const beadSteps = readEventBeadSteps(ev, i);
      if (beadSteps !== 0) {
        const l: Line = { kind, node, port, value, beadSteps };
        const t = targetRow >= 0 ? nodeLabel(targetRow) : "";
        if (t) l.target = t;

        return l;
      }
      return { kind, node, port, value };
    }
    case "arrive": {
      const l: Line = { kind, node, port, value };
      if (bead !== 0) l.bead = bead;
      return l;
    }
    default:
      return { kind, node, port, value };
  }
}
