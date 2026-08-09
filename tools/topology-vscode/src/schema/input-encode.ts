// input-encode.ts — BINARY encoders for the editor→Go input stream, plus the transport
// frame. The webview builds a binary RECORD per message here; the extension host writes
// each record FRAMED as [len:u32-LE][record] to Go's stdin. Go decodes it in
// nodes/Wiring/input_codec.go into the SAME stdinMsg the dispatch loop consumes.
//
// Numbers are little-endian (matching the content buffer's little-endian encoding). Enum
// discriminators (event kind, hit kind, update entity kind, update attr, overlay flag) are
// u8 indices into the shared orderings from ./input-layout-gen.ts.
//
// NOTE: there is no encodeSave here. IN_KIND_SAVE stays defined (Go reads it and it is in
// the INPUT_LAYOUT_FINGERPRINT), but no live TS sender builds that record: `save` has no
// UI affordance today. IN_KIND_EDIT_CREATE/IN_KIND_EDIT_DELETE were removed end-to-end (no
// live TS sender ever emitted them, and their only trigger — a port-drop gesture —
// unconditionally tore down a live wire's in-flight beads via PacedWire.Restore()); their
// kind bytes (20, 21) are left as gaps.

import { ByteWriter, enumIndex } from "./byte-writer";
import {
  IN_KIND_RAW_INPUT,
  IN_KIND_EDIT_UPDATE,
  IN_EVENT_KINDS,
  IN_HIT_KINDS,
  IN_UPDATE_KINDS,
} from "./input-layout-gen";
import {
  IN_OVERLAY_ATTR_TOGGLE,
  IN_CLOCK_ATTR_SPEED,
  IN_DISTANCE_GROUP_ATTR_LENGTH,
  IN_SCENE_ATTR_SELECTED,
  IN_TILT_VECTOR_ATTR_THETA,
  IN_TILT_VECTOR_ATTR_RESET,
  IN_TILT_VECTOR_ATTR_START,
  IN_SCENE_ATTR_LATTICE_POINTS,
  IN_SCENE_ATTR_CREATE,
  IN_SCENE_ATTR_DELETE,
} from "./input-attrs";
import type { RawInputEvent, OverlayFlag } from "../messages";
import { OVERLAY_FLAG_ORDER } from "../messages";

/** Build an overlays TOGGLE record: [22][entityKind=overlays][attr=toggle][u8 flagId].
 *  flagId is the index of `flag` in OVERLAY_FLAG_ORDER — no flag name crosses the wire. */
export function encodeOverlaysToggle(flag: OverlayFlag): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "overlays"));
  w.u8(IN_OVERLAY_ATTR_TOGGLE);
  w.u8(enumIndex(OVERLAY_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}

/** Build a clock SPEED record: [22][entityKind=clock][attr=speed][u8 quarterUnits].
 *  speed is one of the SpeedSlider's six table values (0, 0.25, 0.5, 0.75, 1, 2) — Go owns
 *  the clock; this just signals the multiplier. msg.Num on the Go side is an int, so a
 *  fractional multiplier is sent in QUARTER-UNITS (an integer 0..8: speed*4) rather than
 *  truncated; stdin_dispatch.go's clockAttrHandlers divides back by 4. */
export function encodeClockSpeed(speed: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "clock"));
  w.u8(IN_CLOCK_ATTR_SPEED);
  w.u8(Math.round(speed * 4));
  return w.toArrayBuffer();
}

/** Build a distanceGroup LENGTH record: [22][entityKind=distanceGroup][attr=length]
 *  [u8 groupIndex][u8 dirUp]. groupIndex is the group's WIRE INDEX (0/1/2, into Go's
 *  distanceGroupOrder: time/input/gate — no group name crosses the wire); dirUp is
 *  1 for the up arrow (Go sets target length = currentMax*1.1), 0 for down (÷1.1). Go
 *  owns the group definitions and the ×1.1 math (nodes/Wiring/distance_groups.go); this
 *  just signals which group + which direction. */
export function encodeDistanceGroupAdjust(groupIndex: number, dir: "up" | "down"): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "distanceGroup"));
  w.u8(IN_DISTANCE_GROUP_ATTR_LENGTH);
  w.u8(groupIndex);
  w.u8(dir === "up" ? 1 : 0);
  return w.toArrayBuffer();
}

/** Build a scene SELECTED record: [22][entityKind=scene][attr=selected][u8 tabIndex].
 *  tabIndex indexes Go's OWN scene tab strip (nodes/Wiring/scene_tabs.go's SceneTabs), the
 *  same list whose labels arrive on the VIEW frame — no scene name or directory crosses the
 *  wire. Go owns what the tabs are, which one is selected, and how the switch happens. */
export function encodeSceneSelected(tabIndex: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_SELECTED);
  w.u8(tabIndex);
  return w.toArrayBuffer();
}

