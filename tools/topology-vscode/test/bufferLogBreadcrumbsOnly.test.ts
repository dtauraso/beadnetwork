// Covers buffer-log.ts's `breadcrumbsOnly` filter: when true, decodeStreamFrameEvents
// (and decodeBufferLog, same decodeEventsFromView-family gate) must drop every decoded
// line whose kind !== "breadcrumb" before serialization, so the DEBUG BREADCRUMB channel
// (CLAUDE.md; tools/probe-merge.sh --debug) keeps working while wirefold.probe.trace is
// off — the exact regression this parameter was added to fix.
//
// Frame source (discovery mandate, in order tried):
//  1. No existing binary frame fixture/golden file covers the EVENTS section (grepped
//     test/fixtures and Buffer/ — test/fixtures/trace-events.jsonl is DECODED JSONL, not a
//     binary frame, and cannot drive decodeStreamFrameEvents; see its own test's comment).
//  2. No Go-side exported packer/golden-frame generator exists for the EVENTS section
//     alone (Buffer.BuildEventsSection is unexported and there is no `go run`/test target
//     that writes just an events section to disk).
//  3. So: hand-construct the binary EVENTS section using ONLY generated layout constants —
//     EVENT_COL_*/EVENT_STRIDE from schema/buffer-layout.ts (generated from Buffer/layout.go)
//     — and decode it through the real decodeTrailingEvents (buffer-decode.ts), the same
//     function every real per-owner stream frame is decoded through. No hardcoded offsets.

import { describe, it, expect } from "vitest";
import { decodeStreamFrameEvents } from "../src/buffer-log";
import { decodeTrailingEvents } from "../src/webview/three/buffer-decode";
import {
  EVENT_STRIDE,
  EVENT_COL_KIND, EVENT_COL_NODE_ROW, EVENT_COL_PORT_ROW, EVENT_COL_TARGET_ROW,
  EVENT_COL_TARGET_PORT_ROW, EVENT_COL_EDGE_ROW, EVENT_COL_SLOT, EVENT_COL_VALUE,
  EVENT_COL_BEAD, EVENT_COL_X, EVENT_COL_Y, EVENT_COL_Z, EVENT_COL_F,
  EVENT_COL_LABEL, EVENT_COL_DEBUG, EVENT_COL_TEXT_OFF, EVENT_COL_TEXT_LEN,
} from "../src/schema/buffer-layout";
import { TRACE_EVENT_KINDS, BREADCRUMB_LABELS } from "../src/schema/trace-kinds";

const EDGE_BEAD_KIND_ID = TRACE_EVENT_KINDS.indexOf("edge-bead");
const BREADCRUMB_KIND_ID = TRACE_EVENT_KINDS.indexOf("breadcrumb");
const BREADCRUMB_LABEL_ID = BREADCRUMB_LABELS.indexOf("dwell_start");

/** Builds one raw frame buffer holding an EVENTS section
 *  ([count:u32] + count*EVENT_STRIDE rows + trailing text bytes) exactly as
 *  Buffer.BuildEventsSection lays it out, using only generated column offsets — one
 *  edge-bead row (a representative non-breadcrumb kind) and one breadcrumb row (with a
 *  real label/debug flag/text payload) so both directions of the filter are exercised
 *  from a single decoded frame. */
