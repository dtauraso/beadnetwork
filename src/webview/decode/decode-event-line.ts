import { BREADCRUMB_LABELS } from "../../schema/buffer-layout/trace-kinds";
import { nodeLabel } from "../../Node/buffer-decode-node";
import {
  readRecvNodeRow, readRecvValue,
  readFireNodeRow,
  readSendNodeRow, readSendTargetRow, readSendValue, readSendBeadSteps,
  readArriveNodeRow, readArriveValue, readArriveBead,
  readBreadcrumbNodeRow, readBreadcrumbPortRow, readBreadcrumbTargetRow,
  readBreadcrumbTargetPortRow, readBreadcrumbEdgeRow, readBreadcrumbSlot, readBreadcrumbValue,
  readBreadcrumbX, readBreadcrumbY, readBreadcrumbZ,
  readBreadcrumbLabel, readBreadcrumbDebug, readBreadcrumbTextOff, readBreadcrumbTextLen,
} from "../../schema/buffer-layout/buffer-layout";
import type { DecodedEvents } from "./buffer-decode-shared";

export type Line = Record<string, unknown>;

const EVENT_TEXT_DECODER = new TextDecoder();

const nameOf = (row: number): string => (row >= 0 ? nodeLabel(row) : "");

function recvLine(ev: DataView, i: number): Line {
  return { kind: "recv", node: nameOf(readRecvNodeRow(ev, i)), port: "", value: readRecvValue(ev, i) };
}

function fireLine(ev: DataView, i: number): Line {
  return { kind: "fire", node: nameOf(readFireNodeRow(ev, i)) };
}

function sendLine(ev: DataView, i: number): Line {
  const line: Line = {
    kind: "send", node: nameOf(readSendNodeRow(ev, i)), port: "", value: readSendValue(ev, i),
  };
  const beadSteps = readSendBeadSteps(ev, i);
  if (beadSteps !== 0) {
    line.beadSteps = beadSteps;
    const target = nameOf(readSendTargetRow(ev, i));
    if (target) line.target = target;
  }
  return line;
}

function arriveLine(ev: DataView, i: number): Line {
  const line: Line = {
    kind: "arrive", node: nameOf(readArriveNodeRow(ev, i)), port: "", value: readArriveValue(ev, i),
  };
  const bead = readArriveBead(ev, i);
  if (bead !== 0) line.bead = bead;
  return line;
}

function breadcrumbLine(ev: DataView, textView: DataView, i: number): Line {
  const labelId = readBreadcrumbLabel(ev, i);
  const textOff = readBreadcrumbTextOff(ev, i);
  const textLen = readBreadcrumbTextLen(ev, i);
  const text = textLen > 0 && textView.byteLength >= textOff + textLen
    ? EVENT_TEXT_DECODER.decode(new Uint8Array(textView.buffer, textView.byteOffset + textOff, textLen))
    : "";
  const targetRow = readBreadcrumbTargetRow(ev, i);
  const line: Line = {
    kind: "breadcrumb",
    label: BREADCRUMB_LABELS[labelId] ?? String(labelId),
    debug: readBreadcrumbDebug(ev, i) === 1,
    node: nameOf(readBreadcrumbNodeRow(ev, i)), port: "", value: readBreadcrumbValue(ev, i),
    x: readBreadcrumbX(ev, i), y: readBreadcrumbY(ev, i), z: readBreadcrumbZ(ev, i),
    nodeRow: readBreadcrumbNodeRow(ev, i), portRow: readBreadcrumbPortRow(ev, i),
    targetRow, targetPortRow: readBreadcrumbTargetPortRow(ev, i),
    edgeRow: readBreadcrumbEdgeRow(ev, i), slot: readBreadcrumbSlot(ev, i),
  };
  const target = nameOf(targetRow);
  if (target) line.target = target;
  if (text) line.text = text;
  return line;
}

export function decodeEventLines(ev: DecodedEvents, breadcrumbsOnly: boolean): Line[] {
  const out: Line[] = [];
  if (!breadcrumbsOnly) {
    for (let i = 0; i < ev.recv.count; i++) out.push(recvLine(ev.recv.view, i));
    for (let i = 0; i < ev.fire.count; i++) out.push(fireLine(ev.fire.view, i));
    for (let i = 0; i < ev.send.count; i++) out.push(sendLine(ev.send.view, i));
    for (let i = 0; i < ev.arrive.count; i++) out.push(arriveLine(ev.arrive.view, i));
  }
  for (let i = 0; i < ev.breadcrumb.count; i++) {
    out.push(breadcrumbLine(ev.breadcrumb.view, ev.breadcrumbTextView, i));
  }
  return out;
}