/** Build a scene CREATE record: [22][entityKind=scene][attr=create][u8 kindId][f32 ndcX]
 *  [f32 ndcY]. kindId is the kind's NODE_DEFS id — the same numeric kind identity the Node
 *  block's KindId column carries, so no kind NAME crosses this wire.
 *
 *  SCREEN coordinates, not world. Turning a drop into a place in the scene needs the camera,
 *  and the camera is Go's — it unprojects this with the same ray every node drag already
 *  uses. TS forwards where the pointer was, exactly as raw-input does, and computes no
 *  geometry. Which node the new one connects to is not here either: Go picks the nearest
 *  from its own node positions.
 *
 *  Go persists the new node and ENDS THE RUN; the host's looping runner respawns and the new
 *  tree loads. That is not this function's concern, but it is why a create looks like nothing
 *  happened for the length of a restart. */
export function encodeSceneCreate(kindId: number, ndcX: number, ndcY: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_CREATE);
  w.u8(kindId);
  w.f32(ndcX);
  w.f32(ndcY);
  return w.toArrayBuffer();
}

/** Build a scene DELETE record: [22][entityKind=scene][attr=delete][u8 nodeRow]. The target
 *  is a buffer ROW, never an id or a name — the same no-sidecar rule every addressed edit
 *  follows. Go removes the node and every edge touching it, then ends the run. */
export function encodeSceneDelete(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_DELETE);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

/** Build a scene LATTICE-POINTS record: [22][entityKind=scene][attr=latticePoints]
 *  [u8 points]. points is the pair lattice's new point count (4..64, a multiple of 4 —
 *  Go rejects anything else, nodes/Wiring/stdin_reader.go's applyUpdateScene); Go owns the
 *  valid range and the delivery to every pair node, this just signals the requested count. */
export function encodeSceneLatticePoints(points: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_LATTICE_POINTS);
  w.u8(points);
  return w.toArrayBuffer();
}

/** Build a tiltVector THETA/PHI record: [22][entityKind=tiltVector][attr=theta|phi]
 *  [u8 nodeRow][u8 dirUp]. nodeRow is the target node's buffer ROW (never its id/name —
 *  no sidecar on this wire); dirUp is 1 for the up arrow (+1 CurveParamTiltVectorAngleStep
 *  index), 0 for down (-1). Go owns the step constant and the index math
 *  (nodes/Wiring's CurveParamTiltVectorAngleStep, node_mover.go); this just signals which
 *  node, which axis, which direction. */
export function encodeTiltVectorAdjust(nodeRow: number, dir: "up" | "down"): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_THETA);
  w.u8(nodeRow);
  w.u8(dir === "up" ? 1 : 0);
  return w.toArrayBuffer();
}

/** Build a tiltVector RESET record: [22][entityKind=tiltVector][attr=reset][u8 nodeRow].
 *  nodeRow is the target node's buffer ROW (never its id/name — no sidecar on this wire).
 *  No direction byte: unlike an adjust, a reset always returns BOTH indices to 0 — the
 *  RESET button (TiltResetButton.tsx) sends one of these per row it shows, and places no
 *  bead (nodes/PairNode/node.go's applyTiltEdit, run unmodified by both nodes of a pair — a
 *  stop-and-return, not "the kick"). */
export function encodeTiltVectorReset(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_RESET);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

/** Build a tiltVector START record: [22][entityKind=tiltVector][attr=start][u8 nodeRow].
 *  nodeRow is the target node's buffer ROW (never its id/name — no sidecar on this wire).
 *  No direction byte: Start never touches an index, it only opens the vector exchange from
 *  whatever angles are currently set — sends the node's own outgoing vector and places a
 *  bead ("the kick"), exactly what an adjust click used to do as a side effect
 *  (nodes/PairNode/node.go's applyTiltEdit, run unmodified by both nodes of a pair —
 *  task/pair-node-owns-itself split). The START TILT button (TiltVectorButtons.tsx) sends
 *  one of these per row it shows, same fan-out as reset. */
export function encodeTiltVectorStart(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_START);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

/** Build a raw-input record: fully-numeric fixed fields + enum bytes (no JSON). */
export function encodeRawInput(ev: RawInputEvent): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_RAW_INPUT);
  w.u8(enumIndex(IN_EVENT_KINDS, ev.kind));
  w.f64(ev.x);
  w.f64(ev.y);
  w.f64(ev.rectLeft);
  w.f64(ev.rectTop);
  w.f64(ev.rectWidth);
  w.f64(ev.rectHeight);
  w.i32(ev.button);
  w.bool(ev.ctrl);
  w.bool(ev.shift);
  w.bool(ev.alt);
  w.bool(ev.meta);
  w.f64(ev.deltaX);
  w.f64(ev.deltaY);
  w.f64(ev.fov);
  w.u8(enumIndex(IN_HIT_KINDS, ev.hit.kind));
  w.bool(ev.hit.isInput);
  w.i32(ev.hit.nodeRow);
  w.i32(ev.hit.portRow);
  w.i32(ev.hit.edgeRow);
  return w.toArrayBuffer();
}

/** Wrap a record body with the [len:u32-LE] transport frame (used by the host writer). */
export function frameRecord(record: ArrayBuffer): Uint8Array {
  const rec = new Uint8Array(record);
  const out = new Uint8Array(4 + rec.length);
  new DataView(out.buffer).setUint32(0, rec.length, true);
  out.set(rec, 4);
  return out;
}