function makeEventsFrame(): ArrayBuffer {
  const enc = new TextEncoder();
  const text = enc.encode("dwell-node-a");
  const rows = 2;
  const bytes = rows * EVENT_STRIDE;
  const total = 4 + bytes + text.length;
  const buf = new ArrayBuffer(total);
  const dv = new DataView(buf);
  dv.setUint32(0, rows, true);

  // Row 0: edge-bead — a representative high-volume trace kind, no breadcrumb fields.
  const r0 = 4 + 0 * EVENT_STRIDE;
  dv.setUint8(r0 + EVENT_COL_KIND, EDGE_BEAD_KIND_ID);
  dv.setInt32(r0 + EVENT_COL_NODE_ROW, 0, true);
  dv.setInt32(r0 + EVENT_COL_PORT_ROW, 0, true);
  dv.setInt32(r0 + EVENT_COL_TARGET_ROW, -1, true);
  dv.setInt32(r0 + EVENT_COL_TARGET_PORT_ROW, -1, true);
  dv.setInt32(r0 + EVENT_COL_EDGE_ROW, 0, true);
  dv.setInt32(r0 + EVENT_COL_SLOT, 0, true);
  dv.setInt32(r0 + EVENT_COL_VALUE, 7, true);
  dv.setUint32(r0 + EVENT_COL_BEAD, 3, true);
  dv.setFloat32(r0 + EVENT_COL_X, 1.5, true);
  dv.setFloat32(r0 + EVENT_COL_Y, 2.5, true);
  dv.setFloat32(r0 + EVENT_COL_Z, 3.5, true);
  dv.setFloat32(r0 + EVENT_COL_F, 0.25, true);

  // Row 1: breadcrumb — DEBUG BREADCRUMB channel row, label + debug flag + text payload.
  const r1 = 4 + 1 * EVENT_STRIDE;
  const textOff = 0; // relative to eventTextView's own start (the text section itself).
  dv.setUint8(r1 + EVENT_COL_KIND, BREADCRUMB_KIND_ID);
  dv.setInt32(r1 + EVENT_COL_NODE_ROW, 0, true);
  dv.setInt32(r1 + EVENT_COL_PORT_ROW, -1, true);
  dv.setInt32(r1 + EVENT_COL_TARGET_ROW, -1, true);
  dv.setInt32(r1 + EVENT_COL_TARGET_PORT_ROW, -1, true);
  dv.setInt32(r1 + EVENT_COL_EDGE_ROW, -1, true);
  dv.setInt32(r1 + EVENT_COL_SLOT, 0, true);
  dv.setInt32(r1 + EVENT_COL_VALUE, 0, true);
  dv.setUint32(r1 + EVENT_COL_BEAD, 0, true);
  dv.setFloat32(r1 + EVENT_COL_X, 0, true);
  dv.setFloat32(r1 + EVENT_COL_Y, 0, true);
  dv.setFloat32(r1 + EVENT_COL_Z, 0, true);
  dv.setFloat32(r1 + EVENT_COL_F, 0, true);
  dv.setUint8(r1 + EVENT_COL_LABEL, BREADCRUMB_LABEL_ID);
  dv.setUint8(r1 + EVENT_COL_DEBUG, 1);
  dv.setUint32(r1 + EVENT_COL_TEXT_OFF, textOff, true);
  dv.setUint32(r1 + EVENT_COL_TEXT_LEN, text.length, true);

  new Uint8Array(buf, 4 + bytes, text.length).set(text);
  return buf;
}

describe("decodeStreamFrameEvents breadcrumbsOnly filter", () => {
  it("breadcrumbsOnly=true keeps only the breadcrumb row", () => {
    const buf = makeEventsFrame();
    const { count, view, textView } = decodeTrailingEvents(buf, 0);
    expect(count).toBe(2); // sanity: both rows really decoded before filtering.

    const out = decodeStreamFrameEvents(count, view, textView, null, null, true);
    const lines = out.trim().split("\n").filter(Boolean).map((l) => JSON.parse(l));

    expect(lines).toHaveLength(1);
    expect(lines[0].kind).toBe("breadcrumb");
    // The surviving breadcrumb row must carry its real decoded fields, not a stripped line.
    expect(lines[0].label).toBe("dwell_start");
    expect(lines[0].debug).toBe(true);
    expect(lines[0].text).toBe("dwell-node-a");
    expect(lines[0].nodeRow).toBe(0);
    expect(lines[0].slot).toBe(0);
  });

  it("breadcrumbsOnly=false (default) keeps every row, breadcrumb included", () => {
    const buf = makeEventsFrame();
    const { count, view, textView } = decodeTrailingEvents(buf, 0);

    const out = decodeStreamFrameEvents(count, view, textView, null, null, false);
    const lines = out.trim().split("\n").filter(Boolean).map((l) => JSON.parse(l));

    expect(lines).toHaveLength(2);
    expect(lines.map((l) => l.kind).sort()).toEqual(["breadcrumb", "edge-bead"]);
    const bc = lines.find((l) => l.kind === "breadcrumb");
    expect(bc.label).toBe("dwell_start");
    expect(bc.debug).toBe(true);
    expect(bc.text).toBe("dwell-node-a");
  });
});
