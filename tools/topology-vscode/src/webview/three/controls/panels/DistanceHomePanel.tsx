import React, { useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeDistanceGroupAdjust } from "../../../../schema/input-encode";
import { useDistanceGroupLens } from "../flags/overlay-flags-distance-groups";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  PILL_ANCHOR_STYLE,
  inFlowPopoverStyle,
} from "../pills/overlay-chrome";
import { StepperRow } from "../pills/pill-rows";

// DistanceHomePanel — the RING tab's node-pair distance control: 3 named groups
// (time / input / select — Go's distanceGroupOrder, whose third group is still keyed "gate"
// internally; only the visible label reads "select", matching the renamed Select* gate
// structs — nodes/Wiring/distance_groups.go), each showing its CURRENT max pair length
// (read-only reflect of the Overlay block's GroupLenTime/GroupLenInput/GroupLenGate columns)
// with ▲/▼ that change it.
//
// It is a PILL + POPOVER, the same chrome as OverlaysControl and TiltVectorAnglePanel
// (overlay-chrome.ts for the pill/popover, pill-rows.tsx for the rows), rather than the
// always-open list panel it used to be. That is not just a look: the three panels had three
// answers to the same questions — where the arrows sit, whether a value's width moves them,
// whether the panel is open all the time — and only one of those answers can be the right
// one. This panel now inherits them instead of restating them, and panel-styles.ts, the
// module that existed to share the OLD look between this panel and the angles panel, is gone
// with its last caller.
//
// Clicking an arrow fire-and-forgets an edit-update(distanceGroup, length) record naming the
// group's WIRE INDEX and the direction; Go owns the group definitions AND the ×1.1/÷1.1 math
// (this component sends no length value, only which group + which direction) and repositions
// every pair's target node via RootMove, which rebroadcasts geometry so the node moves and
// its edge redraws. No local domain state — only whether the popover is open, which is view-
// local and deliberately not a Go flag, the same as the other two controls' open state.
const GROUPS: { index: number; label: string }[] = [
  { index: 0, label: "time" },
  { index: 1, label: "input" },
  { index: 2, label: "select" },
];

// The value reservation (StepperRow's `widest`): lengths are Go's own ×1.1 growth rounded to
// an integer, with no bound to read, so this reserves FOUR digits and widens only past them.
// Sized by digit count rather than by the current value, so stepping 99 → 109 does not move
// the arrows; a scene that genuinely runs into five digits gets one width change, once.
function widestLength(values: number[]): string {
  const digits = Math.max(4, ...values.map((v) => String(Math.round(v)).length));
  return "8".repeat(digits);
}

export function DistanceHomePanel() {
  const lens = useDistanceGroupLens();
  const [open, setOpen] = useState(false);

  // Data-driven, not scene-branching: a scene whose nodes don't resolve any of the 3 groups'
  // pairs (e.g. the pair scene, whose only nodes are outside every group in
  // nodes/Wiring/distance_groups.go's distanceGroups table) streams all three GroupLen*
  // columns as 0 forever — distanceGroupMax's `any` never turns true, so
  // ApplyDistanceGroupTarget's math never has anything to act on either. That is exactly the
  // "no groups" condition, so render nothing rather than a pill of dead buttons. A real
  // ring-topology group briefly reading 0 before its first resolvable frame lands is the same
  // signal and self-corrects the next frame.
  if (!lens || (lens.time === 0 && lens.input === 0 && lens.gate === 0)) return null;

  const valueFor = (index: number): number => {
    if (index === 0) return lens.time;
    if (index === 1) return lens.input;
    return lens.gate;
  };

  const adjust = (index: number, dir: "up" | "down") => {
    postGoRecord(encodeDistanceGroupAdjust(index, dir));
  };

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  const widest = widestLength(GROUPS.map(({ index }) => valueFor(index)));

  return (
    // Pill and popover are siblings in the shared-width anchor, never nested: the pill clips
    // its own rounded corners with `overflow: hidden`, which would clip a popover inside it
    // out of existence.
    <div style={PILL_ANCHOR_STYLE}>
      <div style={pillContainerStyle(false)}>
        {/* No master toggle — there is nothing to turn on or off, only lengths to read and
            adjust — so the WHOLE pill opens the popover, as on the angles pill. The label
            takes the pill's slack so the caret stays at the far end. */}
        <div
          onClick={onToggle}
          title={open ? "Close distances" : "Open distances"}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Distances
        </div>
        <div
          onClick={onToggle}
          title={open ? "Close distances" : "Open distances"}
          style={pillCaretStyle}
        >
          {open ? "▲" : "▼"}
        </div>
      </div>

      {open && (
        <div style={inFlowPopoverStyle()}>
          {GROUPS.map(({ index, label }) => (
            <StepperRow
              key={label}
              name={label}
              shown={String(Math.round(valueFor(index)))}
              widest={widest}
              upLabel={`${label} distance up`}
              downLabel={`${label} distance down`}
              onUp={() => adjust(index, "up")}
              onDown={() => adjust(index, "down")}
            />
          ))}
        </div>
      )}
    </div>
  );
}
