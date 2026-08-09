// input-decode.ts — decode one editor→Go input record body. Used by unit tests + the
// encode/decode round-trip; production decode of the SAME wire format is
// nodes/Wiring/input_codec.go (Go decodes for real, TS here never reads its own traffic
// back off the bridge).

import { ByteReader } from "./byte-reader";
import { IN_KIND_SAVE, IN_KIND_RAW_INPUT, IN_KIND_EDIT_UPDATE, IN_EVENT_KINDS, IN_HIT_KINDS, IN_UPDATE_KINDS } from "./input-layout-gen";
import { IN_OVERLAY_ATTR_TOGGLE, IN_CLOCK_ATTR_SPEED, IN_DISTANCE_GROUP_ATTR_LENGTH } from "./input-attrs";
import type { RawInputEvent, OverlayFlag } from "../messages";
import { OVERLAY_FLAG_ORDER } from "../messages";

export type DecodedInput =
  | { kind: "save" }
  | { kind: "raw-input"; event: RawInputEvent }
  | { kind: "edit-update"; entity: "overlays"; attr: "toggle"; flag: OverlayFlag }
  | { kind: "edit-update"; entity: "clock"; attr: "speed"; value: number }
  | { kind: "edit-update"; entity: "distanceGroup"; attr: "length"; group: number; dir: "up" | "down" };

/** Decode one record body (with kind byte, without the [len] frame). */
export function decodeInputRecord(record: ArrayBuffer): DecodedInput | undefined {
  const bytes = new Uint8Array(record);
  if (bytes.length === 0) return undefined;
  const r = new ByteReader(bytes);
  switch (bytes[0]) {
    case IN_KIND_SAVE:
      return { kind: "save" };
    case IN_KIND_RAW_INPUT: {
      const event: RawInputEvent = {
        kind: IN_EVENT_KINDS[r.u8()] ?? "pointermove",
        x: r.f64(),
        y: r.f64(),
        rectLeft: r.f64(),
        rectTop: r.f64(),
        rectWidth: r.f64(),
        rectHeight: r.f64(),
        button: r.i32(),
        ctrl: r.bool(),
        shift: r.bool(),
        alt: r.bool(),
        meta: r.bool(),
        deltaX: r.f64(),
        deltaY: r.f64(),
        fov: r.f64(),
        hit: {
          kind: IN_HIT_KINDS[r.u8()] ?? "empty",
          isInput: r.bool(),
          nodeRow: r.i32(),
          portRow: r.i32(),
          edgeRow: r.i32(),
        },
      };
      return { kind: "raw-input", event };
    }
    case IN_KIND_EDIT_UPDATE: {
      // entityKind byte selects the entity; attr byte + numeric payload follow.
      const entityKind = IN_UPDATE_KINDS[r.u8()];
      if (entityKind === "overlays") {
        const attr = r.u8();
        if (attr === IN_OVERLAY_ATTR_TOGGLE) {
          const flag = OVERLAY_FLAG_ORDER[r.u8()];
          if (!flag) return undefined;
          return { kind: "edit-update", entity: "overlays", attr: "toggle", flag };
        }
        return undefined;
      }
      if (entityKind === "clock") {
        const attr = r.u8();
        if (attr === IN_CLOCK_ATTR_SPEED) {
          // Wire value is QUARTER-UNITS (see input-encode.ts's encodeClockSpeed); divide
          // back to the real multiplier so decodeInputRecord's `value` matches what the
          // caller passed in.
          const value = r.u8() / 4;
          return { kind: "edit-update", entity: "clock", attr: "speed", value };
        }
        return undefined;
      }
      if (entityKind === "distanceGroup") {
        const attr = r.u8();
        if (attr === IN_DISTANCE_GROUP_ATTR_LENGTH) {
          const group = r.u8();
          const dirUp = r.u8();
          return { kind: "edit-update", entity: "distanceGroup", attr: "length", group, dir: dirUp ? "up" : "down" };
        }
        return undefined;
      }
      return undefined;
    }
  }
  return undefined;
}
